package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
)

var tag = "balloon"
var fetch = 100
var palette = 100

func main() {
	cache := FileImageCache{"./cache"}
	inventory := NewImageInventory(cache)

	api := instagram.NewClient()
	if err := inventory.Fetch(api, tag, fetch); err != nil {
		fmt.Printf("Fetch() %s\n", err)
		os.Exit(1)
	}

	palette, err := inventory.ImagePalette(palette)
	if err != nil {
		fmt.Printf("ImagePalette() %s\n", err)
		os.Exit(1)
	}

	fi, err := os.Open("./fixtures/balloon.jpg")
	if err != nil {
		fmt.Printf("os.Open() %s\n", err)
		os.Exit(1)
	}
	defer fi.Close()
	src, err := decodeImage(fi)
	if err != nil {
		fmt.Printf("decodeImage() %s\n", err)
		os.Exit(1)
	}

	grid := PixelGrid{20, 20, 10}
	m := grid.MosaicImage(src, 3000, 3000)
	m.Draw(palette)

	out, err := os.Create("./output.png")
	if err != nil {
		fmt.Printf("os.Create() %s\n", err)
		os.Exit(1)
	}
	defer out.Close()
	err = png.Encode(out, m)
	if err != nil {
		fmt.Printf("png.Encode() %s\n", err)
		os.Exit(1)
	}
}
