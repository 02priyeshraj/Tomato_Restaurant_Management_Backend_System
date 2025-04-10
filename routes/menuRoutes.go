package routes

import (
	"net/http"

	controllers "github.com/02priyeshraj/Hotel_Management_Backend/controllers"

	"github.com/gorilla/mux"
)

func MenuProtectedRoutes(router *mux.Router) {

	router.HandleFunc("/menus", controllers.GetMenus).Methods(http.MethodGet)
	router.HandleFunc("/menus", controllers.CreateMenu).Methods(http.MethodPost)

	router.HandleFunc("/menus/{menu_id}", controllers.GetMenu).Methods(http.MethodGet)
	router.HandleFunc("/menus/{menu_id}", controllers.UpdateMenu).Methods(http.MethodPatch)
	router.HandleFunc("/menus/{menu_id}", controllers.DeleteMenu).Methods(http.MethodDelete)
}
