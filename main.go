package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/mail"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

// ------------------ STRUCTS ------------------

type User struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Username  string    `bson:"username" json:"username"`
	Email     string    `bson:"email" json:"email"`
	Password  string    `bson:"password" json:"password"`
	Status    string    `bson:"status" json:"status"`
	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
}

type Message struct {
	Message string `json:"message"`
}

// -----------------MIDDLEWARES------------------------------------
func panicMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered", slog.Any("error", err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
//method checking middleware 
func 

// ------------------ GLOBAL VARIABLES ------------------

var client *mongo.Client

// ------------------ HELPER FUNCTIONS ------------------
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// ------------------ DB CONNECTION ------------------

func connectDB() *mongo.Client {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: no .env file found")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI not found in environment")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("MongoDB ping error:", err)
	}
	
	logger.Info(" Connected to MongoDB successfully!")
	return client
}

// ------------------ CREATE USER ------------------

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}
	if _, err := mail.ParseAddress(user.Email); err != nil {
		http.Error(w, "Invalid Email", http.StatusBadRequest)
		return
	}
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	user.CreatedAt = time.Now()
	user.Status = "active"

	collection := client.Database("usersdb").Collection("users")
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
	json.NewEncoder(w).Encode(Message{Message: "User created successfully"})
}

// ------------------ UPDATE USER ------------------

func updateUser(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "Missing user ID ", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	var userUpdateData User
	err = json.NewDecoder(r.Body).Decode(&userUpdateData)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := client.Database("usersdb").Collection("users")
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
		pass_hash, err := hashPassword(userUpdateData.Password)
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
	json.NewEncoder(w).Encode(Message{Message: "User updated successfully"})
}

// --------------FETCH ALL  USERS ------------------
func fetchAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "could not fectch all users", http.StatusMethodNotAllowed)
		return
	}
	//connecting to the mongo db collection
	collection := client.Database("usersdb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, "could not fetch users", http.StatusInternalServerError)
		return
	}
	defer results.Close(ctx)
	//decoding the json from
	var users []User
	for results.Next(ctx) {
		var user User
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
func fetchAllEmails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}
	//getting the collection'
	collection := client.Database("usersdb").Collection("users")
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
func deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "user Id is missing", http.StatusBadRequest)
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID ", http.StatusBadRequest)
		return
	}
	collection := client.Database("usersdb").Collection("users")
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
	response := Message{
		Message: "User deleted succesfully",
	}
	json.NewEncoder(w).Encode(response)
}

// ------UPDATE STATUS---------------------
func updateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "Missing user ID ", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	var user User
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

	collection := client.Database("usersdb").Collection("users")
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
	json.NewEncoder(w).Encode(Message{Message: "User updated successfully"})
}

// ------------------ MAIN ------------------

func main() {
	//logger using slog to log in json format
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client = connectDB()
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {

			logger.Error("Error disconnecting from mongoDb", "error", err)
		}
	}()
	//using a server mux to map the requests to the handlers
	mux := http.NewServeMux()

	mux.HandleFunc("/api/create-user", createUser)
	mux.HandleFunc("/api/update-user", updateUser)
	mux.HandleFunc("/api/users", fetchAllUsers)
	mux.HandleFunc("/api/delete-user", deleteUser)
	mux.HandleFunc("/api/update-Status", updateStatus)
	mux.HandleFunc("/api/emails", fetchAllEmails)
	//Wrapping the mux around the panic middleware
	handlerforPanicRecovery := panicMiddleware(logger)(mux)
	server := &http.Server{
		Addr:    ":8080",
		Handler: handlerforPanicRecovery,
	}
	logger.Info("Server started on port 8080")
	go func() {

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server has been shutdown...")
	}
	logger.Info("Server exited...")
}
