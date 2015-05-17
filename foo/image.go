package foo

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"io"
	"math"
	"os"
)

// ImagePalette provides images indexed by their color.
type ImagePalette interface {
	Add(m image.Image)
	IsFull() bool
	AtColor(c color.Color) image.Image
}

// NewImagePalette initializes a new color palette backed by images. The
// palette needs to be populated with images before it's useful.
func NewImagePalette(p color.Palette) ImagePalette {
	return &imagePalette{
		palette: p,
		images:  make(map[int][]image.Image),
		indices: make(map[int]int),
	}
}

// NewSolidImagePalette initializes a new color palette backed by solid images.
// The palette does not need to be popluated before it's useful.
func NewSolidImagePalette(p color.Palette) ImagePalette {
	return &solidImagePalette{palette: p}
}

// imagePalette provides images indexed by their color.
type imagePalette struct {
	palette color.Palette
	images  map[int][]image.Image
	indices map[int]int
}

// Add adds an image to the palette.
func (p *imagePalette) Add(m image.Image) {
	c := AverageColorOfRect(m, m.Bounds(), 1)
	i := p.palette.Index(c)
	p.images[i] = append(p.images[i], m)
	fmt.Printf("Add(%v) %d\n", c, len(p.images[i]))
}

func (p imagePalette) IsFull() bool {
	return len(p.images) >= len(p.palette)
}

// AtColor returns an image whose average color is closest to c.
func (p *imagePalette) AtColor(c color.Color) image.Image {
	i := p.palette.Index(c)
	x := p.palette.Convert(c)
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
	fmt.Printf("AtColor() Missing image for %v\n", x)
	return nil
}

// solidImagePalette implements ImagePalette by returning solid images.
type solidImagePalette struct {
	palette color.Palette
}

func (p solidImagePalette) Add(m image.Image) {
	// no op
}

func (p solidImagePalette) IsFull() bool {
	return true
}

// AtColor returns an image with solid color closest to c.
func (p solidImagePalette) AtColor(c color.Color) image.Image {
	i := p.palette.Convert(c)
	//fmt.Printf("AtColor %v %v\n", c, i)
	return image.NewUniform(i)
}

// MosaicImage is an image broken up into color blocks.
type MosaicImage struct {
	draw.Image
	blocks []PixelBlock
}

// Draw pulls images from the source and composites them into the mosaic grid.
func (m *MosaicImage) Draw(source ImagePalette) {
	for _, b := range m.blocks {
		rect := b.Rectangle
		mi := source.AtColor(b.Color)
		if mi != nil {
			d := draw.FloydSteinberg
			//fmt.Printf("comp %v %v\n", rect, b.Color)
			d.Draw(m.Image, rect, mi, image.ZP)
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
	W      int
	H      int
	Sample float64
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
		color := AverageColorOfRect(m, rect, g.Sample)
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
	mw, mh := float64(mb.Dx()), float64(mb.Dy())
	var bounds image.Rectangle
	if mh < mw {
		ratio := mh / mw
		bounds = image.Rect(0, 0, maxWidth, int(float64(maxHeight)*ratio))
	} else {
		ratio := mw / mh
		bounds = image.Rect(0, 0, int(float64(maxWidth)*ratio), maxHeight)
	}
	mo := image.NewRGBA(bounds)
	colors := g.Blocks(m)
	rects := g.Blocks(mo)
	mergeBlocks := make([]PixelBlock, len(colors))
	for i := range colors {
		mergeBlocks[i] = PixelBlock{rects[i].Rectangle, colors[i].Color}
	}
	return &MosaicImage{mo, mergeBlocks}
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
