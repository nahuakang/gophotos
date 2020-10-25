package controllers

import (
	"github.com/nahuakang/gophotos/models"
	"github.com/nahuakang/gophotos/views"
)

// Galleries is the controller for galleries
type Galleries struct {
	New *views.View
	gs  models.GalleryService
}

// NewGalleries returns a new Galleries controller
func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		New: views.NewView("bootstrap", "galleries/new"),
		gs:  gs,
	}
}
