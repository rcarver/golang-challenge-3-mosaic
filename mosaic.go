package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
)

func main() {
	api := instagram.NewClient()
	images := fetchImagesWithTag(api, "cat", 100)
	fmt.Printf("done with %d tiles\n", len(images))

	palette := &imagePalette{}
	for _, m := range images {
		palette.Add(m)
	}

	src := images[0]
	grid := PixelGrid{20, 20, 10}
	m := grid.MosaicImage(src, 1000, 1000)
	m.Draw(palette)

	out, err := os.Create("./source.png")
	err = png.Encode(out, src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	out, err = os.Create("./mosaic.png")
	err = png.Encode(out, m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func fetchImagesWithTag(api instagram.Client, tag string, count int) []image.Image {
	res, err := api.Tagged(tag, "")

	if err != nil {
		panic(err)
	}

	images := make([]image.Image, 0, count)
	for {
		err := fetchImages(&images, res.Media, count)
		if err != nil {
			panic(err)
		}
		fmt.Printf("got %d tiles\n", len(images))
		if len(images) > count {
			break
		}
		res, err = api.Tagged("cat", res.MaxTagID)
		if err != nil {
			panic(err)
		}
	}
	return images
}

func fetchImages(images *[]image.Image, media []instagram.Media, max int) error {
	for i, m := range media {
		fmt.Printf("> %d Reading tile %s\n", i, m.StandardImage().URL)
		img, err := m.StandardImage().Image()
		if err != nil {
			return err
		}
		fmt.Printf("< %d Reading tile %s\n", i, m.StandardImage().URL)
		*images = append(*images, img)
		if len(*images) > max {
			return nil
		}
	}
	return nil
}
