package routes

import (
	"net/http"

	controller "github.com/02priyeshraj/Hotel_Management_Backend/controllers"

	"github.com/gorilla/mux"
)

func OrderProtectedRoutes(router *mux.Router) {

	router.HandleFunc("/orders", controller.GetOrders).Methods(http.MethodGet)
	router.HandleFunc("/orders", controller.CreateOrder).Methods(http.MethodPost)

	router.HandleFunc("/orders/{order_id}", controller.GetOrderById).Methods(http.MethodGet)
	router.HandleFunc("/orders/{order_id}", controller.UpdateOrder).Methods(http.MethodPatch)
	router.HandleFunc("/orders/{order_id}", controller.DeleteOrder).Methods(http.MethodDelete)
	router.HandleFunc("/orders/{order_id}/status", controller.UpdateOrderStatus).Methods(http.MethodPatch)

	router.HandleFunc("/orders/table/{table_id}", controller.GetOrdersByTableId).Methods(http.MethodGet)
	router.HandleFunc("/orders/user/{user_id}", controller.GetOrdersByUserId).Methods(http.MethodGet)
}
