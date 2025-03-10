package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	database "github.com/02priyeshraj/Hotel_Management_Backend/config"
	"github.com/02priyeshraj/Hotel_Management_Backend/models"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

// Get all foods with pagination
func GetFoods(w http.ResponseWriter, r *http.Request) {
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

	totalFoods, err := foodCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total food count"}`, http.StatusInternalServerError)
		return
	}

	matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}
	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "food_id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "price", Value: 1},
			{Key: "food_image", Value: 1},
			{Key: "menu_id", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}},
	}

	result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving food items"}`, http.StatusInternalServerError)
		return
	}

	var allFoods []bson.M
	if err = result.All(ctx, &allFoods); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding food data"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Foods retrieved successfully",
		"data":    allFoods,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_foods":      totalFoods,
			"total_pages":      (totalFoods + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get a single food
func GetFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	foodId := params["food_id"]

	var food models.Food
	if err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food); err != nil {
		http.Error(w, `{"success": false, "message": "Food item not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Food item retrieved successfully",
		"data": map[string]interface{}{
			"food_id":    food.Food_id,
			"name":       food.Name,
			"price":      food.Price,
			"food_image": food.Food_image,
			"menu_id":    food.Menu_id,
			"created_at": food.Created_at,
			"updated_at": food.Updated_at,
		},
	})
}

// Create a food item
func CreateFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var food models.Food
	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if validationErr := validate.Struct(food); validationErr != nil {
		http.Error(w, `{"success": false, "message": "`+validationErr.Error()+`"}`, http.StatusBadRequest)
		return
	}

	menuID, err := primitive.ObjectIDFromHex(*food.Menu_id)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Invalid menu_id format"}`, http.StatusBadRequest)
		return
	}

	uniqueFoodID := *food.Menu_id + "-" + *food.Name
	food.UniqueFoodID = uniqueFoodID

	existingCount, err := foodCollection.CountDocuments(ctx, bson.M{"unique_food_id": uniqueFoodID})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error checking existing food items"}`, http.StatusInternalServerError)
		return
	}
	if existingCount > 0 {
		http.Error(w, `{"success": false, "message": "Food item with the same name already exists in this menu"}`, http.StatusConflict)
		return
	}

	food.ID = primitive.NewObjectID()
	food.Food_id = food.ID.Hex()
	menuIDHex := menuID.Hex()
	food.Menu_id = &menuIDHex
	food.Created_at = time.Now()
	food.Updated_at = time.Now()

	_, insertErr := foodCollection.InsertOne(ctx, food)
	if insertErr != nil {
		http.Error(w, `{"success": false, "message": "Food item could not be created"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Food item created successfully",
		"data": map[string]interface{}{
			"food_id":    food.Food_id,
			"name":       food.Name,
			"price":      food.Price,
			"food_image": food.Food_image,
			"menu_id":    food.Menu_id,
			"created_at": food.Created_at,
			"updated_at": food.Updated_at,
		},
	})
}

// Delete a food item
func DeleteFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	foodId := params["food_id"]

	result, err := foodCollection.DeleteOne(ctx, bson.M{"food_id": foodId})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error deleting food item"}`, http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, `{"success": false, "message": "No food item found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Food item deleted successfully",
	})
}

// Get all foods for a specific menu
func GetFoodsByMenu(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	menuId := params["menu_id"]

	// Validate if menu_id is a valid MongoDB ObjectID
	menuObjID, err := primitive.ObjectIDFromHex(menuId)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Invalid menu ID format"}`, http.StatusBadRequest)
		return
	}

	// Check if menu exists
	menuCount, err := menuCollection.CountDocuments(ctx, bson.M{"_id": menuObjID})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error checking menu existence"}`, http.StatusInternalServerError)
		return
	}
	if menuCount == 0 {
		http.Error(w, `{"success": false, "message": "Menu not found"}`, http.StatusNotFound)
		return
	}

	// Pagination parameters
	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	// Get total food count for the menu
	totalFoods, err := foodCollection.CountDocuments(ctx, bson.M{"menu_id": menuId})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total food count"}`, http.StatusInternalServerError)
		return
	}

	// Fetch paginated food items linked to this menu
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "menu_id", Value: menuId}}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}
	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "food_id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "price", Value: 1},
			{Key: "food_image", Value: 1},
			{Key: "menu_id", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}},
	}

	cursor, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving food items"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var foodItems []bson.M
	if err := cursor.All(ctx, &foodItems); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding food items"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Food items retrieved successfully",
		"data":    foodItems,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_foods":      totalFoods,
			"total_pages":      (totalFoods + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Update a food item
func UpdateFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	foodId := params["food_id"]

	var food models.Food
	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Fetch existing food details
	var existingFood models.Food
	if err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&existingFood); err != nil {
		http.Error(w, `{"success": false, "message": "Food item not found"}`, http.StatusNotFound)
		return
	}

	updateObj := bson.M{"updated_at": time.Now()}

	// If name is being updated, check for duplicates
	if food.Name != nil && *food.Name != *existingFood.Name {
		newUniqueFoodID := *existingFood.Menu_id + "-" + *food.Name

		duplicateCount, err := foodCollection.CountDocuments(ctx, bson.M{"unique_food_id": newUniqueFoodID})
		if err != nil {
			http.Error(w, `{"success": false, "message": "Error checking duplicate food items"}`, http.StatusInternalServerError)
			return
		}
		if duplicateCount > 0 {
			http.Error(w, `{"success": false, "message": "Another food item with the same name exists in this menu"}`, http.StatusConflict)
			return
		}

		updateObj["name"] = food.Name
		updateObj["unique_food_id"] = newUniqueFoodID // Update unique identifier
	}

	if food.Price != nil {
		updateObj["price"] = food.Price
	}
	if food.Food_image != nil {
		updateObj["food_image"] = food.Food_image
	}
	if food.Menu_id != nil {
		updateObj["menu_id"] = food.Menu_id
	}

	filter := bson.M{"food_id": foodId}
	update := bson.M{"$set": updateObj}

	_, err := foodCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Food item update failed"}`, http.StatusInternalServerError)
		return
	}

	// Fetch the updated food item
	var updatedFood models.Food
	if err := foodCollection.FindOne(ctx, filter).Decode(&updatedFood); err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving updated food item"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Food item updated successfully",
		"data":    updatedFood,
	})
}
