package service

import (
	"fmt"
	"image"
	"log"
	"sync"

	"github.com/rcarver/golang-challenge-3-mosaic/instagram"
	"github.com/rcarver/golang-challenge-3-mosaic/mosaic"
)

const (
	MosaicStatusNew     = "new"
	MosaicStatusWorking = "working"
	MosaicStatusFailed  = "failed"
	MosaicStatusCreated = "created"
)

var (
	mosaicIDCounter = 0
	imagesPerTag    = 400
)

// Inventory of mosaics that have been created.
type mosaicInventory struct {
	mosaic.ImageCache
	mosaics []*mosaicRecord
}

type mosaicID string

type mosaicRecord struct {
	ID     mosaicID
	Tag    string
	Status string
}

func (i *mosaicInventory) Create(tag string) (*mosaicRecord, error) {
	mosaicIDCounter++
	id := mosaicID(fmt.Sprintf("%d", mosaicIDCounter))
	d := &mosaicRecord{
		ID:     id,
		Tag:    tag,
		Status: MosaicStatusNew,
	}
	i.mosaics = append(i.mosaics, d)
	return d, nil
}

func (i *mosaicInventory) SetStatus(id mosaicID, status string) error {
	for _, d := range i.mosaics {
		if d.ID == id {
			d.Status = status
			break
		}
	}
	return nil
}

func (i *mosaicInventory) StoreImage(id mosaicID, m image.Image) error {
	if err := i.ImageCache.Put(string(id), m); err != nil {
		return err
	}
	if err := i.SetStatus(id, MosaicStatusCreated); err != nil {
		return i.SetStatus(id, MosaicStatusFailed)
	}
	return nil
}

func (i *mosaicInventory) Size() int {
	return len(i.mosaics)
}

func (i *mosaicInventory) List() []*mosaicRecord {
	return i.mosaics
}

// thumbInventory tracks the thumbnails that have been acquired.
type thumbInventory struct {
	*mosaic.ImageInventory
	api *instagram.Client

	mu   sync.Mutex
	tags map[string]chan bool
}

func (i *thumbInventory) AddTag(tag string) chan bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if ch, ok := i.tags[tag]; ok {
		log.Printf("AddTag(%s) already has it\n", tag)
		return ch
	}
	log.Printf("AddTag(%s) beginning fetch\n", tag)
	i.tags[tag] = make(chan bool)
	go func() {
		if err := i.Fetch(i.api, tag, imagesPerTag); err != nil {
			log.Printf("Failed to fetch tag %s: %s", tag, err)
		}
		close(i.tags[tag])
	}()
	return i.tags[tag]
}
