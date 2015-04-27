package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func Test_PixelGrid_Blocks(t *testing.T) {
	g := PixelGrid{2, 3}
	bounds := image.Rect(0, 0, 40, 30)
	// blocks should be w: 20, h: 10
	blocks := g.Blocks(bounds)

	if want := 6; len(blocks.rects) != want {
		t.Fatalf("len(blocks) got %d, want %d", len(blocks.rects), want)
	}
	tests := []struct {
		x int
		y int
		b [4]int
	}{
		// row 1
		{0, 0, [4]int{0, 0, 20, 10}},
		{1, 0, [4]int{20, 0, 40, 10}},
		// row 2
		{0, 1, [4]int{0, 10, 20, 20}},
		{1, 1, [4]int{20, 10, 40, 20}},
		// row 3
		{0, 2, [4]int{0, 20, 20, 30}},
		{1, 2, [4]int{20, 20, 40, 30}},
	}
	for i, test := range tests {
		want := image.Rect(
			test.b[0], test.b[1],
			test.b[2], test.b[3],
		)
		got := blocks.rects[i]
		if got != want {
			t.Errorf("block %d, got %v, want %v", i, got, want)
		}
		at := blocks.At(test.x, test.y)
		if at != want {
			t.Errorf("block at %d,%d, got %v want %v", test.x, test.y, at, want)
		}
	}
}

func Test_PixelGrid_ColorPalette(t *testing.T) {
	// Draw an image with the left side blue and the right side red.
	m := image.NewRGBA(image.Rect(0, 0, 640, 480))

	left := image.Rect(0, 0, 320, 480)
	right := image.Rect(320, 0, 640, 480)

	blue := color.RGBA{0, 0, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	purple := color.RGBA{255, 0, 255, 255}

	draw.Draw(m, left, &image.Uniform{blue}, image.ZP, draw.Src)
	draw.Draw(m, right, &image.Uniform{red}, image.ZP, draw.Src)

	// Palette of image thats 3x2 units.
	grid := PixelGrid{3, 2}
	palette := grid.ColorPalette(m)

	if want := 3; len(palette) != want {
		t.Fatalf("len(palette) got %d, want %d", len(palette), want)
	}
	tests := []color.RGBA{
		blue, purple, red,
	}
	for i, want := range tests {
		got := palette[i]
		if got != want {
			t.Errorf("palette %d, got %v, want %v", i, got, want)
		}
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

func Test_AverageColorOfRect_mixed(t *testing.T) {
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
