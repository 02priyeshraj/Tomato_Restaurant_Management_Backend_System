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

// Open the Invoice collection
var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

// GetInvoices retrieves all invoices with pagination.
func GetInvoices(w http.ResponseWriter, r *http.Request) {
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
	skip := (page - 1) * recordPerPage

	// Build aggregation pipeline
	matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
	skipStage := bson.D{{Key: "$skip", Value: int64(skip)}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}
	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "invoice_id", Value: 1},
		{Key: "order_id", Value: 1},
		{Key: "user_id", Value: 1},
		{Key: "payment_method", Value: 1},
		{Key: "payment_status", Value: 1},
		{Key: "total_price", Value: 1},
		{Key: "payment_date", Value: 1},
		{Key: "created_at", Value: 1},
		{Key: "updated_at", Value: 1},
	}}}

	cursor, err := invoiceCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving invoices"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var invoices []bson.M
	if err = cursor.All(ctx, &invoices); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding invoice data"}`, http.StatusInternalServerError)
		return
	}

	totalCount, err := invoiceCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total invoice count"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Invoices retrieved successfully",
		"data":    invoices,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_invoices":   totalCount,
			"total_pages":      (totalCount + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetInvoiceById retrieves a single invoice by its invoice_id.
func GetInvoiceById(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	invoiceId := mux.Vars(r)["invoice_id"]
	if invoiceId == "" {
		http.Error(w, `{"success": false, "message": "Invalid invoice ID"}`, http.StatusBadRequest)
		return
	}

	var invoice models.Invoice
	err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, `{"success": false, "message": "Invoice not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving invoice"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Invoice retrieved successfully",
		"data": map[string]interface{}{
			"invoice_id":     invoice.Invoice_id,
			"order_id":       invoice.Order_id,
			"user_id":        invoice.User_id,
			"payment_method": invoice.Payment_method,
			"payment_status": invoice.Payment_status,
			"total_price":    invoice.TotalPrice,
			"payment_date":   invoice.Payment_date,
			"created_at":     invoice.Created_at,
			"updated_at":     invoice.Updated_at,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateInvoice creates a new invoice.
func CreateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var invoice models.Invoice
	if err := json.NewDecoder(r.Body).Decode(&invoice); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Set default Payment_status if missing (assume default is "PENDING")
	if invoice.Payment_status == nil || *invoice.Payment_status == "" {
		defaultStatus := "PENDING"
		invoice.Payment_status = &defaultStatus
	}

	// Calculate total price from all order items for the given order_id
	if invoice.Order_id == nil || *invoice.Order_id == "" {
		http.Error(w, `{"success": false, "message": "Order ID is required in invoice"}`, http.StatusBadRequest)
		return
	}

	// Query all order items with the given order_id
	cursor, err := orderItemCollection.Find(ctx, bson.M{"order_id": *invoice.Order_id})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving order items"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var orderItems []models.OrderItem
	if err := cursor.All(ctx, &orderItems); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding order items"}`, http.StatusInternalServerError)
		return
	}

	// Sum up the total prices from order items
	var calculatedTotal float64 = 0.0
	for _, item := range orderItems {
		calculatedTotal += item.TotalPrice
	}
	// Assign calculated total price to invoice.TotalPrice field
	invoice.TotalPrice = calculatedTotal

	// Set timestamps and unique Invoice ID
	invoice.Created_at = time.Now()
	invoice.Updated_at = time.Now()
	invoice.ID = primitive.NewObjectID()
	invoice.Invoice_id = invoice.ID.Hex()

	// Insert the invoice into the database
	_, err = invoiceCollection.InsertOne(ctx, invoice)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Invoice creation failed"}`, http.StatusInternalServerError)
		return
	}

	// If the payment status is PAID, update the related order status to "Order Paid"
	if strings.EqualFold(*invoice.Payment_status, "PAID") {
		update := bson.M{
			"$set": bson.M{
				"status":     "Order Paid",
				"updated_at": time.Now(),
			},
		}
		orderFilter := bson.M{"order_id": *invoice.Order_id}
		_, err = orderCollection.UpdateOne(ctx, orderFilter, update)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Failed to update order status to 'Order Paid' after payment"}`, http.StatusInternalServerError)
			return
		}
	}

	// Construct Success Response
	response := map[string]interface{}{
		"success": true,
		"message": "Invoice created successfully",
		"data":    invoice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateInvoice updates an existing invoice.
func UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	invoiceId := mux.Vars(r)["invoice_id"]
	var invoice models.Invoice
	if err := json.NewDecoder(r.Body).Decode(&invoice); err != nil {
		http.Error(w, `{"success": false, "message": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Initialize update object
	updateObj := bson.D{
		{Key: "updated_at", Value: time.Now()},
	}
	updateOrderStatus := false

	// Update payment_method, payment_status, payment_date, and total_price if provided
	if invoice.Payment_method != nil {
		updateObj = append(updateObj, bson.E{Key: "payment_method", Value: invoice.Payment_method})
	}
	if invoice.Payment_status != nil {
		updateObj = append(updateObj, bson.E{Key: "payment_status", Value: invoice.Payment_status})
		if *invoice.Payment_status == "PAID" {
			updateOrderStatus = true
		}
	}
	if !invoice.Payment_date.IsZero() {
		updateObj = append(updateObj, bson.E{Key: "payment_date", Value: invoice.Payment_date})
	}
	if invoice.TotalPrice > 0 {
		updateObj = append(updateObj, bson.E{Key: "total_price", Value: invoice.TotalPrice})
	}

	// Update the invoice in the database
	filter := bson.M{"invoice_id": invoiceId}
	opt := options.Update().SetUpsert(false)

	result, err := invoiceCollection.UpdateOne(ctx, filter, bson.D{{Key: "$set", Value: updateObj}}, opt)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Invoice update failed"}`, http.StatusInternalServerError)
		return
	}
	if result.MatchedCount == 0 {
		http.Error(w, `{"success": false, "message": "Invoice not found"}`, http.StatusNotFound)
		return
	}

	// Fetch the updated invoice
	var updatedInvoice models.Invoice
	err = invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&updatedInvoice)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving updated invoice"}`, http.StatusInternalServerError)
		return
	}

	// If payment_status is updated to PAID, update the corresponding order status
	if updateOrderStatus {
		_, err := orderCollection.UpdateOne(ctx,
			bson.M{"order_id": updatedInvoice.Order_id},
			bson.M{"$set": bson.M{"status": "Order Paid", "updated_at": time.Now()}},
		)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Failed to update order status"}`, http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Invoice updated successfully",
		"data":    updatedInvoice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteInvoice deletes an invoice.
func DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	invoiceId := mux.Vars(r)["invoice_id"]
	var invoice models.Invoice
	err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, `{"success": false, "message": "Invoice not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving invoice"}`, http.StatusInternalServerError)
		return
	}

	result, err := invoiceCollection.DeleteOne(ctx, bson.M{"invoice_id": invoiceId})
	if err != nil || result.DeletedCount == 0 {
		http.Error(w, `{"success": false, "message": "Error deleting invoice"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Invoice deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get Invoice by Order ID
func GetInvoiceByOrderId(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderId := mux.Vars(r)["order_id"]
	if orderId == "" {
		http.Error(w, `{"success": false, "message": "Invalid order ID"}`, http.StatusBadRequest)
		return
	}

	var invoice models.Invoice
	err := invoiceCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&invoice)
	if errors.Is(err, mongo.ErrNoDocuments) {
		http.Error(w, `{"success": false, "message": "Invoice not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving invoice"}`, http.StatusInternalServerError)
		return
	}

	// Construct Response
	response := map[string]interface{}{
		"success": true,
		"message": "Invoice retrieved successfully",
		"data":    invoice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get Invoices by User ID (with Pagination)
func GetInvoicesByUserId(w http.ResponseWriter, r *http.Request) {
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
		recordPerPage = 10 // Default records per page
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1 // Default to first page
	}

	startIndex := (page - 1) * recordPerPage

	// MongoDB aggregation pipeline for pagination
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "user_id", Value: userId}}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}

	cursor, err := invoiceCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving invoices"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var invoices []models.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding invoices data"}`, http.StatusInternalServerError)
		return
	}

	if len(invoices) == 0 {
		http.Error(w, `{"success": false, "message": "No invoices found for this user"}`, http.StatusNotFound)
		return
	}

	// Get total invoice count for the given user_id
	totalInvoices, err := invoiceCollection.CountDocuments(ctx, bson.M{"user_id": userId})
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving total invoice count"}`, http.StatusInternalServerError)
		return
	}

	// Construct response
	response := map[string]interface{}{
		"success": true,
		"message": "Invoices retrieved successfully",
		"data":    invoices,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_invoices":   totalInvoices,
			"total_pages":      (totalInvoices + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPendingInvoices returns paginated invoices with payment_status "PENDING"
func GetPendingInvoices(w http.ResponseWriter, r *http.Request) {
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
	skip := (page - 1) * recordPerPage

	// Query filter and options
	filter := bson.M{"payment_status": "PENDING"}
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(recordPerPage))

	// Query the database
	cursor, err := invoiceCollection.Find(ctx, filter, findOptions)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving pending invoices"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var invoices []models.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding pending invoices"}`, http.StatusInternalServerError)
		return
	}

	// Count total matching documents
	totalCount, err := invoiceCollection.CountDocuments(ctx, filter)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error counting pending invoices"}`, http.StatusInternalServerError)
		return
	}

	if len(invoices) == 0 {
		http.Error(w, `{"success": false, "message": "No pending invoices found"}`, http.StatusNotFound)
		return
	}

	// Success response with pagination
	response := map[string]interface{}{
		"success": true,
		"message": "Pending invoices retrieved successfully",
		"data":    invoices,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_invoices":   totalCount,
			"total_pages":      (totalCount + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPaidInvoices returns paginated invoices with payment_status "PAID"
func GetPaidInvoices(w http.ResponseWriter, r *http.Request) {
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
	skip := (page - 1) * recordPerPage

	// Filter and pagination options
	filter := bson.M{"payment_status": "PAID"}
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(recordPerPage))

	cursor, err := invoiceCollection.Find(ctx, filter, findOptions)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error retrieving paid invoices"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var invoices []models.Invoice
	if err := cursor.All(ctx, &invoices); err != nil {
		http.Error(w, `{"success": false, "message": "Error decoding paid invoices"}`, http.StatusInternalServerError)
		return
	}

	// Count total documents
	totalCount, err := invoiceCollection.CountDocuments(ctx, filter)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Error counting paid invoices"}`, http.StatusInternalServerError)
		return
	}

	if len(invoices) == 0 {
		http.Error(w, `{"success": false, "message": "No paid invoices found"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Paid invoices retrieved successfully",
		"data":    invoices,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_invoices":   totalCount,
			"total_pages":      (totalCount + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
