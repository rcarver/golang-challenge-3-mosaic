package mosaic

import (
	"image"
	"image/color"
	"image/draw"
)

func New(in image.Image, w, h int, p *ImagePalette) image.Image {
	grid := PixelGrid{w, h, .5}
	return grid.Mosaic(in, p)
}

// PixelGrid defines a W by H units grid. For example, PixelGrid{4, 3} is:
//   ****
//   ****
//   ****
type PixelGrid struct {
	W      int
	H      int
	Sample float64
}

// GridImage reduces input image into a grid of pixels.
func (g PixelGrid) GridImage(in image.Image) image.Image {
	// Allocate the output image.
	out := image.NewRGBA(image.Rect(0, 0, g.W, g.H))
	// Calculate pixels size of each block in the input.
	bounds := in.Bounds()
	size := bounds.Size()
	px, py := (size.X / g.W), (size.Y / g.H)
	// Find the color of each block in the input and assign that color to
	// the output.
	for i := range make([]struct{}, g.W*g.H) {
		x := i % g.W
		y := i / g.W
		rect := image.Rectangle{
			image.Point{x * px, y * py},
			image.Point{(x + 1) * px, (y + 1) * py},
		}
		color := AverageColorOfRect(in, rect, g.Sample)
		out.Set(x, y, color)
		//fmt.Printf("Block %d (%d,%d) at %v %v\n", i, x, y, rect, color)
	}
	return out
}

// DitherImage reduces the input image into a grid of pixels that have been
// dithered to a color palette.
func (g PixelGrid) DitherImage(in image.Image, p color.Palette) image.Image {
	m := g.GridImage(in)
	o := image.NewPaletted(m.Bounds(), p)
	draw.FloydSteinberg.Draw(o, o.Bounds(), m, image.ZP)
	return o
}

// Mosaic generates a new image, a composite of images from ImagePalette.
func (g PixelGrid) Mosaic(in image.Image, p *ImagePalette) image.Image {
	// Create the dither pattern image.
	d := g.DitherImage(in, p.Palette)
	db := d.Bounds()

	// Create an output image.
	dx, dy := db.Dx(), db.Dy()
	out := image.NewRGBA(image.Rect(0, 0, dx*p.ImgX, dy*p.ImgY))

	// Iterate over the dither pattern and pull an image from the palette,
	// then draw it onto the output at its size.
	for y := db.Min.Y; y < db.Max.Y; y += 1 {
		for x := db.Min.X; x < db.Max.X; x += 1 {
			c := d.At(x, y)
			t := p.AtColor(c)
			rect := image.Rect(
				x*p.ImgX,
				y*p.ImgY,
				(x+1)*p.ImgX,
				(y+1)*p.ImgY,
			)
			//fmt.Printf("Draw %d,%d %v\n", x, y, rect)
			draw.Draw(out, rect, t, image.ZP, draw.Src)
		}
	}
	return out
}
