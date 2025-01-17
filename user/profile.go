package user

import (
	"astromatch/db"
	"encoding/json"
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

// GetOrUpdateProfile - Handles GET and PUT for user profiles
func GetOrUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getUserProfile(w, userID)
	case http.MethodPut:
		updateUserProfile(w, r, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Fetch User Profile
func getUserProfile(w http.ResponseWriter, userID string) {
	user, err := db.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch user profile", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Update User Profile
func updateUserProfile(w http.ResponseWriter, r *http.Request, userID string) {
	var updatedUser db.User
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = db.UpdateUserProfile(userID, updatedUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update user profile", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "User updated successfully"}
	json.NewEncoder(w).Encode(response)
}
