package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"io"
	"math"
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

// ImagePalette provides images indexed by their color.
type ImagePalette struct {
	images  []image.Image
	palette color.Palette
}

// Add adds an image to the palette.
func (p ImagePalette) Add(m image.Image) {
	c := AverageColorOfRect(m, m.Bounds(), 0)
	p.palette = append(p.palette, c)
	p.images = append(p.images, m)
}

// AtColor returns an image whose average color is closest to c.
func (p ImagePalette) AtColor(c color.Color) image.Image {
	i := p.palette.Index(c)
	return p.images[i]
}

// MosaicImage is an image broken up into color blocks.
type MosaicImage struct {
	blocks []PixelBlock
	draw.Image
}

// Draw pulls images from the source and composites them into the mosaic grid.
func (m *MosaicImage) Draw(source ImagePalette) {
	for _, b := range m.blocks {
		rect := b.Rectangle
		mi := source.AtColor(b.Color)
		if mi != nil {
			d := draw.FloydSteinberg
			d.Draw(m.Image, rect, mi, rect.Min)
		}
	}
}

// PixelBlock is a Rectangle and a Color.
type PixelBlock struct {
	image.Rectangle
	color.Color
}

// PixelGrid defines a W by H units grid. For example, PixelGrid{4, 3} is:
//   ****
//   ****
//   ****
type PixelGrid struct {
	W    int
	H    int
	Step int
}

// Blocks calculates the bounds of each grid unit given the bounds of an image.
func (g PixelGrid) Blocks(m image.Image) []PixelBlock {
	// Calculate pixels size of each block.
	bounds := m.Bounds()
	size := bounds.Size()
	px, py := (size.X / g.W), (size.Y / g.H)
	// Allocate enough space to hold all blocks.
	blocks := make([]PixelBlock, g.W*g.H)
	// Find the bounds of each block.
	for i := range blocks {
		x := i % g.W
		y := i / g.W
		rect := image.Rectangle{
			image.Point{x * px, y * py},
			image.Point{(x + 1) * px, (y + 1) * py},
		}
		color := AverageColorOfRect(m, rect, g.Step)
		blocks[i] = PixelBlock{rect, color}
		//fmt.Printf("Block %d (%d,%d) at %v %v\n", i, x, y, rect, color)
	}
	return blocks
}

// MosaicImage calculates the foundation for a mosaic from the input image. The
// resulting image's dimensions are maxWidth or maxHeight with the other
// dimension sized proportionally to the input image.
func (g PixelGrid) MosaicImage(m image.Image, maxWidth, maxHeight int) *MosaicImage {
	mb := m.Bounds()
	gw, gh := float64(g.W), float64(g.H)
	mw, mh := float64(mb.Dx())/gw, float64(mb.Dy())/gh
	ratio := math.Min(float64(maxWidth)/mw, float64(maxHeight)/mh)
	bounds := image.Rect(0, 0, int(ratio*mw), int(ratio*mh))
	blocks := g.Blocks(m)
	mo := image.NewRGBA(bounds)
	return &MosaicImage{blocks, mo}
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
			//fmt.Printf("%d,%d - %d %d %d\n", x, y, xr, xg, xb)
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
