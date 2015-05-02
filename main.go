package main

import (
	"fmt"
	"image/color/palette"
	"image/png"
	"os"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
)

var tag = "balloon"
var inventorySize = 400
var solidPalette = false

func main() {
	units := 40
	thumbSize := 150
	var p *mosaic.ImagePalette

	if solidPalette {
		p = mosaic.NewSolidPalette(palette.WebSafe, thumbSize, thumbSize)
	} else {
		cache := mosaic.FileImageCache{"./cache"}
		inv := mosaic.NewImageInventory(cache)

		api := instagram.NewClient()
		if err := inv.Fetch(api, tag, inventorySize); err != nil {
			fmt.Printf("Fetch() %s\n", err)
			os.Exit(1)
		}

		p = mosaic.NewImagePalette(256, thumbSize, thumbSize)
		err := inv.PopulatePalette(tag, p)
		if err != nil {
			fmt.Printf("PopulatePalette() %s\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Palette size %d with %d images\n", p.Size(), inventorySize)

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

	m := mosaic.Compose(src, units, units, p)

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
