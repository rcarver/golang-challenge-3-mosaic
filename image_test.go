package main

import (
	"image"
	"image/color"
	"testing"
)

func Test_PixelGrid_Blocks(t *testing.T) {
	g := PixelGrid{2, 3}
	bounds := image.Rectangle{
		image.Point{0, 0},
		image.Point{40, 30},
	}
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
		want := image.Rectangle{
			image.Pt(test.b[0], test.b[1]),
			image.Pt(test.b[2], test.b[3]),
		}
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

func Test_AverageColorOfRect_uniform(t *testing.T) {
	c := color.RGBA{100, 120, 140, 255}
	m := image.NewUniform(c)
	a := AverageColorOfRect(
		m,
		image.Rectangle{
			image.Pt(0, 0),
			image.Pt(100, 100),
		},
		20,
	)
	if a != c {
		t.Errorf("AverageColorOfRect() got %v, want %v", a, c)
	}
}
