package user

import (
	"astromatch/auth"
	"astromatch/db"
	"encoding/json"
	"net/http"
)

func UpdatePreferencesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateJWT(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var prefs db.UserPreferences
	err = json.NewDecoder(r.Body).Decode(&prefs)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = db.UpdateUserPreferences(claims.UserID, prefs)
	if err != nil {
		http.Error(w, "Preferences update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Preferences updated successfully"})
}
