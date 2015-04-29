package main

import (
	"fmt"
	"image/color/palette"
	"image/png"
	"os"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
)

var tag = "balloon"
var inventorySize = 200
var paletteSize = 200
var solidPalette = false

func main() {
	var p ImagePalette

	if solidPalette {
		p = solidImagePalette{palette.WebSafe}
	} else {
		cache := FileImageCache{"./cache"}
		inventory := NewImageInventory(cache)

		api := instagram.NewClient()
		if err := inventory.Fetch(api, tag, inventorySize); err != nil {
			fmt.Printf("Fetch() %s\n", err)
			os.Exit(1)
		}

		ip, err := inventory.ImagePalette(paletteSize)
		if err != nil {
			fmt.Printf("ImagePalette() %s\n", err)
			os.Exit(1)
		}
		p = ip
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

	units := 40
	thumbSize := 150
	grid := PixelGrid{units, units, .5}
	m := grid.MosaicImage(src, thumbSize*units, thumbSize*units)
	m.Draw(p)

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
