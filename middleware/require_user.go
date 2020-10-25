package middleware

import (
	"fmt"
	"net/http"

	"github.com/nahuakang/gophotos/models"
)

// RequireUser is the middleware that checks if a user is logged in
type RequireUser struct {
	models.UserService
}

// Apply applies middleware to http.Handler interfaces
func (mw *RequireUser) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn returns an http.HandlerFunc that checks if a user is
// logged in and then either calls  next(w, r) if they are, or
// redirect the user to the login page if they are not.
func (mw *RequireUser) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	// Return a dynamically created func(http.ResponseWriter, *http.Request)
	// but also convert it into an http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is logged in.
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			// If user is not logged in, http.Redirect to "/login"
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, err := mw.UserService.ByRemember(cookie.Value)
		if err != nil {
			// If user does not exist in DB, http.Redirect to "/login"
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		fmt.Println("User found: ", user)

		// If user exists, call next(w, r)
		next(w, r)
	})
}
