package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"
)

func writeImage(m image.Image) {
	out, err := os.Create("./output.png")
	err = png.Encode(out, m)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Test_PixelGrid_Blocks(t *testing.T) {
	// Draw an image with the left side blue and the right side red.
	bounds := image.Rect(0, 0, 40, 30)
	m := image.NewRGBA(bounds)

	top := image.Rect(0, 0, 40, 15)
	bottom := image.Rect(0, 15, 40, 30)

	blue := color.RGBA{0, 0, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	purple := color.RGBA{255, 0, 255, 255}

	draw.Draw(m, top, &image.Uniform{blue}, image.ZP, draw.Src)
	draw.Draw(m, bottom, &image.Uniform{red}, image.ZP, draw.Src)
	//writeImage(m)

	grid := PixelGrid{2, 3, 1}
	// blocks should be w: 20, h: 10
	blocks := grid.Blocks(m)

	if want := 6; len(blocks) != want {
		t.Fatalf("len(blocks) got %d, want %d", len(blocks), want)
	}
	tests := []struct {
		r image.Rectangle
		c color.Color
	}{
		// row 1
		{image.Rect(0, 0, 20, 10), blue},
		{image.Rect(20, 0, 40, 10), blue},
		// row 2
		{image.Rect(0, 10, 20, 20), purple},
		{image.Rect(20, 10, 40, 20), purple},
		// row 3
		{image.Rect(0, 20, 20, 30), red},
		{image.Rect(20, 20, 40, 30), red},
	}
	for i, test := range tests {
		got := blocks[i]
		rect := test.r
		color := test.c
		if got.Rectangle != rect {
			t.Errorf("block %d, rect got %v, want %v", i, got.Rectangle, rect)
		}
		if got.Color != color {
			t.Errorf("block %d, color got %v, want %v", i, got.Color, color)

		}
	}
}

func Test_PixelGrid_MosaicImage(t *testing.T) {
	// Draw an image with the left side blue and the right side red.
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))

	left := image.Rect(0, 0, 320, 480)
	right := image.Rect(320, 0, 640, 480)

	blue := color.RGBA{0, 0, 255, 255}
	red := color.RGBA{255, 0, 0, 255}

	draw.Draw(m, left, &image.Uniform{blue}, image.ZP, draw.Src)
	draw.Draw(m, right, &image.Uniform{red}, image.ZP, draw.Src)

	// Define a grid then extrapolate a new image.
	grid := PixelGrid{3, 2, 1}
	mo := grid.MosaicImage(m, 2000, 2000)

	// Check output pixels.
	if want := 1777; mo.Bounds().Dx() != want {
		t.Errorf("Dx got %d, want %d", mo.Bounds().Dx(), want)
	}
	if want := 2000; mo.Bounds().Dy() != want {
		t.Errorf("Dy got %d, want %d", mo.Bounds().Dy(), want)
	}
	// Check color palette.
	if want := 6; len(mo.blocks) != want {
		t.Errorf("len(p.Palette) got %d, want %d", len(mo.blocks), want)
	}

}

func Test_AverageColorOfRect_uniform(t *testing.T) {
	c := color.RGBA{100, 120, 140, 255}
	m := image.NewUniform(c)
	a := AverageColorOfRect(m, image.Rect(0, 0, 100, 100), 20)
	if a != c {
		t.Errorf("AverageColorOfRect() got %v, want %v", a, c)
	}
}

func Test_AverageColorOfRect_mixed_horizontal(t *testing.T) {
	// Draw an image with the left side blue and the right side red.
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))

	left := image.Rect(0, 0, 320, 480)
	right := image.Rect(320, 0, 640, 480)

	blue := color.RGBA{0, 0, 255, 255}
	red := color.RGBA{255, 0, 0, 255}

	draw.Draw(m, left, &image.Uniform{blue}, image.ZP, draw.Src)
	draw.Draw(m, right, &image.Uniform{red}, image.ZP, draw.Src)

	// Average over the entire image.
	a := AverageColorOfRect(m, m.Bounds(), 20)

	// Average is purple.
	purple := color.RGBA{255, 0, 255, 255}
	if want := purple; a != want {
		t.Errorf("AverageColorOfRect() got %v, want %v", a, want)
	}
}

func Test_AverageColorOfRect_mixed_vertical(t *testing.T) {
	// Draw an image with the left side blue and the right side red.
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))

	top := image.Rect(0, 0, 640, 240)
	bottom := image.Rect(0, 240, 640, 480)

	blue := color.RGBA{0, 0, 255, 255}
	red := color.RGBA{255, 0, 0, 255}

	draw.Draw(m, top, &image.Uniform{blue}, image.ZP, draw.Src)
	draw.Draw(m, bottom, &image.Uniform{red}, image.ZP, draw.Src)

	// Average over the entire image.
	a := AverageColorOfRect(m, m.Bounds(), 20)

	// Average is purple.
	purple := color.RGBA{255, 0, 255, 255}
	if want := purple; a != want {
		t.Errorf("AverageColorOfRect() got %v, want %v", a, want)
	}
	writeImage(m)
}
