package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nahuakang/gophotos/controllers"
	"github.com/nahuakang/gophotos/middleware"
	"github.com/nahuakang/gophotos/models"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "qwerty"
	dbname   = "gophotos_dev"
)

func main() {
	// Create a DB connection string to create model services
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	// Mux Router
	r := mux.NewRouter()
	// Controllers
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)
	galleriesController := controllers.NewGalleries(services.Gallery, r)

	// Middleware
	requireUserMw := middleware.RequireUser{
		UserService: services.User,
	}

	// galleriesController.New is http.Handler, use Apply
	newGallery := requireUserMw.Apply(galleriesController.New)
	// galleriesController.Create is http.HandlerFunc, use ApplFn
	createGallery := requireUserMw.ApplyFn(galleriesController.Create)

	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.HandleFunc("/signup", usersController.New).Methods("GET")
	r.HandleFunc("/signup", usersController.Create).Methods("POST")
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")
	r.Handle("/galleries/new", newGallery).Methods("GET")
	r.HandleFunc("/galleries", createGallery).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesController.Show).
		Methods("GET").
		Name(controllers.ShowGallery)

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
