package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"agendum/internal/db"
	"agendum/internal/models"
)

var dbClient *db.DynamoDBClient

func init() {
	var err error
	dbClient, err = db.NewDynamoDBClient()
	if err != nil {
		panic(err)
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := dbClient.CreateUser(context.Background(), user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}