package controllers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nahuakang/gophotos/context"
	"github.com/nahuakang/gophotos/models"
	"github.com/nahuakang/gophotos/views"
)

// ShowGallery is the show galleries route
const ShowGallery = "show_gallery"

// NewGalleries returns a new Galleries controller
func NewGalleries(gs models.GalleryService, r *mux.Router) *Galleries {
	return &Galleries{
		New:      views.NewView("bootstrap", "galleries/new"),
		ShowView: views.NewView("bootstrap", "galleries/show"),
		EditView: views.NewView("bootstrap", "galleries/edit"),
		gs:       gs,
		router:   r,
	}
}

// Galleries is the controller for galleries
type Galleries struct {
	New      *views.View
	ShowView *views.View
	EditView *views.View
	gs       models.GalleryService
	router   *mux.Router
}

// GalleryForm represents a form for new gallery
type GalleryForm struct {
	Title string `schema:"title"`
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

	user := context.User(r.Context())
	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}

	url, err := g.router.Get(ShowGallery).URL("id", strconv.Itoa(int(gallery.ID)))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, url.Path, http.StatusFound)
}

// Show handles GET /galleries/:id requests
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return // galleryByID already handled the errors
	}

	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, vd)
}

// Edit handles GET /galleries/:id/edit requests
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return // galleryByID already handled the errors
	}

	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "You do not have permission to edit "+
			"this gallery", http.StatusForbidden)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	g.EditView.Render(w, vd)
}

func (g *Galleries) galleryByID(w http.ResponseWriter, r *http.Request) (*models.Gallery, error) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}

	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			http.Error(w, "Whoops! Something went wrong", http.StatusInternalServerError)
		}
		return nil, err
	}

	return gallery, nil
}
