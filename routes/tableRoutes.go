package routes

import (
	controller "github.com/02priyeshraj/Hotel_Management_Backend/controllers"
	"github.com/gorilla/mux"
)

func TableProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/tables", controller.GetTables).Methods("GET")
	router.HandleFunc("/tables/{table_id}", controller.GetTable).Methods("GET")
	router.HandleFunc("/tables", controller.CreateTable).Methods("POST")
	router.HandleFunc("/tables/{table_id}", controller.UpdateTable).Methods("PATCH")
	router.HandleFunc("/tables/{table_id}", controller.DeleteTable).Methods("DELETE")
}
