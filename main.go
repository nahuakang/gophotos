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
	defer services.User.Close()
	services.User.AutoMigrate()

	// Controllers
	staticController := controllers.NewStatic()
	usersController := controllers.NewUsers(services.User)

	r := mux.NewRouter()
	r.Handle("/", staticController.Home).Methods("GET")
	r.Handle("/contact", staticController.Contact).Methods("GET")
	r.HandleFunc("/signup", usersController.New).Methods("GET")
	r.HandleFunc("/signup", usersController.Create).Methods("POST")
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersController.CookieTest).Methods("GET")

	http.ListenAndServe(":3000", r)
}
