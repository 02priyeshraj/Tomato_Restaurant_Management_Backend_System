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

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderitems")

// Function to fetch food details and replace food_id with food_name
func replaceFoodIdsWithNames(ctx context.Context, items map[string]int) (map[string]int, error) {
	transformedItems := make(map[string]int)
	var missingFoodIDs []string

	for foodID, quantity := range items {
		var food models.Food
		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodID}).Decode(&food)
		if err != nil {
			missingFoodIDs = append(missingFoodIDs, foodID)
			continue
		}
		transformedItems[*food.Name] = quantity
	}

	if len(missingFoodIDs) > 0 {
		return nil, errors.New("food items not found: " + strings.Join(missingFoodIDs, ", "))
	}

	return transformedItems, nil
}

// GetOrderItems retrieves all order items with food names
// GetOrderItems retrieves all order items (no need to replace IDs with names)
func GetOrderItems(w http.ResponseWriter, r *http.Request) {
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

	cursor, err := orderItemCollection.Find(ctx, bson.M{}, options.Find().SetSkip(int64(startIndex)).SetLimit(int64(recordPerPage)))
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving order items"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var orderItems []models.OrderItem
	if err = cursor.All(ctx, &orderItems); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding order items"}`, http.StatusInternalServerError)
		return
	}

	totalCount, _ := orderItemCollection.CountDocuments(ctx, bson.M{})

	response := map[string]interface{}{
		"success": true,
		"message": "Order items retrieved successfully",
		"data":    orderItems, // No need to replace IDs
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_orderitems": totalCount,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrderItemById retrieves a single order item with food names
// GetOrderItemById retrieves a single order item (no need to replace IDs with names)
func GetOrderItemById(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderItemId := mux.Vars(r)["order_item_id"]

	var orderItem models.OrderItem
	err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Order item not found"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Order item retrieved successfully",
		"data":    orderItem, // No need to transform
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateOrderItem creates a new order item
func CreateOrderItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var orderItem models.OrderItem
	if err := json.NewDecoder(r.Body).Decode(&orderItem); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate order existence and status
	var order models.Order
	err := orderCollection.FindOne(ctx, bson.M{"order_id": orderItem.Order_id}).Decode(&order)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	// Validate that the provided table_id matches the order's table_id
	if orderItem.Table_id == "" || orderItem.Table_id != *order.Table_id {
		http.Error(w, `{"success": false, "message": "Invalid table ID for this order"}`, http.StatusBadRequest)
		return
	}

	// If order status is "Order Pending", update it to "Order Placed"
	if order.Status == "Order Pending" {
		_, err := orderCollection.UpdateOne(ctx,
			bson.M{"order_id": orderItem.Order_id},
			bson.M{"$set": bson.M{"status": "Order Placed", "updated_at": time.Now()}},
		)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Failed to update order status"}`, http.StatusInternalServerError)
			return
		}
	}

	// Initialize total price calculation
	var totalPrice float64
	transformedItems := make(map[string]int)
	var missingFoodIDs []string

	for foodID, quantity := range orderItem.Items {
		var food models.Food
		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodID}).Decode(&food)
		if err != nil {
			missingFoodIDs = append(missingFoodIDs, foodID)
			continue
		}

		transformedItems[*food.Name] = quantity
		totalPrice += (*food.Price) * float64(quantity)
	}

	if len(missingFoodIDs) > 0 {
		http.Error(w, `{"success": false, "message": "Food items not found: `+strings.Join(missingFoodIDs, ", ")+`"}`, http.StatusBadRequest)
		return
	}

	// Assign calculated total price and transformed items
	orderItem.Items = transformedItems
	orderItem.TotalPrice = totalPrice
	orderItem.Created_at = time.Now()
	orderItem.Updated_at = time.Now()
	orderItem.ID = primitive.NewObjectID()
	orderItem.Order_item_id = orderItem.ID.Hex()

	// Insert the new order item
	_, err = orderItemCollection.InsertOne(ctx, orderItem)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Order item creation failed"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Order item created successfully",
		"data":    orderItem,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateOrderItem updates an existing order item
func UpdateOrderItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderItemId := mux.Vars(r)["order_item_id"]
	var updateRequest models.OrderItem

	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Fetch the existing order item
	var existingOrderItem models.OrderItem
	err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Decode(&existingOrderItem)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Order item not found"}`, http.StatusNotFound)
		return
	}

	// Convert food IDs to food names and check for missing IDs
	foodIDToName := make(map[string]string)
	var missingFoodIDs []string

	for foodID := range updateRequest.Items {
		var food models.Food
		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodID}).Decode(&food)
		if err != nil {
			missingFoodIDs = append(missingFoodIDs, foodID)
			continue
		}
		foodIDToName[foodID] = *food.Name
	}

	// If any food IDs are missing, return an error
	if len(missingFoodIDs) > 0 {
		http.Error(w, `{"success": false, "message": "Food items not found for IDs: `+strings.Join(missingFoodIDs, ", ")+`"}`, http.StatusBadRequest)
		return
	}

	// Update only the quantities of the items present in the request
	for foodID, newQuantity := range updateRequest.Items {
		foodName := foodIDToName[foodID] // Since all IDs are verified, this will always exist
		if _, itemExists := existingOrderItem.Items[foodName]; itemExists {
			existingOrderItem.Items[foodName] = newQuantity
		}
	}

	// Update the order item in the database
	updateObj := bson.D{
		{Key: "items", Value: existingOrderItem.Items},
		{Key: "updated_at", Value: time.Now()},
	}

	filter := bson.M{"order_item_id": orderItemId}
	opt := options.Update().SetUpsert(false)

	_, err = orderItemCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, opt)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Order item update failed"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"message":       "Order item updated successfully",
		"updated_items": updateRequest.Items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteOrderItem deletes an order item.
func DeleteOrderItem(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderItemId := mux.Vars(r)["order_item_id"]

	// Check if the order item exists before deleting
	err := orderItemCollection.FindOne(ctx, bson.M{"order_item_id": orderItemId}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, `{"success": false, "message": "Order item not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving order item"}`, http.StatusInternalServerError)
		return
	}

	// Proceed with deletion
	result, err := orderItemCollection.DeleteOne(ctx, bson.M{"order_item_id": orderItemId})
	if err != nil || result.DeletedCount == 0 {
		http.Error(w, `{"success": false, "message": "Order item deletion failed"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Order item deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrderItemsByOrderId retrieves all order items for a given order_id with food names
func GetOrderItemsByOrderId(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderId := mux.Vars(r)["order_id"]
	if orderId == "" {
		http.Error(w, `{"success": false, "message": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	// Find order items by order_id
	cursor, err := orderItemCollection.Find(ctx, bson.M{"order_id": orderId})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving order items"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var orderItems []models.OrderItem
	if err = cursor.All(ctx, &orderItems); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding order items"}`, http.StatusInternalServerError)
		return
	}

	if len(orderItems) == 0 {
		http.Error(w, `{"success": false, "message": "No order items found for this order ID"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Order items retrieved successfully",
		"data":    orderItems,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
