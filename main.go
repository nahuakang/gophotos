package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nahuakang/gophotos/views"
)

var homeView *views.View
var contactView *views.View

// Handler function for home
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(homeView.Render(w, nil))
}

// Handler function for contact
func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	must(contactView.Render(w, nil))
}

// Helper function to panic if View.Render returns an error
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	homeView = views.NewView("bootstrap", "views/home.gohtml")
	contactView = views.NewView("bootstrap", "views/contact.gohtml")

	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/contact", contact)

	http.ListenAndServe(":3000", r)
}
