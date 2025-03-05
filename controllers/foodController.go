package controller

import (
	"context"
	"encoding/json"

	"log"
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

	matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
	groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "null"}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "total_count", Value: 1},
			{Key: "food_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
		}},
	}

	result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
	if err != nil {
		http.Error(w, "Error occurred while listing food items", http.StatusInternalServerError)
		return
	}

	var allFoods []bson.M
	if err = result.All(ctx, &allFoods); err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allFoods[0])
}

func GetFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	foodId := params["food_id"]

	var food models.Food
	if err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food); err != nil {
		http.Error(w, "Error occurred while fetching the food item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(food)
}

func CreateFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var food models.Food
	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validationErr := validate.Struct(food)
	if validationErr != nil {
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	food.ID = primitive.NewObjectID()
	food.Food_id = food.ID.Hex()
	food.Created_at = time.Now()
	food.Updated_at = time.Now()

	_, insertErr := foodCollection.InsertOne(ctx, food)
	if insertErr != nil {
		http.Error(w, "Food item was not created", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(food)
}

func DeleteFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	foodId := params["food_id"]

	result, err := foodCollection.DeleteOne(ctx, bson.M{"food_id": foodId})
	if err != nil {
		http.Error(w, "Error deleting food item", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "No food item found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Food item deleted successfully"})
}

func UpdateFood(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	foodId := params["food_id"]

	var food models.Food
	if err := json.NewDecoder(r.Body).Decode(&food); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updateObj := bson.M{"updated_at": time.Now()}
	if food.Name != nil {
		updateObj["name"] = food.Name
	}
	if food.Price != nil {
		updateObj["price"] = food.Price
	}
	if food.Food_image != nil {
		updateObj["food_image"] = food.Food_image
	}

	filter := bson.M{"food_id": foodId}
	update := bson.M{"$set": updateObj}

	_, err := foodCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Food item update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Food item updated successfully"})
}
