package repository

import (
	"Users/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
bringing the mongodb functions into the repository layer
*/
type mongoClient struct {
	client *mongo.Client
}

/*
function that sends the functions in this repo.go layer
to a an interface called UserRepository in interfaces.go
*/
func NewMongo(client *mongo.Client) UserRepository {
	return &mongoClient{client: client}
}

//-----CREATE USER FUNCTION-----

// users *models.User is a pointer to the user struct
func (m *mongoClient) CreateUser(user *models.User) error {
	user.CreatedAt = time.Now()
	user.Status = "active"
	//database actions for creating user
	collection := m.client.Database("usersdb").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user already exists")
		}
		return err

	}
	//return user id
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid.Hex()
	}
	return nil
}

// -----UPDATE USER FUNCTION-----
func (m *mongoClient) UpdateUser(user *models.User) error {
	//update user logic
	objID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return errors.New("invalid user ID")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := m.client.Database("usersdb").Collection("users")
	updateFields := bson.M{}
	if user.Email != "" {
		updateFields["email"] = user.Email
	}
	if user.Username != "" {
		updateFields["username"] = user.Username
	}
	if user.Password != "" {
		updateFields["password"] = user.Password
	}
	if len(updateFields) == 0 {
		return errors.New("no fields to update")
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil

}

// -----------DELETE USER FUNCTION---------------
func (m *mongoClient) DeleteUser(user *models.User) error {
	//delete user logic
	objID, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return errors.New("invalid user ID")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := m.client.Database("usersdb").Collection("users")
	filter := bson.M{"_id": objID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

// -----------FETCH ALL USERS FUNCTION--------
func (m *mongoClient) FetchAllUsers() ([]models.User, error) {
	//fetch all users logic
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := m.client.Database("usersdb").Collection("users")
	filter := bson.M{}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var users []models.User
	for cursor.Next(ctx) {
		var user models.User
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
func (m *mongoClient) FetchUserByID(id string) (*models.User, error) {
	//fetch user by id logic
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := m.client.Database("usersdb").Collection("users")
	filter := bson.M{"_id": objID}
	var user models.User
	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil

}
func (m *mongoClient) UpdateUserStatus(id string, status string) error {
	//update user status logic
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := m.client.Database("usersdb").Collection("users")
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": status}}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil

}
