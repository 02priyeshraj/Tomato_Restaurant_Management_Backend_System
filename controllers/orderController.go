package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	database "github.com/02priyeshraj/Hotel_Management_Backend/config"
	"github.com/02priyeshraj/Hotel_Management_Backend/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

// Get all orders
func GetOrders(w http.ResponseWriter, r *http.Request) {
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
			{Key: "order_id", Value: 1},
			{Key: "order_date", Value: 1},
			{Key: "table_id", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "status", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}},
	}

	// Execute aggregation
	cursor, err := orderCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, `{"error": "Error retrieving orders"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var allOrders []bson.M
	if err = cursor.All(ctx, &allOrders); err != nil {
		http.Error(w, `{"error": "Error decoding order data"}`, http.StatusInternalServerError)
		return
	}

	// Get total order count
	totalOrders, err := orderCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		http.Error(w, `{"error": "Error retrieving total order count"}`, http.StatusInternalServerError)
		return
	}

	// Construct response
	response := map[string]interface{}{
		"success": true,
		"message": "Orders retrieved successfully",
		"data":    allOrders,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_orders":     totalOrders,
			"total_pages":      (totalOrders + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetOrderById(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderId := mux.Vars(r)["order_id"]
	if orderId == "" {
		http.Error(w, `{"success": false, "message": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	var order models.Order
	err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, `{"success": false, "message": "Order not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving order"}`, http.StatusInternalServerError)
		return
	}

	// Construct response
	response := map[string]interface{}{
		"success": true,
		"message": "Order retrieved successfully",
		"data": map[string]interface{}{
			"order_id":   order.Order_id,
			"user_id":    order.User_id,
			"table_id":   order.Table_id,
			"status":     order.Status,
			"order_date": order.Order_Date,
			"created_at": order.Created_at,
			"updated_at": order.Updated_at,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetOrdersByTableId(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	tableId := mux.Vars(r)["table_id"]
	if tableId == "" {
		http.Error(w, `{"success": false, "message": "Invalid table ID"}`, http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10 // Default records per page
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1 // Default to first page
	}

	startIndex := (page - 1) * recordPerPage

	// MongoDB aggregation pipeline for pagination
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "table_id", Value: tableId}}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}

	cursor, err := orderCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving orders"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding orders data"}`, http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		http.Error(w, `{"success": false, "message": "No orders found for this table"}`, http.StatusNotFound)
		return
	}

	// Get total order count for the given table_id
	totalOrders, err := orderCollection.CountDocuments(ctx, bson.M{"table_id": tableId})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total order count"}`, http.StatusInternalServerError)
		return
	}

	// Construct response
	response := map[string]interface{}{
		"success": true,
		"message": "Orders retrieved successfully",
		"data":    orders,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_orders":     totalOrders,
			"total_pages":      (totalOrders + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetOrdersByUserId(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	userId := mux.Vars(r)["user_id"]
	if userId == "" {
		http.Error(w, `{"success": false, "message": "Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	// MongoDB aggregation pipeline for pagination
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "user_id", Value: userId}}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}

	cursor, err := orderCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving orders"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding orders data"}`, http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		http.Error(w, `{"success": false, "message": "No orders found for this user"}`, http.StatusNotFound)
		return
	}

	// Get total order count for the given user_id
	totalOrders, err := orderCollection.CountDocuments(ctx, bson.M{"user_id": userId})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total order count"}`, http.StatusInternalServerError)
		return
	}

	// Construct response
	response := map[string]interface{}{
		"success": true,
		"message": "Orders retrieved successfully",
		"data":    orders,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_orders":     totalOrders,
			"total_pages":      (totalOrders + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// **Ensure default order status if missing or empty**
	if order.Status == "" {
		order.Status = "Order Pending"
	}

	// Validate Order Data
	if validationErr := validate.StructPartial(order, "Order_Date", "Table_id", "User_id"); validationErr != nil {
		http.Error(w, `{"success": false, "message": "%s"}`, http.StatusBadRequest)
		return
	}

	// Validate Table ID and check if the table is reserved
	var table models.Table
	err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Invalid table ID, table not found"}`, http.StatusNotFound)
		return
	}

	if table.Status != "Reserved" {
		http.Error(w, `{"success": false, "message": "Table is not reserved. Reserve the table first."}`, http.StatusBadRequest)
		return
	}

	// Validate User ID exists
	count, err := userCollection.CountDocuments(ctx, bson.M{"user_id": order.User_id})
	if err != nil || count == 0 {
		http.Error(w, `{"success": false, "message": "Invalid user ID, user not found"}`, http.StatusNotFound)
		return
	}

	// Set timestamps and unique Order ID
	order.Created_at = time.Now()
	order.Updated_at = time.Now()
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	_, err = orderCollection.InsertOne(ctx, order)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Order creation failed"}`, http.StatusInternalServerError)
		return
	}

	// Construct Success Response
	response := map[string]interface{}{
		"success": true,
		"message": "Order created successfully",
		"data":    order,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderId := mux.Vars(r)["order_id"]
	var order models.Order

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	updateObj := bson.D{}

	// Validate Table ID before updating
	if order.Table_id != nil {
		// Check if the new table exists
		var table models.Table
		err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Invalid table ID, table not found"}`, http.StatusNotFound)
			return
		}

		// Check if the table is reserved
		if table.Status != "Reserved" {
			http.Error(w, `{"success": false, "message": "Table is not reserved. Reserve the table first."}`, http.StatusBadRequest)
			return
		}

		// Check if the table is already assigned to another order
		existingOrderCount, err := orderCollection.CountDocuments(ctx, bson.M{"table_id": order.Table_id, "order_id": bson.M{"$ne": orderId}})
		if err != nil {
			http.Error(w, `{"success": false, "message": "Error checking table availability"}`, http.StatusInternalServerError)
			return
		}

		if existingOrderCount > 0 {
			http.Error(w, `{"success": false, "message": "Table is already assigned to another order."}`, http.StatusBadRequest)
			return
		}

		updateObj = append(updateObj, bson.E{Key: "table_id", Value: order.Table_id})
	}

	// order.Status is ignored here, use UpdateOrderStatus endpoint to update status

	// Update order timestamp
	order.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: order.Updated_at})

	if len(updateObj) == 0 {
		http.Error(w, `{"success": false, "message": "No fields to update"}`, http.StatusBadRequest)
		return
	}

	filter := bson.M{"order_id": orderId}
	opt := options.Update().SetUpsert(false)

	result, err := orderCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, opt)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Order update failed"}`, http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, `{"success": false, "message": "Order not found"}`, http.StatusNotFound)
		return
	}

	// Fetch the updated order
	var updatedOrder models.Order
	err = orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&updatedOrder)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving updated order"}`, http.StatusInternalServerError)
		return
	}

	// Construct Success Response
	response := map[string]interface{}{
		"success": true,
		"message": "Order updated successfully",
		"data":    updatedOrder,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderId := mux.Vars(r)["order_id"]

	// Find the order before deleting
	var order models.Order
	err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, `{"success": false, "message": "Order not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving order"}`, http.StatusInternalServerError)
		return
	}

	// Delete the order
	result, err := orderCollection.DeleteOne(ctx, bson.M{"order_id": orderId})
	if err != nil || result.DeletedCount == 0 {
		http.Error(w, `{"success": false, "message": "Error deleting order"}`, http.StatusInternalServerError)
		return
	}

	// Construct Success Response
	response := map[string]interface{}{
		"success": true,
		"message": "Order deleted successfully",
		"data":    order,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderId := mux.Vars(r)["order_id"]

	// Parse request body
	var requestBody struct {
		Status string `json:"status" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate Status
	validStatuses := map[string]bool{
		"Order Confirmed": true, "Preparing Order": true, "Order Served": true,
		"Order Paid": true, "Order Cancelled": true, "Order Rejected": true, "Order Pending": true, "Order Placed": true,
	}

	if !validStatuses[requestBody.Status] {
		http.Error(w, `{"success": false, "message": "Invalid order status"}`, http.StatusBadRequest)
		return
	}

	// Check if order exists
	var order models.Order
	err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, `{"success": false, "message": "Order not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving order"}`, http.StatusInternalServerError)
		return
	}

	// Update status and timestamp
	update := bson.M{
		"$set": bson.M{
			"status":     requestBody.Status,
			"updated_at": time.Now(),
		},
	}

	result, err := orderCollection.UpdateOne(ctx, bson.M{"order_id": orderId}, update)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Failed to update order status"}`, http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, `{"success": false, "message": "Order not found"}`, http.StatusNotFound)
		return
	}

	// Fetch updated order
	err = orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving updated order"}`, http.StatusInternalServerError)
		return
	}

	// Construct Success Response
	response := map[string]interface{}{
		"success": true,
		"message": "Order status updated successfully",
		"data":    order,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
