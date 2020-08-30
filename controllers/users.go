package controllers

import (
	"fmt"
	"net/http"

	"github.com/nahuakang/gophotos/views"
)

// NewUsers creates a new Users
func NewUsers() *Users {
	return &Users{
		NewView: views.NewView("bootstrap", "views/users/new.gohtml"),
	}
}

// Users Controller contains data for users
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

// Create propcesses the signup form when a user creates a new user account
// POST /signup
func (u *Users) Create(w http.ResponseWriter, r *http.Request) {
	// Parse the submitted form
	var form SignupForm // Initialized to fields' zero values
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}

	fmt.Fprintln(w, "Email is", form.Email)
	fmt.Fprintln(w, "Password is", form.Password)
}
