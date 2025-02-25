package main

import (
	"log"
	"net/http"
	"os"

	database "github.com/02priyeshraj/Hotel_Management_Backend/config"
	middleware "github.com/02priyeshraj/Hotel_Management_Backend/middlewares"
	routes "github.com/02priyeshraj/Hotel_Management_Backend/routes"
	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

// LoadEnv loads environment variables from the .env file
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	// Load environment variables
	LoadEnv()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := mux.NewRouter()

	// Public Routes (No Authentication)
	routes.PublicRoutes(router)

	// Apply Authentication Middleware to Protected Routes
	securedRoutes := router.PathPrefix("/").Subrouter()
	securedRoutes.Use(middleware.Authentication)
	routes.ProtectedRoutes(securedRoutes)

	log.Printf("Server running on port %s", port)
	http.ListenAndServe(":"+port, router)
}
