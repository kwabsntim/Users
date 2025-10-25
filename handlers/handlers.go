package handlers

import (
	"Users/models"
	"Users/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
	Client *mongo.Client
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}
	if _, err := mail.ParseAddress(user.Email); err != nil {
		http.Error(w, "Invalid Email", http.StatusBadRequest)
		return
	}
	user.Password, err = utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	user.CreatedAt = time.Now()
	user.Status = "active"

	collection := h.Client.Database("usersdb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// check if user exists
	existing := collection.FindOne(ctx, bson.M{"email": user.Email})
	if existing.Err() == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// insert user
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	fmt.Println(" User created with ID:", result.InsertedID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Message{Message: "User created successfully"})
}

// ------------------ UPDATE USER ------------------

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {

	userIDStr := r.PathValue("id")
	if userIDStr == "" {
		http.Error(w, "Missing user ID ", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	var userUpdateData models.User
	err = json.NewDecoder(r.Body).Decode(&userUpdateData)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := h.Client.Database("usersdb").Collection("users")
	if userUpdateData.Email != "" {
		filter := bson.M{"email": userUpdateData.Email,
			"_id": bson.M{"$ne": objID},
		}
		count, err := collection.CountDocuments(ctx, filter)
		if err != nil {
			http.Error(w, "Error checking email uniqueness", http.StatusInternalServerError)
			return
		}
		if count > 0 {
			http.Error(w, "Invalid value for email", http.StatusBadRequest)
			return
		}

	}
	updateFields := bson.M{}
	if userUpdateData.Email != "" {
		updateFields["email"] = userUpdateData.Email

	}
	if userUpdateData.Username != "" {
		updateFields["username"] = userUpdateData.Username
	}
	if userUpdateData.Password != "" {
		pass_hash, err := utils.HashPassword(userUpdateData.Password)
		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}
		updateFields["password"] = pass_hash
	}
	if len(updateFields) == 0 {
		http.Error(w, " No fields to update", http.StatusBadRequest)
		return
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": updateFields}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Message{Message: "User updated successfully"})
}

// --------------FETCH ALL  USERS ------------------
func (h *Handler) FetchAllUsers(w http.ResponseWriter, r *http.Request) {

	//connecting to the mongo db collection
	collection := h.Client.Database("usersdb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "could not fetch users", http.StatusInternalServerError)
		return
	}
	defer results.Close(ctx)
	//decoding the json from
	var users []models.User
	for results.Next(ctx) {
		var user models.User
		if err = results.Decode(&user); err != nil {
			http.Error(w, "error decoding user data", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	if err = results.Err(); err != nil {
		http.Error(w, "error fectching users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(users)

}

// -----------------FETCH ALL EMAILS-----------------
func (h *Handler) FetchAllEmails(w http.ResponseWriter, r *http.Request) {

	//getting the collection'
	collection := h.Client.Database("usersdb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	projection := bson.M{
		"email": 1, "_id": 0,
	}
	findOptions := options.Find().SetProjection(projection)
	results, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		http.Error(w, "Error fetching emails", http.StatusInternalServerError)
		return
	}

	type email struct {
		Email string `bson:"email" json:"email"`
	}
	//decoding the JSON from the DB
	var emails []email
	for results.Next(ctx) {
		var email email
		results.Decode(&email)
		emails = append(emails, email)
	}

	if err = results.All(ctx, &emails); err != nil {
		http.Error(w, "Error fetching emails", http.StatusInternalServerError)
	}
	defer results.Close(ctx)
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(emails)
}

// --------------DELETE USER ------------------
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {

	userIDStr := r.PathValue("id")
	if userIDStr == "" {
		http.Error(w, "user Id is missing", http.StatusBadRequest)
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID ", http.StatusBadRequest)
		return
	}
	collection := h.Client.Database("usersdb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": userID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("content-type", "application/json")
	response := models.Message{
		Message: "User deleted succesfully",
	}
	json.NewEncoder(w).Encode(response)
}

// ------UPDATE STATUS---------------------
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {

	userIDStr := r.PathValue("id")
	if userIDStr == "" {
		http.Error(w, "Missing user ID ", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	var user models.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"status": user.Status,
	}}

	collection := h.Client.Database("usersdb").Collection("users")
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		http.Error(w, "Error updating user status ", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Message{Message: "User updated successfully"})
}
