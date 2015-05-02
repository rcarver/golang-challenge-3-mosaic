package mosaic

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
)

type ImageCache interface {
	Put(key string, m image.Image) error
	Get(key string) (image.Image, error)
	Has(key string) bool
}

type FileImageCache struct {
	Dir string
}

func (c FileImageCache) Put(key string, m image.Image) error {
	fo, err := os.Create(c.path(key))
	if err != nil {
		return err
	}
	defer fo.Close()
	if err := jpeg.Encode(fo, m, nil); err != nil {
		return err
	}
	return nil
}

func (c FileImageCache) Get(key string) (image.Image, error) {
	fi, err := os.Open(c.path(key))
	defer fi.Close()
	m, err := jpeg.Decode(fi)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c FileImageCache) Has(key string) bool {
	if _, err := os.Stat(c.path(key)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c FileImageCache) path(key string) string {
	k := sha1.Sum([]byte(key))
	h := hex.EncodeToString(k[:])
	return fmt.Sprintf("%s/%s.jpg", c.Dir, h)
}

type ImageInventory struct {
	ImageCache
	Tags map[string][]string
}

func NewImageInventory(cache ImageCache) *ImageInventory {
	return &ImageInventory{
		ImageCache: cache,
		Tags:       make(map[string][]string),
	}
}

func (i ImageInventory) PopulatePalette(tag string, palette *ImagePalette) error {
	keys := i.Tags[tag]
	for _, key := range keys {
		m, err := i.Get(key)
		if err != nil {
			fmt.Println("Error reading from cache %s", err)
			continue
		}
		palette.Add(m)
	}
	return nil
}

func (ii *ImageInventory) Fetch(api *instagram.Client, tag string, max int) error {
	res, err := api.Tagged(tag, "")
	if err != nil {
		return err
	}
	for {
		for _, m := range res.Media {
			if len(ii.Tags[tag]) >= max {
				return nil
			}
			key, err := ii.cacheImage(m)
			if err != nil {
				return err
			}
			ii.Tags[tag] = append(ii.Tags[tag], key)
		}
		fmt.Printf("got %d tiles\n", len(ii.Tags[tag]))
		res, err = api.Tagged(tag, res.MaxTagID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ii *ImageInventory) cacheImage(media instagram.Media) (string, error) {
	res := media.ThumbnailImage()
	if ii.Has(res.URL) {
		//fmt.Printf("Has %s\n", res.URL)
		return res.URL, nil
	}
	img, err := res.Image()
	if err != nil {
		return "", nil
	}
	//fmt.Printf("Get %s\n", res.URL)
	if err := ii.Put(res.URL, img); err != nil {
		return "", err
	}
	return res.URL, nil
}
