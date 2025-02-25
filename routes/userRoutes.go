package routes

import (
	"net/http"

	controller "github.com/02priyeshraj/Hotel_Management_Backend/controllers"

	"github.com/gorilla/mux"
)

func PublicRoutes(router *mux.Router) {
	router.HandleFunc("/users/signup", controller.SignUp).Methods(http.MethodPost)
	router.HandleFunc("/users/login", controller.Login).Methods(http.MethodPost)
}

func ProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/users", controller.GetUsers).Methods(http.MethodGet)
	router.HandleFunc("/users/{user_id}", controller.GetUser).Methods(http.MethodGet)
}
