package controllers

import "github.com/nahuakang/gophotos/views"

// NewStatic returns a controller for static pages
func NewStatic() *Static {
	return &Static{
		Home: views.NewView(
			"bootstrap", "static/home",
		),
		Contact: views.NewView(
			"bootstrap", "static/contact",
		),
	}
}

// Static holds information for static pages
type Static struct {
	Home    *views.View
	Contact *views.View
}
