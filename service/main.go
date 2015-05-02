package main

import (
	"encoding/json"
	"image"
	"log"
	"net/http"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
)

var imageInventory *mosaic.ImageInventory
var instagramClient *instagram.Client

func init() {
	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleAddTagToInventory(w, r)
			return
		}
		handleGetInventory(w, r)
	})
}

func main() {
	cache := mosaic.FileImageCache{"./cache"}
	imageInventory = mosaic.NewImageInventory(cache)
	instagramClient = instagram.NewClient()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// POST /mosaic?tag=<tag> DATA

type mosaicJob struct {
	OK bool   `json:"ok"`
	ID string `json:"id"`
}

// GET /mosaic?id=<id>

type Mosaic image.Image

// POST /inventory?tag=<tag>

type inventoryReq struct {
	OK bool `json:"ok"`
}

func handleAddTagToInventory(w http.ResponseWriter, r *http.Request) {
	tag := r.FormValue("tag")
	if tag == "" {
		respondErr(w, http.StatusBadRequest, "missing 'tag' param")
		return
	}
	fetchCount := 800
	go func() {
		if err := imageInventory.Fetch(instagramClient, tag, fetchCount); err != nil {
			log.Printf("Failed to fetch tag %s: %s", tag, err)
		}
	}()
	res := &inventoryReq{true}
	respondOK(w, res)
}

// GET /inventory

type inventoryRes struct {
	OK bool `json:"ok"`
	// Map of tag to images count.
	Images map[string]int `json:"images"`
}

func handleGetInventory(w http.ResponseWriter, r *http.Request) {
	res := &inventoryRes{true, make(map[string]int)}
	for t, keys := range imageInventory.Tags {
		res.Images[t] = len(keys)
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
