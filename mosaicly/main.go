package main

import (
	"flag"
	"fmt"
	"image"
	"image/color/palette"
	"image/jpeg"
	"log"
	"os"
	"path"
	"time"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
	"github.com/rcarver/golang-challenge-3-mosaic/service"
)

var (
	command     string
	tag         string
	baseDirName string
	imgDirName  string
	inName      string
	outName     string
	units       int
	numImages   int
	solid       bool
	port        int
)

var help = `
# Download images by tag
mosaic -run download -dir images -tag balloon 

# Generate mosaic with tag
mosaic -run generate -dir images -tag balloon -in balloon.jpg -out balloon-mosaic.jpg
`

func init() {
	flag.StringVar(&command, "run", "", "command to run: fetch | gen | serve")

	// fetch, gen, serve
	flag.StringVar(&baseDirName, "dir", "./cache/thumbs", "dir to store images by tag")

	// fetch, gen
	flag.StringVar(&tag, "tag", "cat", "image tag to use")

	// fetch
	flag.IntVar(&numImages, "num", 1000, "number of images to download")

	// gen
	flag.StringVar(&inName, "in", "", "image file to read")
	flag.StringVar(&outName, "out", "./mosaic.jpg", "image file to write")
	flag.StringVar(&imgDirName, "imgdir", "", "dir to find images (uses tagdir + tag by default)")

	// gen, serve
	flag.IntVar(&units, "units", 40, "number of units wide to generate the mosaic")
	flag.BoolVar(&solid, "solid", false, "generate a mosaic with solid colors, not images")

	// serve
	flag.IntVar(&port, "port", 8080, "port number of the server")
}

func main() {
	flag.Parse()

	// thumbsDir is imgDirName if set, or join(baseDirName, "thumbs", tag)
	var thumbsDir string
	if imgDirName != "" {
		thumbsDir = imgDirName
	} else {
		thumbsDir = path.Join(baseDirName, "thumbs", tag)
	}
	if err := os.MkdirAll(thumbsDir, 0755); err != nil {
		fmt.Printf("Error initializing: %s\n", err)
	}

	// inventory reads and writes from the dir.
	inventory := newInventory(thumbsDir)

	switch command {
	case "fetch":
		if err := downloadImages(tag, numImages, inventory); err != nil {
			fmt.Printf("Download error: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	case "gen":
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
		service.HostPort = fmt.Sprintf(":%d", port)
		service.MosaicsDir = path.Join(baseDirName, "mosaics")
		service.ThumbsDir = path.Join(baseDirName, "thumbs")
		service.ImagesPerTag = numImages
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
	fetcher := instagram.NewTagFetcher(api, tag)
	if err := inv.Fetch(fetcher, numImages); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

var (
	// Number of colors in the mosaic color palette.
	paletteSize = 256
	// Size of the mosaic thumbnail images.
	thumbSize = 150
)

func generateMosaic(src image.Image, tag string, units int, solid bool, inv *mosaic.ImageInventory) (image.Image, error) {
	var p *mosaic.ImagePalette
	if solid {
		p = mosaic.NewSolidPalette(palette.WebSafe)
		log.Printf("Generating %dx%d solid mosaic with %d colors", units, units, p.Size())
	} else {
		p = mosaic.NewImagePalette(paletteSize)
		if err := inv.PopulatePalette(p); err != nil {
			return nil, err
		}
		if p.Size() == 0 {
			return nil, fmt.Errorf("No images are available")
		}
		log.Printf("Generating %dx%d %s mosaic with %d colors and %d images\n", units, units, tag, p.Size(), p.NumImages())
	}
	return mosaic.Compose(src, units, units, thumbSize, thumbSize, p), nil
}
