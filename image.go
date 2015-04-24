package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"io"
	"os"
)

type TileImage struct {
	bytes.Buffer
}

type TargetImage struct {
	bytes.Buffer
	image.Image
}

type colorGrid struct {
	W      int
	H      int
	pixels []color.Color
}

// PixelGrid defines a W by H units grid. For example, PixelGrid{4, 3} is:
//   ****
//   ****
//   ****
type PixelGrid struct {
	W int
	H int
}

// PixelGridBlocks is a set of image.Rectangle representing the grid units of
// an image.
type PixelGridBlocks struct {
	w     int
	rects []image.Rectangle
}

// At returns the image.Rectangle of the block at X,Y.
func (b PixelGridBlocks) At(x, y int) image.Rectangle {
	return b.rects[(y*b.w)+x]
}

// Blocks calculates the bounds of each grid unit given the bounds of an image.
func (g PixelGrid) Blocks(bounds image.Rectangle) PixelGridBlocks {
	// Calculate pixels size of each block.
	size := bounds.Size()
	px, py := (size.X / g.W), (size.Y / g.H)
	// Allocate enough space to hold all blocks.
	rects := make([]image.Rectangle, g.W*g.H)
	// Find the bounds of each block.
	for i := range rects {
		x := i % g.W
		y := i / g.W
		rects[i] = image.Rectangle{
			image.Point{x * px, y * py},
			image.Point{(x + 1) * px, (y + 1) * py},
		}
		//fmt.Printf("Block %d (%d,%d) at %v\n", i, x, y, rects[i])
	}
	return PixelGridBlocks{g.W, rects}
}

// AverageColorOfRect calcluates the average color of an area of an image. Step
// determines how many pixels to sample, 1 being every pixel, 10 being every
// 10th pixel.
func AverageColorOfRect(m image.Image, bounds image.Rectangle, step int) color.Color {
	if step <= 0 {
		step = 1
	}
	r, g, b := 0, 0, 0
	c := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			xr, xg, xb, _ := m.At(x, y).RGBA()
			r += int(xr)
			g += int(xg)
			b += int(xb)
			c++
			fmt.Printf("%d: %d %d %d\n", c, xr, xg, xb)
		}
	}
	return color.RGBA{uint8(r / c), uint8(g / c), uint8(b / c), 255}
}

func imageFromFile(path string) (image.Image, error) {
	fi, err := os.Open(path)
	defer fi.Close()
	if err != nil {
		return nil, err
	}
	return decodeImage(fi)
}

func decodeImage(reader io.Reader) (image.Image, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}
	return img, err
}
