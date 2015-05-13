package mosaic

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
)

// ImageInventory fetches and uses images to drive mosaic creation.
type ImageInventory struct {
	cache ImageCache
}

// NewImageInventory creates an inventory using the given cache.
func NewImageInventory(cache ImageCache) *ImageInventory {
	return &ImageInventory{
		cache: cache,
	}
}

// PopulatePalette pulls images from the inventory and adds them to a palette.
func (ii *ImageInventory) PopulatePalette(palette *ImagePalette) error {
	keys, err := ii.cache.Keys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		m, err := ii.cache.Get(key)
		if err != nil {
			//log.Printf("Error reading from cache: %s, key:%#v\n", err, key)
			continue
		}
		palette.Add(m)
	}
	return nil
}

// Fetch pulls new images from the api and adds them to the inventory.
func (ii *ImageInventory) Fetch(api *instagram.Client, tag string, max int) error {
	res, err := api.Tagged(tag, "")
	if err != nil {
		return err
	}
	for {
		for _, m := range res.Media {
			if ii.cache.Size() >= max {
				return nil
			}
			err := ii.cacheImage(m)
			if err != nil {
				return err
			}
		}
		log.Printf("Fetch(%s) got %d of %d\n", tag, ii.cache.Size(), max)
		res, err = api.Tagged(tag, res.MaxTagID)
		if err != nil {
			return err
		}
	}
	return nil
}

// Size returns the number of images in the inventory.
func (ii *ImageInventory) Size(tag string) int {
	return ii.cache.Size()
}

func (ii *ImageInventory) cacheImage(media instagram.Media) error {
	res := media.ThumbnailImage()
	key := ii.cache.Key(res.URL)
	if ii.cache.Has(key) {
		//log.Printf("Has %s\n", res.URL)
		return nil
	}
	img, err := res.Image()
	if err != nil {
		return nil
	}
	//log.Printf("Get %s\n", res.URL)
	if err := ii.cache.Put(key, img); err != nil {
		return err
	}
	return nil
}

// ImageCacheKey identifies an image in the cache.
type ImageCacheKey string

// ImageCache is the interface for any cache implementation.
type ImageCache interface {
	// Key returns a consistent cache key from string.
	Key(name string) ImageCacheKey
	// Put stores an image in the cache by key.
	Put(ImageCacheKey, image.Image) error
	// Get returns an image in the cache by key
	Get(ImageCacheKey) (image.Image, error)
	// Has returns true if an image exists at key.
	Has(ImageCacheKey) bool
	// Keys returns a list of the stored keys, unordered.
	Keys() ([]ImageCacheKey, error)
	// Size returns the number of images stores.
	Size() int
}

// FileImageCache implements an ImageCache on the filesystem.
type FileImageCache struct {
	Dir string
}

func (c FileImageCache) Key(name string) string {
	k := sha1.Sum([]byte(name))
	return hex.EncodeToString(k[:])
}

func (c FileImageCache) Put(key ImageCacheKey, m image.Image) error {
	fo, err := os.Create(c.keyToPath(key))
	if err != nil {
		return err
	}
	defer fo.Close()
	if err := jpeg.Encode(fo, m, nil); err != nil {
		return err
	}
	return nil
}

func (c FileImageCache) Get(key ImageCacheKey) (image.Image, error) {
	fi, err := os.Open(c.keyToPath(key))
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	m, err := jpeg.Decode(fi)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c FileImageCache) Keys() ([]ImageCacheKey, error) {
	list, err := filepath.Glob(path.Join(c.Dir, "*.jpg"))
	if err != nil {
		return []ImageCacheKey{}, err
	}
	keys := make([]ImageCacheKey, len(list))
	for i, o := range list {
		keys[i] = c.pathToKey(o)
	}
	return keys, nil
}

func (c FileImageCache) Size() int {
	list, err := c.Keys()
	if err == nil {
		return len(list)
	}
	return 0
}

func (c FileImageCache) Has(key ImageCacheKey) bool {
	if _, err := os.Stat(c.keyToPath(key)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c FileImageCache) keyToPath(key ImageCacheKey) string {
	return fmt.Sprintf("%s/%s.jpg", c.Dir, key)
}

func (c FileImageCache) pathToKey(path string) ImageCacheKey {
	return ImageCacheKey(strings.Trim(filepath.Base(path), filepath.Ext(path)))
}
