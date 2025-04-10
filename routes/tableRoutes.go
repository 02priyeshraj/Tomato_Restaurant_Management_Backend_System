package routes

import (
	"net/http"

	controller "github.com/02priyeshraj/Hotel_Management_Backend/controllers"
	"github.com/gorilla/mux"
)

func TableProtectedRoutes(router *mux.Router) {

	router.HandleFunc("/tables", controller.GetTables).Methods(http.MethodGet)
	router.HandleFunc("/tables", controller.CreateTable).Methods(http.MethodPost)

	router.HandleFunc("/tables/reserved", controller.GetReservedTables).Methods(http.MethodGet)
	router.HandleFunc("/tables/unreserved", controller.GetUnreservedTables).Methods(http.MethodGet)

	router.HandleFunc("/tables/{table_id}", controller.GetTable).Methods(http.MethodGet)
	router.HandleFunc("/tables/{table_id}", controller.UpdateTable).Methods(http.MethodPatch)
	router.HandleFunc("/tables/{table_id}", controller.DeleteTable).Methods(http.MethodDelete)

	router.HandleFunc("/tables/reserve/{table_id}", controller.ReserveTable).Methods(http.MethodPut)
	router.HandleFunc("/tables/unreserve/{table_id}", controller.UnreserveTable).Methods(http.MethodPut)
}
