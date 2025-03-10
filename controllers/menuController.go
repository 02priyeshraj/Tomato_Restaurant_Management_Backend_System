package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
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

// Get all menus with pagination
func GetMenus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}
	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "menu_id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "category", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}},
	}

	result, err := menuCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, "Error retrieving menus", http.StatusInternalServerError)
		return
	}

	var allMenus []bson.M
	if err = result.All(ctx, &allMenus); err != nil {
		http.Error(w, "Error decoding menu data", http.StatusInternalServerError)
		return
	}

	totalMenus, err := menuCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Error retrieving total menu count", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Menus retrieved successfully",
		"data":    allMenus,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_menus":      totalMenus,
			"total_pages":      (totalMenus + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get a single menu
func GetMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	menuId := mux.Vars(r)["menu_id"]
	if menuId == "" {
		http.Error(w, "Invalid menu ID", http.StatusBadRequest)
		return
	}

	var menu models.Menu
	err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving menu", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu retrieved successfully",
		"data": map[string]interface{}{
			"menu_id":    menu.Menu_id,
			"name":       menu.Name,
			"category":   menu.Category,
			"created_at": menu.Created_at,
			"updated_at": menu.Updated_at,
		},
	})
}

// Create a menu
func CreateMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu
	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate UniqueID (lowercase version of name)
	menu.UniqueID = strings.ToLower(menu.Name)

	// Check if a menu with the same UniqueID already exists
	count, err := menuCollection.CountDocuments(ctx, bson.M{"unique_id": menu.UniqueID})
	if err != nil {
		http.Error(w, "Error checking menu existence", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Menu with this name already exists", http.StatusConflict)
		return
	}

	// Set timestamps and unique menu ID
	menu.Created_at = time.Now()
	menu.Updated_at = time.Now()
	menu.ID = primitive.NewObjectID()
	menu.Menu_id = menu.ID.Hex()

	_, err = menuCollection.InsertOne(ctx, menu)
	if err != nil {
		http.Error(w, "Error creating menu", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu created successfully",
		"data": map[string]interface{}{
			"menu_id":    menu.Menu_id,
			"name":       menu.Name,
			"category":   menu.Category,
			"created_at": menu.Created_at,
			"updated_at": menu.Updated_at,
		},
	})
}

// Update a menu
func UpdateMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	menuId := mux.Vars(r)["menu_id"]
	var menu models.Menu
	if err := json.NewDecoder(r.Body).Decode(&menu); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate UniqueID (lowercase version of name)
	newUniqueID := strings.ToLower(menu.Name)

	// Check if a menu with the same UniqueID already exists (excluding current menu)
	count, err := menuCollection.CountDocuments(ctx, bson.M{"unique_id": newUniqueID, "menu_id": bson.M{"$ne": menuId}})
	if err != nil {
		http.Error(w, "Error checking menu existence", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Another menu with this name already exists", http.StatusConflict)
		return
	}

	updateObj := bson.D{}
	if menu.Name != "" {
		updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
		updateObj = append(updateObj, bson.E{Key: "unique_id", Value: newUniqueID})
	}
	if menu.Category != "" {
		updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
	}

	menu.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

	opt := options.Update().SetUpsert(false)
	result, err := menuCollection.UpdateOne(ctx, bson.M{"menu_id": menuId}, bson.D{{Key: "$set", Value: updateObj}}, opt)
	if err != nil {
		http.Error(w, "Error updating menu", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}

	// Fetch the updated menu
	var updatedMenu models.Menu
	err = menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&updatedMenu)
	if err != nil {
		http.Error(w, "Error retrieving updated menu", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu updated successfully",
		"data": map[string]interface{}{
			"menu_id":    updatedMenu.Menu_id,
			"name":       updatedMenu.Name,
			"category":   updatedMenu.Category,
			"created_at": updatedMenu.Created_at,
			"updated_at": updatedMenu.Updated_at,
		},
	})
}

// Delete a menu
func DeleteMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	menuId := mux.Vars(r)["menu_id"]

	// Find the menu before deleting
	var menu models.Menu
	err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving menu", http.StatusInternalServerError)
		return
	}

	// Delete the menu
	result, err := menuCollection.DeleteOne(ctx, bson.M{"menu_id": menuId})
	if err != nil {
		http.Error(w, "Error deleting menu", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Menu not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Menu deleted successfully",
		"data": map[string]interface{}{
			"menu_id":    menu.Menu_id,
			"name":       menu.Name,
			"category":   menu.Category,
			"created_at": menu.Created_at,
			"updated_at": menu.Updated_at,
		},
	})
}
