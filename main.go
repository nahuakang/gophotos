package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nahuakang/gophotos/controllers"
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

	// Controllers
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)
	galleriesController := controllers.NewGalleries(services.Gallery)

	r := mux.NewRouter()
	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.HandleFunc("/signup", usersController.New).Methods("GET")
	r.HandleFunc("/signup", usersController.Create).Methods("POST")
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")
	r.Handle("/galleries/new", galleriesController.New).Methods("GET")

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
