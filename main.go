package main

import (
	"log"
	"net/http"
	"os"

	middleware "github.com/02priyeshraj/Hotel_Management_Backend/middlewares"
	routes "github.com/02priyeshraj/Hotel_Management_Backend/routes"
	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// Load environment variables
	LoadEnv()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := mux.NewRouter()

	// Public Routes (No Authentication)
	routes.UserPublicRoutes(router)

	//Authentication Middleware to Protected Routes
	securedRoutes := router.PathPrefix("/").Subrouter()
	securedRoutes.Use(middleware.Authentication)
	routes.UserProtectedRoutes(securedRoutes)
	routes.TableProtectedRoutes(securedRoutes)
	routes.MenuProtectedRoutes(securedRoutes)
	routes.FoodProtectedRoutes(securedRoutes)
	routes.OrderProtectedRoutes(securedRoutes)
	routes.OrderItemProtectedRoutes(securedRoutes)
	routes.InvoiceProtectedRoutes(securedRoutes)

	log.Printf("Server running on port %s", port)
	http.ListenAndServe(":"+port, router)
}
