package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
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

	// Parse pagination parameters
	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	// MongoDB Aggregation Pipeline
	matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}
	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "table_id", Value: 1},
			{Key: "number_of_guests", Value: 1},
			{Key: "table_number", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}},
	}

	// Execute aggregation
	cursor, err := tableCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving tables"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var allTables []models.Table
	if err = cursor.All(ctx, &allTables); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding table data"}`, http.StatusInternalServerError)
		return
	}

	// Separate valid tables
	var validTables []map[string]interface{}
	for _, table := range allTables {
		if table.Table_number != nil {
			tableData := map[string]interface{}{
				"table_id":         table.Table_id,
				"number_of_guests": table.Number_of_guests,
				"table_number":     table.Table_number,
				"created_at":       table.Created_at,
				"updated_at":       table.Updated_at,
			}
			validTables = append(validTables, tableData)
		}
	}

	// Get total table count
	totalTables, err := tableCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total table count"}`, http.StatusInternalServerError)
		return
	}

	// Construct response
	response := map[string]interface{}{
		"success": true,
		"message": "Tables retrieved successfully",
		"data":    validTables,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_tables":     totalTables,
			"total_pages":      (totalTables + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
			"message": "Invalid request payload",
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

	// Set default status if missing
	if table.Status == "" {
		table.Status = "Not Reserved"
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
			"status":           table.Status,
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

	// table.Status is ignored in this function , it can be updated using ReserveTable and UnreserveTable functions

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

func ReserveTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	tableId := params["table_id"]

	// Check if table exists
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

	// Check if the table is already reserved
	if existingTable.Status == "Reserved" {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table is already reserved",
		})
		return
	}

	// Update the table status to Reserved
	existingTable.Status = "Reserved"
	existingTable.Updated_at = time.Now()
	update := bson.D{{Key: "$set", Value: bson.M{"status": existingTable.Status, "updated_at": existingTable.Updated_at}}}

	_, err = tableCollection.UpdateOne(ctx, bson.M{"table_id": tableId}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to reserve the table",
		})
		return
	}

	// Return updated table details
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Table reserved successfully",
		"data": map[string]interface{}{
			"table_id":         existingTable.Table_id,
			"table_number":     existingTable.Table_number,
			"number_of_guests": existingTable.Number_of_guests,
			"status":           existingTable.Status,
			"created_at":       existingTable.Created_at,
			"updated_at":       existingTable.Updated_at,
		},
	})
}

func UnreserveTable(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	tableId := params["table_id"]

	// Check if table exists
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

	// Check if the table is already not reserved
	if existingTable.Status == "Not Reserved" {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Table is already not reserved",
		})
		return
	}

	// Update the table status to Not Reserved
	existingTable.Status = "Not Reserved"
	existingTable.Updated_at = time.Now()
	update := bson.D{{Key: "$set", Value: bson.M{"status": existingTable.Status, "updated_at": existingTable.Updated_at}}}

	_, err = tableCollection.UpdateOne(ctx, bson.M{"table_id": tableId}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to unreserve the table",
		})
		return
	}

	// Return updated table details
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Table unreserved successfully",
		"data": map[string]interface{}{
			"table_id":         existingTable.Table_id,
			"table_number":     existingTable.Table_number,
			"number_of_guests": existingTable.Number_of_guests,
			"status":           existingTable.Status,
			"created_at":       existingTable.Created_at,
			"updated_at":       existingTable.Updated_at,
		},
	})
}

func GetReservedTables(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Parse pagination query parameters
	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	// Create aggregation pipeline with filtering and pagination
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": "Reserved"}}},
		{{Key: "$skip", Value: startIndex}},
		{{Key: "$limit", Value: int64(recordPerPage)}},
	}

	cursor, err := tableCollection.Aggregate(ctx, pipeline)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving reserved tables"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var reservedTables []models.Table
	if err = cursor.All(ctx, &reservedTables); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding table data"}`, http.StatusInternalServerError)
		return
	}

	// Get total count of reserved tables
	totalCount, err := tableCollection.CountDocuments(ctx, bson.M{"status": "Reserved"})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total reserved table count"}`, http.StatusInternalServerError)
		return
	}

	// Prepare JSON response
	response := map[string]interface{}{
		"success":    true,
		"message":    "Reserved tables retrieved successfully",
		"data":       reservedTables,
		"pagination": map[string]interface{}{"current_page": page, "records_per_page": recordPerPage, "total_reserved_tables": totalCount},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetUnreservedTables(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Parse pagination query parameters
	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	// Create aggregation pipeline with filtering and pagination
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"status": "Not Reserved"}}},
		{{Key: "$skip", Value: startIndex}},
		{{Key: "$limit", Value: int64(recordPerPage)}},
	}

	cursor, err := tableCollection.Aggregate(ctx, pipeline)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving unreserved tables"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var notReservedTables []models.Table
	if err = cursor.All(ctx, &notReservedTables); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding table data"}`, http.StatusInternalServerError)
		return
	}

	// Get total count of unreserved tables
	totalCount, err := tableCollection.CountDocuments(ctx, bson.M{"status": "Not Reserved"})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total unreserved table count"}`, http.StatusInternalServerError)
		return
	}

	// Prepare JSON response
	response := map[string]interface{}{
		"success":    true,
		"message":    "Not reserved tables retrieved successfully",
		"data":       notReservedTables,
		"pagination": map[string]interface{}{"current_page": page, "records_per_page": recordPerPage, "total_unreserved_tables": totalCount},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
