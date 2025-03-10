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
)

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := tableCollection.Find(ctx, bson.M{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Error occurred while listing table items",
		})
		return
	}

	var allTables []models.Table
	if err = result.All(ctx, &allTables); err != nil {
		log.Fatal(err)
	}

	// Separate valid and invalid tables
	var validTables []map[string]interface{}

	for _, table := range allTables {
		tableData := map[string]interface{}{
			"table_id":         table.Table_id,
			"number_of_guests": table.Number_of_guests,
			"created_at":       table.Created_at,
			"updated_at":       table.Updated_at,
		}

		if table.Table_number != nil {
			tableData["table_number"] = table.Table_number
			validTables = append(validTables, tableData)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Tables retrieved successfully",
		"valid_tables": validTables,
	})
}

func GetTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	tableId := params["table_id"]
	var table models.Table

	err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table not found",
		})
		return
	}

	// If table_number is nil, return "Table does not exist"
	if table.Table_number == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table does not exist",
		})
		return
	}

	// Construct response while excluding table_number if nil
	responseData := map[string]interface{}{
		"table_id":         table.Table_id,
		"number_of_guests": table.Number_of_guests,
		"created_at":       table.Created_at,
		"updated_at":       table.Updated_at,
	}

	if table.Table_number != nil {
		responseData["table_number"] = table.Table_number
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Table retrieved successfully",
		"data":    responseData,
	})
}

func CreateTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var table models.Table
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Check if the table number already exists
	count, err := tableCollection.CountDocuments(ctx, bson.M{"table_number": table.Table_number})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Error checking table number",
		})
		return
	}
	if count > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table number already exists",
		})
		return
	}

	// Set metadata fields
	table.Created_at = time.Now()
	table.Updated_at = time.Now()
	table.ID = primitive.NewObjectID()
	table.Table_id = table.ID.Hex()

	// Insert into MongoDB
	_, insertErr := tableCollection.InsertOne(ctx, table)
	if insertErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table item was not created",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Table created successfully",
		"data": map[string]interface{}{
			"table_id":         table.Table_id,
			"number_of_guests": table.Number_of_guests,
			"table_number":     table.Table_number,
			"created_at":       table.Created_at,
			"updated_at":       table.Updated_at,
		},
	})
}

func UpdateTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	tableId := params["table_id"]
	var table models.Table

	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request payload",
		})
		return
	}

	// Fetch the existing table
	var existingTable models.Table
	err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&existingTable)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table not found",
		})
		return
	}

	// If table_number is nil, return "Table does not exist"
	if existingTable.Table_number == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table does not exist",
		})
		return
	}

	// Prepare the update object
	updateObj := bson.D{}
	if table.Number_of_guests != nil {
		updateObj = append(updateObj, bson.E{Key: "number_of_guests", Value: table.Number_of_guests})
	}
	if table.Table_number != nil {
		updateObj = append(updateObj, bson.E{Key: "table_number", Value: table.Table_number})
	}
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: time.Now()})

	filter := bson.M{"table_id": tableId}
	update := bson.D{{Key: "$set", Value: updateObj}}

	_, err = tableCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to update table",
		})
		return
	}

	// Fetch updated table data
	var updatedTable models.Table
	err = tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&updatedTable)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Error fetching updated table",
		})
		return
	}

	// Construct response
	responseData := map[string]interface{}{
		"table_id":         updatedTable.Table_id,
		"number_of_guests": updatedTable.Number_of_guests,
		"created_at":       updatedTable.Created_at,
		"updated_at":       updatedTable.Updated_at,
	}

	if updatedTable.Table_number != nil {
		responseData["table_number"] = updatedTable.Table_number
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Table updated successfully",
		"data":    responseData,
	})
}

func DeleteTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	tableId := params["table_id"]

	// Fetch the existing table
	var existingTable models.Table
	err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&existingTable)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table not found",
		})
		return
	}

	// If table_number is nil, return "Table does not exist"
	if existingTable.Table_number == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table does not exist",
		})
		return
	}

	// Delete the document from MongoDB
	result, err := tableCollection.DeleteOne(ctx, bson.M{"table_id": tableId})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table item deletion failed",
		})
		return
	}

	// Successful deletion response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Table deleted successfully",
		"data":    result,
	})
}
