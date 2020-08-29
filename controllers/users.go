package controllers

import (
	"net/http"

	"github.com/nahuakang/gophotos/views"
)

// NewUsers creates a new Users
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

// Users contains data for users
type Users struct {
	NewView *views.View
}

// New renders the form where a user creates a new user account
// GET /signup
func (u *Users) New(w http.ResponseWriter, r *http.Request) {
	if err := u.NewView.Render(w, nil); err != nil {
		panic(err)
	}
}
