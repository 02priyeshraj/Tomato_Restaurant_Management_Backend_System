package routes

import (
	"net/http"

	controller "github.com/02priyeshraj/Hotel_Management_Backend/controllers"
	"github.com/gorilla/mux"
)

func InvoiceProtectedRoutes(router *mux.Router) {

	router.HandleFunc("/invoices", controller.GetInvoices).Methods(http.MethodGet)
	router.HandleFunc("/invoices", controller.CreateInvoice).Methods(http.MethodPost)

	router.HandleFunc("/invoices/{invoice_id}", controller.GetInvoiceById).Methods(http.MethodGet)
	router.HandleFunc("/invoices/{invoice_id}", controller.UpdateInvoice).Methods(http.MethodPatch)
	router.HandleFunc("/invoices/{invoice_id}", controller.DeleteInvoice).Methods(http.MethodDelete)

	router.HandleFunc("/invoices/order/{order_id}", controller.GetInvoiceByOrderId).Methods(http.MethodGet)
	router.HandleFunc("/invoices/user/{user_id}", controller.GetInvoicesByUserId).Methods(http.MethodGet)

	router.HandleFunc("/invoices/status/pending", controller.GetPendingInvoices).Methods(http.MethodGet)
	router.HandleFunc("/invoices/status/paid", controller.GetPaidInvoices).Methods(http.MethodGet)
}
