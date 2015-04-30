package mosaic

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Compose returns a new composite mosaic image from the input source. The output
// image size is determined by the number of units and the size of images in
// the palette - (w * p.ImgX) x (h * p.ImgY).
func Compose(in image.Image, w, h int, p *ImagePalette) image.Image {
	m := Mosaic{w, h, in}
	return m.Compose(p)
}

// samplePixels is the percentage of pixels are sampled to find the average color.
var samplePixels = .5

// sampleRadius is the percentage overlap used for color overaging during
// downsample.
var sampleRadius = .5

// Mosaic is an image that is downsampled to a very coarse pixel grid and then
// rendered with a full image represending each pixel.
type Mosaic struct {
	// UnitsX is how many grid units (pixels) wide.
	UnitsX int
	// UnitsY is how many grid units (pixels) tall.
	UnitsY int
	img    image.Image
}

// Dither generates a new image that has been downsampled and dithered to a
// mosaic grid. The image's dimensions are UnitsX x UnitsY pixels.
func (m Mosaic) Dither(p color.Palette) image.Image {
	down := downsample(m.img, m.UnitsX, m.UnitsY, samplePixels, sampleRadius)
	dith := dither(down, p)
	return dith
}

// Compose generates a new image, a composite of images from ImagePalette. The
// image's dimensinos are (UnitsX * Palette.ImgX) x (UnitsY * Palette.ImgY).
func (m Mosaic) Compose(p *ImagePalette) image.Image {
	// Create the dither pattern image.
	d := m.Dither(p.Palette)
	db := d.Bounds()

	// Create an output image.
	out := image.NewRGBA(image.Rect(0, 0, m.UnitsX*p.ImgX, m.UnitsY*p.ImgY))

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

// dither reduces the colors in an image.
func dither(in image.Image, p color.Palette) image.Image {
	o := image.NewPaletted(in.Bounds(), p)
	draw.FloydSteinberg.Draw(o, o.Bounds(), in, image.ZP)
	return o
}

// downsample reduces an image size.
func downsample(in image.Image, dx, dy int, samplePixels, sampleRadius float64) image.Image {
	// Calculate pixels size of each block in the input.
	ib := in.Bounds()
	px, py := (ib.Dx() / dx), (ib.Dy() / dy)
	// Allocate the output image.
	out := image.NewRGBA(image.Rect(0, 0, dx, dy))
	ob := out.Bounds()
	// Blend regions for color averaging.
	var min = func(a, b int) int {
		return int(math.Min(float64(a), float64(b)))
	}
	var max = func(a, b int) int {
		return int(math.Max(float64(a), float64(b)))
	}
	spx := int(float64(px) * sampleRadius)
	spy := int(float64(py) * sampleRadius)
	// Find the color of each block in the input and assign that color to
	// the output.
	for x := ob.Min.X; x < ob.Max.X; x++ {
		for y := ob.Min.Y; y < ob.Max.Y; y++ {
			rect := image.Rect(
				max(x*px-spx, ib.Min.X),
				max(y*py-spy, ib.Min.Y),
				min((x+1)*px+spx, ib.Max.X),
				min((y+1)*py+spy, ib.Max.Y),
			)
			color := average(in, rect, samplePixels)
			out.Set(x, y, color)
			//fmt.Printf("Block (%d,%d)(%d,%d) at %v %v\n", x, y, spx, spy, rect, color)
		}
	}
	return out
}

// average calcluates the average color of an area of an image.
func average(m image.Image, bounds image.Rectangle, sample float64) color.Color {
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
