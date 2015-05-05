package service

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"net/http"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
)

func init() {
	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleAddTagToInventory(w, r)
			return
		}
		handleGetInventory(w, r)
	})
	http.HandleFunc("/mosaics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleCreateMosaic(w, r)
			return
		}
		if r.FormValue("id") != "" {
			handleGetMosaic(w, r)
			return
		}
		handleListMosaics(w, r)

	})
}

// thumbs is the database of thumbnail images available.
var thumbs *thumbInventory

// mosaics is the database of mosaics generated.
var mosaics *mosaicInventory

func Serve() {
	thumbsCache := mosaic.FileImageCache{"./cache/thumbs"}
	mosaicsCache := mosaic.FileImageCache{"./cache/mosaics"}

	thumbs = &thumbInventory{
		ImageInventory: mosaic.NewImageInventory(thumbsCache),
		api:            instagram.NewClient(),
		tags:           make(map[string]chan bool),
	}
	mosaics = &mosaicInventory{
		ImageCache: mosaicsCache,
	}

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GET /mosaics
// List all mosaics that have been created.

type mosaicsListRes struct {
	OK      bool        `json:"ok"`
	Mosaics []mosaicRes `json:"mosaics"`
}

type mosaicRes struct {
	ID     string `json:"id"`
	Tag    string `json:"tag"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

func handleListMosaics(w http.ResponseWriter, r *http.Request) {
	res := &mosaicsListRes{
		true,
		make([]mosaicRes, mosaics.Size()),
	}
	for i, m := range mosaics.List() {
		res.Mosaics[i] = mosaicRes{
			ID:     m.ID,
			Tag:    m.Tag,
			Status: m.Status,
			URL:    fmt.Sprintf("/mosaics?id=%s", m.ID),
		}
	}
	respondOK(w, res)
}

// POST /mosaics?tag=<tag> img=<FILE>
// Create a new mosaic.

var (
	units       = 40
	unitX       = 150
	unitY       = 150
	paletteSize = 256
)

type mosaicReq struct {
	OK bool   `json:"ok"`
	ID string `json:"id"`
}

func handleCreateMosaic(w http.ResponseWriter, r *http.Request) {
	// Read tag.
	tag := r.FormValue("tag")
	if tag == "" {
		respondErr(w, http.StatusBadRequest, "missing 'tag' param")
		return
	}
	ch := thumbs.AddTag(tag)

	// Read image upload.
	fi, _, err := r.FormFile("img")
	if err != nil {
		respondErr(w, http.StatusBadRequest, "upload failed")
	}
	defer fi.Close()
	in, _, err := image.Decode(fi)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "image parsing failed")
	}

	// Generate mosaic offline.
	id := mosaics.NextID()
	go func() {
		if err := mosaics.Create(id, tag); err != nil {
			log.Printf("Failed to store mosaic: %s", err)
			return
		}

		log.Printf("Waiting for tags...\n")
		<-ch
		log.Printf("Tags are ready...\n")

		log.Printf("Mosaic[%s] Create Palette...", id)
		p := mosaic.NewImagePalette(paletteSize, unitX, unitY)
		if err := thumbs.PopulatePalette(tag, p); err != nil {
			log.Printf("Failed to populate palette: %s", err)
			if err := mosaics.SetStatus(id, MosaicStatusFailed); err != nil {
				log.Printf("Failed to set mosaic failed: %s", err)
			}
			return
		}
		log.Printf("Mosaic[%s] Create Palette Done with %d colors, %d images.", id, p.Size(), p.NumImages())

		if err := mosaics.SetStatus(id, MosaicStatusWorking); err != nil {
			log.Printf("Failed to set working status: %s", err)
			return
		}

		log.Printf("Mosaic[%s] Compose...", id)
		out := mosaic.Compose(in, units, units, p)
		log.Printf("Mosaic[%s] Compose Done.", id)

		if err := mosaics.StoreImage(id, out); err != nil {
			log.Printf("Failed to store mosaic image: %s", err)
			return
		}
	}()

	// Respond immediately with the ID.
	res := mosaicReq{true, id}
	respondOK(w, res)
}

// GET /mosaics?id=<id>
// Get a mosaic that was created.

func handleGetMosaic(w http.ResponseWriter, r *http.Request) {
	// Read id.
	id := r.FormValue("id")
	if id == "" {
		respondErr(w, http.StatusBadRequest, "missing 'id' param")
		return
	}
	if !mosaics.Has(id) {
		http.NotFound(w, r)
		return
	}
	img, err := mosaics.Get(id)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "getting image")
	}
	err = jpeg.Encode(w, img, nil)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "writing jpeg")
	}
	w.Header().Set("Content-Type", "image/jpg")
}

// POST /inventory?tag=<tag>
// Add images to the thumbnails inventory.

var (
	fetchImages = 100
)

type inventoryReq struct {
	OK bool `json:"ok"`
}

func handleAddTagToInventory(w http.ResponseWriter, r *http.Request) {
	tag := r.FormValue("tag")
	if tag == "" {
		respondErr(w, http.StatusBadRequest, "missing 'tag' param")
		return
	}
	thumbs.AddTag(tag)
	res := &inventoryReq{true}
	respondOK(w, res)
}

// GET /inventory
// Get information about the thumbnails inventory.

type inventoryRes struct {
	OK bool `json:"ok"`
	// Map of tag to images count.
	Images []inventoryImageRes `json:"images"`
}

type inventoryImageRes struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

func handleGetInventory(w http.ResponseWriter, r *http.Request) {
	res := &inventoryRes{
		true,
		make([]inventoryImageRes, 0, len(thumbs.Tags)),
	}
	for t, keys := range thumbs.Tags {
		res.Images = append(res.Images, inventoryImageRes{
			Tag:   t,
			Count: len(keys),
		})
	}
	respondOK(w, res)
}

// Utils

type errorRes struct {
	OK     bool   `json:"ok"`
	Reason string `json:"reason"`
}

func respondOK(w http.ResponseWriter, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func respondErr(w http.ResponseWriter, code int, msg string) {
	data := errorRes{false, msg}
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(js)
}
