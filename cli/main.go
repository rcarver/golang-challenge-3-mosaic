package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
	"github.com/rcarver/golang-challenge-3-mosaic/service"
)

var (
	command    string
	tag        string
	tagDirName string
	imgDirName string
	inName     string
	outName    string
	units      int
	numImages  int
	solid      bool
)

var help = `
# Download images by tag
mosaic -run download -dir images -tag balloon 

# Generate mosaic with tag
mosaic -run generate -dir images -tag balloon -in balloon.jpg -out balloon-mosaic.jpg
`

func init() {
	flag.StringVar(&command, "run", "", "command to run: download | generate | serve")

	// Download, Generate
	flag.StringVar(&tag, "tag", "cat", "image tag to use")
	flag.StringVar(&tagDirName, "dir", "./images", "dir to store images by tag")

	// Download
	flag.IntVar(&numImages, "num", 1000, "number of images to download")

	// Generate
	flag.StringVar(&inName, "in", "", "image file to read")
	flag.StringVar(&outName, "out", "./mosaic.jpg", "image file to write")
	flag.StringVar(&imgDirName, "imgdir", "", "dir to find images (uses tagdir + tag by default)")
	flag.IntVar(&units, "units", 40, "number of units wide to generate the mosaic")
	flag.BoolVar(&solid, "solid", false, "generate a mosaic with solid colors, not images")
}

func main() {
	flag.Parse()

	// dir is imgDirName if set, or join(tagDirName, tag)
	var dir string
	if imgDirName != "" {
		dir = imgDirName
	} else {
		dir = path.Join(tagDirName, tag)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Error initializing: %s\n", err)
	}

	// inventory reads and writes from the dir.
	inventory := newInventory(dir)

	switch command {
	case "download":
		if err := downloadImages(tag, numImages, inventory); err != nil {
			fmt.Printf("Download error: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "generate":
		if inName == "" {
			fmt.Printf("Missing -in file\n")
			os.Exit(1)
		}

		// Read and decode input image.
		in, err := os.Open(inName)
		if err != nil {
			fmt.Printf("Error initializing: %s\n", err)
			os.Exit(1)
		}
		defer in.Close()
		src, _, err := image.Decode(in)
		if err != nil {
			fmt.Printf("Error initializing: %s\n", err)
			os.Exit(1)
		}

		// Generate the mosaic.
		img, err := generateMosaic(src, tag, units, solid, inventory)
		if err != nil {
			fmt.Printf("Error generating: %s\n", err)
			os.Exit(1)
		}

		// Encode and write the output image.
		out, err := os.Create(outName)
		if err != nil {
			fmt.Printf("Error outputting: %s\n", err)
			os.Exit(1)
		}
		defer out.Close()
		err = jpeg.Encode(out, img, nil)
		if err != nil {
			fmt.Printf("Error outputting: %s\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	case "serve":
		service.Serve()
		os.Exit(0)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func newInventory(dir string) *mosaic.ImageInventory {
	cache := mosaic.FileImageCache{dir}
	return mosaic.NewImageInventory(cache)
}

func downloadImages(tag string, numImages int, inv *mosaic.ImageInventory) error {
	api := instagram.NewClient()
	if err := inv.Fetch(api, tag, numImages); err != nil {
		return err
	}
	return nil
}

var (
	// Number of colors in the mosaic color palette.
	paletteSize = 256
	// Size of the mosaic thumbnail images.
	thumbSize = 150
)

func generateMosaic(src image.Image, tag string, units int, solid bool, inv *mosaic.ImageInventory) (image.Image, error) {
	p := mosaic.NewImagePalette(paletteSize, thumbSize, thumbSize)
	if !solid {
		if err := inv.PopulatePalette(p); err != nil {
			return nil, err
		}
	}
	return mosaic.Compose(src, units, units, p), nil
}
