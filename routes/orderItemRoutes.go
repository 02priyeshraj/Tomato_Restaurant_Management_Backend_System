package routes

import (
	"net/http"

	controller "github.com/02priyeshraj/Hotel_Management_Backend/controllers"
	"github.com/gorilla/mux"
)

func OrderItemProtectedRoutes(router *mux.Router) {
	router.HandleFunc("/orderitems", controller.CreateOrderItem).Methods(http.MethodPost)
	router.HandleFunc("/orderitems", controller.GetOrderItems).Methods(http.MethodGet)
	router.HandleFunc("/orderitems/{order_item_id}", controller.GetOrderItemById).Methods(http.MethodGet)
	router.HandleFunc("/orderitems/{order_item_id}", controller.UpdateOrderItem).Methods(http.MethodPatch)
	router.HandleFunc("/orderitems/{order_item_id}", controller.DeleteOrderItem).Methods(http.MethodDelete)
	router.HandleFunc("/orderitems/{order_id}/order", controller.GetOrderItemsByOrderId).Methods(http.MethodGet)

}
