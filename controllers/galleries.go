package controllers

import (
	"fmt"
	"net/http"

	"github.com/nahuakang/gophotos/models"
	"github.com/nahuakang/gophotos/views"
)

// NewGalleries returns a new Galleries controller
func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		New: views.NewView("bootstrap", "galleries/new"),
		gs:  gs,
	}
}

// Galleries is the controller for galleries
type Galleries struct {
	New *views.View
	gs  models.GalleryService
}

// Create handles POST requests for galleries
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm

	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}

	gallery := models.Gallery{
		Title: form.Title,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}

	fmt.Fprintln(w, gallery)
}

// GalleryForm represents a form for new gallery
type GalleryForm struct {
	Title string `schema:"title"`
}
