package main

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

func NewImageInventory(cache ImageCache) *ImageInventory {
	return &ImageInventory{ImageCache: cache}
}

type ImageInventory struct {
	ImageCache
	keys []string
}

func (i ImageInventory) ImagePalette(max int) (ImagePalette, error) {
	palette := &imagePalette{}
	for j, key := range i.keys {
		m, err := i.Get(key)
		if err != nil {
			return palette, err
		}
		if j > max {
			break
		}
		palette.Add(m)
	}
	return palette, nil
}

func (ii *ImageInventory) Fetch(api instagram.Client, tag string, max int) error {
	res, err := api.Tagged(tag, "")
	if err != nil {
		return err
	}

	for {
		for _, m := range res.Media {
			if len(ii.keys) >= max {
				return nil
			}
			err := ii.cacheImage(m)
			if err != nil {
				return err
			}
		}
		fmt.Printf("got %d tiles\n", len(ii.keys))
		res, err = api.Tagged(tag, res.MaxTagID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ii *ImageInventory) cacheImage(media instagram.Media) error {
	res := media.ThumbnailImage()
	if ii.Has(res.URL) {
		fmt.Printf("Has %s\n", res.URL)
		ii.keys = append(ii.keys, res.URL)
		return nil
	}
	img, err := res.Image()
	if err != nil {
		return err
	}
	fmt.Printf("Get %s\n", res.URL)
	ii.keys = append(ii.keys, res.URL)
	if err := ii.Put(res.URL, img); err != nil {
		return err
	}
	return nil
}
