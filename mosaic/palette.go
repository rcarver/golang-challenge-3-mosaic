package mosaic

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

type ImagePalette struct {
	color.Palette
	ImgX, ImgY    int
	solidFallback bool
	images        map[int][]image.Image
	indices       map[int]int
}

func NewImagePalette(colors, width, height int) *ImagePalette {
	return &ImagePalette{
		Palette:       make(color.Palette, 0, colors),
		ImgX:          width,
		ImgY:          height,
		solidFallback: false,
		images:        make(map[int][]image.Image),
		indices:       make(map[int]int),
	}
}

func NewSolidPalette(palette color.Palette, width, height int) *ImagePalette {
	return &ImagePalette{
		Palette:       palette,
		ImgX:          width,
		ImgY:          height,
		solidFallback: true,
	}
}

// Add adds an image to the palette.
func (p *ImagePalette) Add(m image.Image) {
	c := AverageColorOfRect(m, m.Bounds(), 1)
	// If we don't have a full color palette, use every image as a new
	// entry (unless it's a dup).
	if len(p.Palette) < cap(p.Palette) {
		var found bool
		for _, x := range p.Palette {
			if x == c {
				found = true
				break
			}
		}
		if found {
			//fmt.Printf("Add(%v) has color\n", c)
		} else {
			p.Palette = append(p.Palette, c)
			//fmt.Printf("Add(%v) new color\n", c)
		}
	}
	// Index images by their nearest color in the palette.
	i := p.Index(c)
	// TODO: crop image to ImgX, ImgY
	p.images[i] = append(p.images[i], m)
	//fmt.Printf("Add(%v) %d\n", c, len(p.images[i]))
}

// Size returns the number of colors in the palette.
func (p *ImagePalette) Size() int {
	return len(p.Palette)
}

// AtColor returns an image whose average color is closest to c.
func (p *ImagePalette) AtColor(c color.Color) image.Image {
	i := p.Index(c)
	x := p.Convert(c)
	a, ok := p.images[i]
	if ok {
		idx := p.indices[i]
		p.indices[i]++
		if p.indices[i] > len(p.images[i]) {
			p.indices[i] = 0
			idx = 0
		}
		fmt.Printf("AtColor() Got image for %v, choosing idx %d of %d\n", x, idx, len(p.images[i]))
		return a[idx]
	}
	if p.solidFallback {
		return image.NewUniform(x)
	}
	fmt.Printf("AtColor() Missing image for %v\n", x)
	return nil
}

// AverageColorOfRect calcluates the average color of an area of an image. Step
// determines how many pixels to sample, 1 being every pixel, 10 being every
// 10th pixel.
func AverageColorOfRect(m image.Image, bounds image.Rectangle, sample float64) color.Color {
	if sample < 1 || sample > 1 {
		sample = 1
	}
	r, g, b := uint64(0), uint64(0), uint64(0)
	c := uint64(0)
	sample = 1 - sample
	xStep := int(math.Max(float64(bounds.Dx())*sample, 1))
	yStep := int(math.Max(float64(bounds.Dy())*sample, 1))
	//fmt.Printf("> Avg %v step:%d x:%d y:%d\n", bounds, sample, xStep, yStep)
	for y := bounds.Min.Y; y <= bounds.Max.Y; y += yStep {
		for x := bounds.Min.X; x <= bounds.Max.X; x += xStep {
			xr, xg, xb, _ := m.At(x, y).RGBA()
			r += uint64(xr)
			g += uint64(xg)
			b += uint64(xb)
			c++
			//fmt.Printf("  %d,%d - %d %d %d\n", x, y, xr, xg, xb)
		}
	}
	color := color.RGBA64{uint16(r / c), uint16(g / c), uint16(b / c), 65535}
	//fmt.Printf("< Avg %v %d %d %d, c:%d, %v\n", bounds, r, g, b, c, color)
	return color
}
