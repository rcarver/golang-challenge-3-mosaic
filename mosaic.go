package main

import (
	"fmt"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
)

func main() {
	api := instagram.NewClient()
	tiles := fetchImagesWithTag(api, "cat", 100)
	fmt.Printf("done with %d tiles\n", len(tiles))

}

func fetchImagesWithTag(api instagram.Client, tag string, count int) []*TileImage {
	ti := make([]*TileImage, 0, count)
	res, err := api.Tagged(tag, "")

	if err != nil {
		panic(err)
	}

	for {
		fetchImages(&ti, res.Media)
		if len(ti) > 100 {
			break
		}
		fmt.Printf("got %d tiles\n", len(ti))
		res, err = api.Tagged("cat", res.MaxTagID)
		if err != nil {
			panic(err)
		}
	}
	return ti
}

func fetchImages(tiles *[]*TileImage, media []instagram.Media) error {
	for i, m := range media {
		ti := &TileImage{}
		fmt.Printf("> %d Reading tile %s\n", i, m.StandardImage().URL)
		//_, err := ti.ReadFrom(m.StandardImage())
		//if err != nil {
		//return err
		//}
		fmt.Printf("< %d Reading tile %s\n", i, m.StandardImage().URL)
		*tiles = append(*tiles, ti)
	}
	return nil
}
