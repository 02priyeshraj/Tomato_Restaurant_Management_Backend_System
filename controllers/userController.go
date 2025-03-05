package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	database "github.com/02priyeshraj/Hotel_Management_Backend/config"
	"github.com/02priyeshraj/Hotel_Management_Backend/helper"
	"github.com/02priyeshraj/Hotel_Management_Backend/models"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// pagination parameters
	recordPerPage, err := strconv.Atoi(r.URL.Query().Get("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage

	// MongoDB aggregation pipeline
	matchStage := bson.D{{Key: "$match", Value: bson.D{}}}
	skipStage := bson.D{{Key: "$skip", Value: startIndex}}
	limitStage := bson.D{{Key: "$limit", Value: int64(recordPerPage)}}
	projectStage := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "email", Value: 1},
			{Key: "first_name", Value: 1},
			{Key: "last_name", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "phone", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "updated_at", Value: 1},
		}},
	}

	result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, "Error occurred while listing users", http.StatusInternalServerError)
		return
	}

	var allUsers []bson.M
	if err = result.All(ctx, &allUsers); err != nil {
		http.Error(w, "Error decoding user data", http.StatusInternalServerError)
		return
	}

	// Count total users for pagination
	totalUsers, err := userCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		http.Error(w, "Error retrieving total user count", http.StatusInternalServerError)
		return
	}

	// Prepare JSON response
	response := map[string]interface{}{
		"success": true,
		"message": "Users retrieved successfully",
		"data":    allUsers,
		"pagination": map[string]interface{}{
			"current_page":     page,
			"records_per_page": recordPerPage,
			"total_users":      totalUsers,
			"total_pages":      (totalUsers + int64(recordPerPage) - 1) / int64(recordPerPage),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	params := mux.Vars(r)
	userId := params["user_id"]

	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Prepare JSON response
	response := map[string]interface{}{
		"success": true,
		"message": "User fetched successfully",
		"data": map[string]interface{}{
			"user_id":    user.User_id,
			"first_name": user.First_name,
			"last_name":  user.Last_name,
			"email":      user.Email,
			"phone":      user.Phone,
			"created_at": user.Created_at,
			"updated_at": user.Updated_at,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if email already exists
	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		http.Error(w, "Error checking email", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	// Hash password
	password := HashPassword(*user.Password)
	user.Password = &password

	// Set user metadata
	user.Created_at = time.Now()
	user.Updated_at = time.Now()
	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()

	// Generate authentication tokens
	token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
	user.Token = &token
	user.Refresh_Token = &refreshToken

	// Insert into MongoDB
	_, insertErr := userCollection.InsertOne(ctx, user)
	if insertErr != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	// Prepare JSON response
	response := map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"data": map[string]interface{}{
			"user_id":       user.User_id,
			"first_name":    user.First_name,
			"last_name":     user.Last_name,
			"email":         user.Email,
			"phone":         user.Phone,
			"token":         user.Token,
			"refresh_token": user.Refresh_Token,
			"created_at":    user.Created_at,
			"updated_at":    user.Updated_at,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User
	var foundUser models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find the user by email
	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Verify password
	passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
	if !passwordIsValid {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	// Generate new tokens
	token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
	helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

	// Update the foundUser object with tokens
	foundUser.Token = &token
	foundUser.Refresh_Token = &refreshToken

	// Prepare JSON response
	response := map[string]interface{}{
		"success": true,
		"message": "User logged-in successfully",
		"data": map[string]interface{}{
			"user_id":       foundUser.User_id,
			"first_name":    foundUser.First_name,
			"last_name":     foundUser.Last_name,
			"email":         foundUser.Email,
			"phone":         foundUser.Phone,
			"token":         foundUser.Token,
			"refresh_token": foundUser.Refresh_Token,
			"created_at":    foundUser.Created_at,
			"updated_at":    foundUser.Updated_at,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Extract token from the Authorization header
	clientToken := r.Header.Get("Authorization")
	if clientToken == "" {
		http.Error(w, "No Authorization header provided", http.StatusUnauthorized)
		return
	}

	tokenParts := strings.Split(clientToken, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
		return
	}

	tokenString := tokenParts[1]
	claims, errMsg := helper.ValidateToken(tokenString)
	if errMsg != "" {
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	// Remove token from the database
	updateObj := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "token", Value: nil},
			{Key: "refresh_token", Value: nil},
		}},
	}

	filter := bson.M{"user_id": claims.Uid}
	_, err := userCollection.UpdateOne(ctx, filter, updateObj)
	if err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	// Success response
	response := map[string]interface{}{
		"success": true,
		"message": "User logged out successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	if err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword)); err != nil {
		return false, "Incorrect password"
	}
	return true, ""
}
