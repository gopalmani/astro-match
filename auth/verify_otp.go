package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"time"

	"astromatch/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VerifyUserRequest struct handles incoming OTP verification requests
type VerifyUserRequest struct {
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
	OTP   string `json:"otp"`
}

// VerifyUserHandler verifies a user using email or phone and OTP
func VerifyUserHandler(w http.ResponseWriter, r *http.Request) {
	if db.Client == nil {
		http.Error(w, "MongoDB client not initialized", http.StatusInternalServerError)
		return
	}

	var req VerifyUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validateRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := verifyUser(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User verified successfully."})
}

// validateRequest validates email/phone and OTP
func validateRequest(req VerifyUserRequest) error {
	if req.Email == "" && req.Phone == "" {
		return errors.New("either email or phone is required")
	}
	if req.OTP == "" {
		return errors.New("OTP is required")
	}

	if req.Email != "" {
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		if match, _ := regexp.MatchString(emailRegex, req.Email); !match {
			return errors.New("invalid email format")
		}
	}

	if req.Phone != "" {
		phoneRegex := `^\d{10,15}$`
		if match, _ := regexp.MatchString(phoneRegex, req.Phone); !match {
			return errors.New("invalid phone number")
		}
	}

	return nil
}

// verifyUser handles OTP verification and updates the user's isVerified flag
func verifyUser(req VerifyUserRequest) error {

	collection := db.Client.Database("astromatch").Collection("user_otp")

	// Convert current time to UTC and truncate precision
	now := time.Now().UTC().Truncate(time.Second)

	// Log all OTPs for the identifier to check if the record exists
	cursor, _ := collection.Find(context.TODO(), bson.M{"identifier": req.Email})
	var otpResults []db.UserOTP
	if err := cursor.All(context.TODO(), &otpResults); err == nil {
		log.Printf("All OTPs for identifier %s: %+v", req.Email, otpResults)
	}

	// Create the filter for OTP lookup
	filter := bson.D{
		{Key: "otp", Value: req.OTP},
		{Key: "expires_at", Value: bson.D{{Key: "$gt", Value: now}}},
		{Key: "identifier", Value: req.Email},
	}

	log.Printf("Running Query for identifier: %s, OTP: %s, Expires at (greater than): %v", req.Email, req.OTP, now)

	// Perform the query to fetch the OTP
	var userOTP db.UserOTP
	err := collection.FindOne(context.TODO(), filter).Decode(&userOTP)
	if err != nil {
		log.Printf("OTP Verification failed. Filter used: %v | Error: %v", filter, err)
		return errors.New("invalid or expired OTP")
	}

	// Log the result of the comparison
	isValid := userOTP.ExpiresAt.After(now)
	log.Printf("OTP Found. Expires at: %v | Current time (UTC): %v | Expires at > now: %v", userOTP.ExpiresAt, now, isValid)

	if !isValid {
		return errors.New("otp has expired")
	}

	// Update user's verification status
	usersCollection := db.Client.Database("astroMatch").Collection("users")

	updateFilter := bson.D{}
	if req.Email != "" {
		updateFilter = append(updateFilter, bson.E{Key: "email", Value: req.Email})
	} else {
		updateFilter = append(updateFilter, bson.E{Key: "phone", Value: req.Phone})
	}
	updateFilter = append(updateFilter, bson.E{Key: "isverified", Value: false})

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "isverified", Value: true}}}}
	_, err = usersCollection.UpdateOne(context.TODO(), updateFilter, update, options.Update())
	if err != nil {
		log.Printf("Failed to update user verification status for %s. Error: %v", req.Email, err)
		return errors.New("failed to update user status")
	}

	log.Printf("User marked as verified for %s.", req.Email)
	return nil
}
