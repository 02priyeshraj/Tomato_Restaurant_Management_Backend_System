package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	database "github.com/02priyeshraj/Hotel_Management_Backend/config"
	"github.com/02priyeshraj/Hotel_Management_Backend/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

// Get all menus
func GetMenus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := menuCollection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Error occurred while listing the menu items",
		})
		return
	}

	var allMenus []bson.M
	if err = result.All(ctx, &allMenus); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menus retrieved successfully",
		"data":    allMenus,
	})
}

// Get a single menu
func GetMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	menuId := params["menu_id"]

	var menu models.Menu
	err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Menu not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu retrieved successfully",
		"data":    menu,
	})
}

// Create a menu
func CreateMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu
	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	menu.Created_at = time.Now()
	menu.Updated_at = time.Now()
	menu.ID = primitive.NewObjectID()
	menu.Menu_id = menu.ID.Hex()

	result, insertErr := menuCollection.InsertOne(ctx, menu)
	if insertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Menu item was not created",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu created successfully",
		"data":    result,
	})
}

// Update a menu
func UpdateMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	menuId := params["menu_id"]

	var menu models.Menu
	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	filter := bson.M{"menu_id": menuId}
	updateObj := bson.D{}

	if menu.Start_Date != nil && menu.End_Date != nil {
		if menu.Start_Date.After(*menu.End_Date) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Invalid date range: start_date must be before end_date",
			})
			return
		}
		updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_Date})
		updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_Date})
	}

	if menu.Name != "" {
		updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
	}
	if menu.Category != "" {
		updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
	}

	menu.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

	upsert := true
	opt := options.Update().SetUpsert(upsert)

	result, err := menuCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, opt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Menu update failed",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu updated successfully",
		"data":    result,
	})
}

// Delete a menu
func DeleteMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	menuId := params["menu_id"]

	result, err := menuCollection.DeleteOne(ctx, bson.M{"menu_id": menuId})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Menu deletion failed",
		})
		return
	}

	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Menu not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu deleted successfully",
	})
}
