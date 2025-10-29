package handlers

import (
	"Users/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

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
func TestCreateUser_UserExits(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("User exists", func(mtT *mtest.T) {
		mtT.AddMockResponses(
			mtest.CreateCursorResponse(1, "usersdb.users", mtest.FirstBatch),bson.M{"email":"xxx@mail.com"}),
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
