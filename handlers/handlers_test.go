package handlers

import (
	"Users/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// testing successful creation of user
func TestCreateUser(t *testing.T) {
	// Setup mock MongoDB
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("success", func(mtT *mtest.T) {
		// Mock responses: user doesn't exist, then insert succeeds
		mtT.AddMockResponses(
			//if the mocktest does not return a user insert instead ie 0
			mtest.CreateCursorResponse(0, "usersdb.users", mtest.FirstBatch), // FindOne returns nothing
			//user created successfully
			mtest.CreateSuccessResponse(), // InsertOne succeeds
		)
		payload := `{"username":"fake user","email":"fake@example.com","password":"fake1234"}`
		//creating the request for creating a user
		req := httptest.NewRequest(http.MethodPost, "/api/create-user", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		//calling the handler struct
		h := &Handler{
			Client: mtT.Client,
		}
		//calling the handler function
		h.CreateUser(rec, req)
		//comparing the codes of the request and the recorder(rec)
		if rec.Code != http.StatusOK {
			t.Fatal("Expected 200 but got", rec.Code)
		}
		//creating the response struct and decoding it into the response variable
		var response models.Message
		err := json.NewDecoder(rec.Body).Decode(&response)
		if err != nil {
			t.Errorf("Error decoding the response message")
		}
		//aserting if the response and request was right
		expected := "User created successfully"
		if response.Message != expected {
			t.Errorf("expected message %s but got %s", expected, response.Message)
		}
	})
}

// testing invalid email
func TestCreateUser_InvalidEmail(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("invalid email", func(mtT *mtest.T) {
		payload := `{"email":"invalidemail"}`
		req := httptest.NewRequest(http.MethodPost, "/api/create-user", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		h := &Handler{
			Client: mtT.Client,
		}
		h.CreateUser(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatal("Expected 400 but got", rec.Code)
		}
	})
}

// checking if a user exits test
func TestCreateUser_UserExits(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("User exists", func(mtT *mtest.T) {
		mtT.AddMockResponses(
			mtest.CreateCursorResponse(1, "usersdb.users", mtest.FirstBatch, bson.D{{Key: "email", Value: "fakeemail@mail.com"}}),
		)
		payload := `{"username":"FakeUser","email":"fakeemail@mail.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/create-user", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		h := &Handler{
			Client: mtT.Client,
		}
		h.CreateUser(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("Expected 409 but got %d", rec.Code)
		}
	})

}

// testing successful update of user
func TestUpdateUser(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("success", func(mtT *mtest.T) {
		mtT.AddMockResponses(
			//Every success test must return 0 or output zero remember!!
			mtest.CreateCursorResponse(0, "usersdb.users", mtest.FirstBatch), // No duplicate email
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}),          // UpdateOne matches 1 document
		)
		payload := `{"username":"Newname","email":"new@mail.com","password":"newpass123"}`
		req := httptest.NewRequest(http.MethodPut, "/api/update-user/507f1f77bcf86cd799439011", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		h := &Handler{
			Client: mtT.Client,
		}
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/update-user/{id}", h.UpdateUser)

		// Use the mux instead of calling handler directly
		mux.ServeHTTP(rec, req)
		t.Logf("Status: %d, Body: %s", rec.Code, rec.Body.String())

		if rec.Code != http.StatusOK {
			t.Errorf("Expected 200 but got %d", rec.Code)
		}
	})

}
func TestUpdateUser_InavlidID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("Invalid_ID", func(mtT *mtest.T) {
		payload := `{"invalid-email-address":"invalid"}`
		req := httptest.NewRequest(http.MethodPut, "/api/update-user/invalid-id", bytes.NewBufferString(payload))
		rec := httptest.NewRecorder()
		h := &Handler{
			Client: mtT.Client,
		}
		// Add ServeMux to make PathValue work
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/update-user/{id}", h.UpdateUser)
		mux.ServeHTTP(rec, req)
		h.UpdateUser(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 200 but got %d", rec.Code)
		}
	})
}
