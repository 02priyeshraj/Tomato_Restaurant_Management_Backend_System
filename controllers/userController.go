package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
			{Key: "email", Value: 1},
			{Key: "first_name", Value: 1},
			{Key: "last_name", Value: 1},
			{Key: "user_id", Value: 1},
		}},
	}

	result, err := userCollection.Aggregate(ctx, mongo.Pipeline{matchStage, skipStage, limitStage, projectStage})
	if err != nil {
		http.Error(w, "Error occurred while listing users", http.StatusInternalServerError)
		return
	}

	var allUsers []bson.M
	if err = result.All(ctx, &allUsers); err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allUsers)
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

	responseUser := struct {
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Email     string    `json:"email"`
		Avatar    *string   `json:"avatar"`
		Phone     string    `json:"phone"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UserID    string    `json:"user_id"`
	}{
		FirstName: *user.First_name,
		LastName:  *user.Last_name,
		Email:     *user.Email,
		Avatar:    user.Avatar,
		Phone:     *user.Phone,
		CreatedAt: user.Created_at,
		UpdatedAt: user.Updated_at,
		UserID:    user.User_id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseUser)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		http.Error(w, "Error checking email", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	password := HashPassword(*user.Password)
	user.Password = &password

	user.Created_at = time.Now()
	user.Updated_at = time.Now()
	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()

	token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
	user.Token = &token
	user.Refresh_Token = &refreshToken

	_, insertErr := userCollection.InsertOne(ctx, user)
	if insertErr != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
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

	err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
	if !passwordIsValid {
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
	helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)

	// Create a response struct excluding the password
	responseUser := struct {
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		UserID       string `json:"user_id"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}{
		Email:        *foundUser.Email,
		FirstName:    *foundUser.First_name,
		LastName:     *foundUser.Last_name,
		UserID:       foundUser.User_id,
		Token:        token,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseUser)
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
