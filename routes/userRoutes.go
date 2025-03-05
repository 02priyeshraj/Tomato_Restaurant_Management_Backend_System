package routes

import (
	controller "github.com/02priyeshraj/Hotel_Management_Backend/controllers"

	"github.com/gorilla/mux"
)

func PublicRoutes(router *mux.Router) {
	router.HandleFunc("/users/signup", controller.SignUp).Methods("POST")
	router.HandleFunc("/users/login", controller.Login).Methods("POST")
}

func ProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/users", controller.GetUsers).Methods("GET")
	router.HandleFunc("/users/{user_id}", controller.GetUser).Methods("GET")
	router.HandleFunc("/users/logout", controller.Logout).Methods("POST")
}
