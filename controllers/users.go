package controllers

import (
	"fmt"
	"net/http"

	"github.com/nahuakang/gophotos/models"
	"github.com/nahuakang/gophotos/views"
)

// NewUsers creates a new Users
func NewUsers(us *models.UserService) *Users {
	return &Users{
		NewView:   views.NewView("bootstrap", "users/new"),
		LoginView: views.NewView("bootstrap", "users/login"),
		us:        us,
	}
}

// Users Controller contains data for users
type Users struct {
	NewView   *views.View
	LoginView *views.View
	us        *models.UserService
}

// LoginForm contains email and password
type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
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

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}
	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "User is", user)
}

// Login processes the login form when a user logs in as existing user
//
// POST /login
func (u *Users) Login(w http.ResponseWriter, r *http.Request) {
	form := LoginForm{}
	if err := parseForm(r, &form); err != nil {
		panic(err)
	}
	// We will eventually authenticate the user here
	fmt.Println("Login form is", form)
}
