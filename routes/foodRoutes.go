package routes

import (
	"net/http"

	controllers "github.com/02priyeshraj/Hotel_Management_Backend/controllers"
	"github.com/gorilla/mux"
)

func FoodProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/foods", controllers.GetFoods).Methods(http.MethodGet)
	router.HandleFunc("/foods/{food_id}", controllers.GetFood).Methods(http.MethodGet)
	router.HandleFunc("/foods", controllers.CreateFood).Methods(http.MethodPost)
	router.HandleFunc("/foods/{food_id}", controllers.UpdateFood).Methods(http.MethodPatch)
	router.HandleFunc("/foods/{food_id}", controllers.DeleteFood).Methods(http.MethodDelete)
}
