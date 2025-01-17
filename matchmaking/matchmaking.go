package matchmaking

import (
	"astromatch/auth"
	"astromatch/db"
	"encoding/json"
	"net/http"
)

func MatchHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := db.GetUserByID(claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	compatibleUsers, _ := FindCompatibleUsers(user)
	json.NewEncoder(w).Encode(compatibleUsers)
}
