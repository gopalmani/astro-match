package auth

import (
	"astromatch/db"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

// SignupRequest struct to decode initial signup data
type SignupRequest struct {
	Token        string `json:"token"`
	SignupMethod string `json:"signupMethod"`
}

// SignupHandler handles the main signup flow
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}

	switch req.SignupMethod {
	case "email":
		var creds db.User
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, "Invalid user data", http.StatusBadRequest)
			return
		}
		handleEmailSignup(w, creds)
	case "google":
		handleGoogleSignupWithToken(w, req.Token)
	case "facebook":
		handleFacebookSignup(w, r)
	case "phone":
		var creds db.User
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, "Invalid user data", http.StatusBadRequest)
			return
		}
		handlePhoneSignup(w, creds)
	default:
		http.Error(w, "Invalid signup method", http.StatusBadRequest)
	}
}

func handleGoogleSignupWithToken(w http.ResponseWriter, token string) {
	user, err := VerifyGoogleTokenFromToken(token)
	if err != nil {
		http.Error(w, "Google signup failed", http.StatusUnauthorized)
		log.Printf("Token verification failed: %v", err)
		return
	}

	collection := db.Client.Database("astromatch").Collection("users")

	var existingUser db.User
	err = collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		// User already exists – login flow
		log.Printf("User already exists, logging in: %s", existingUser.Email)

		// Create JWT token for session (or session-based logic)
		tokenString, err := GenerateJWT(user.ID)
		if err != nil {
			http.Error(w, "Failed to generate login token", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"message": "User logged in successfully",
			"token":   tokenString,
		})
		return
	}

	// New user registration flow
	user.ID = uuid.New().String()
	user.SignupMethod = "google"
	user.IsVerified = true

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	tokenString, err := GenerateJWT(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate login token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Google user registered successfully",
		"token":   tokenString,
	})
}

// VerifyGoogleTokenFromToken verifies the token and extracts user info from the ID token (JWT)
func VerifyGoogleTokenFromToken(token string) (db.User, error) {
	ctx := context.Background()

	// Create OAuth2 service for token verification
	oauth2Service, err := oauth2.NewService(ctx, option.WithoutAuthentication())
	if err != nil {
		log.Printf("Failed to initialize OAuth2 service: %v", err)
		return db.User{}, err
	}

	tokenInfoService := oauth2Service.Tokeninfo()
	tokenInfoService.IdToken(token) // Verify token validity

	_, err = tokenInfoService.Do()
	if err != nil {
		log.Printf("Token verification failed: %v", err)
		return db.User{}, errors.New("invalid Google token")
	}

	// Decode the ID token (JWT) to extract user info
	claims, err := decodeGoogleIDToken(token)
	if err != nil {
		log.Printf("Failed to decode ID token: %v", err)
		return db.User{}, errors.New("failed to extract user info")
	}

	user := db.User{
		Name:       claims["name"].(string),
		Email:      claims["email"].(string),
		ProfilePic: claims["picture"].(string),
		IsVerified: true,
	}

	return user, nil
}

// decodeGoogleIDToken decodes the JWT and extracts user claims (email, name, picture)
func decodeGoogleIDToken(idToken string) (map[string]interface{}, error) {
	// Split the JWT (Header.Payload.Signature)
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid JWT format")
	}

	// Base64 decode the payload
	payload, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return nil, err
	}

	// Unmarshal payload into a map
	var claims map[string]interface{}
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, err
	}

	// Check for required fields
	if claims["email"] == nil || claims["name"] == nil || claims["picture"] == nil {
		return nil, errors.New("incomplete token claims")
	}

	return claims, nil
}

// handleEmailSignup manages email-based signup with password hashing
func handleEmailSignup(w http.ResponseWriter, creds db.User) {
	if db.Client == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		log.Println("MongoDB client is nil")
		return
	}

	collection := db.Client.Database("astromatch").Collection("users")
	ctx := context.TODO()

	// Check if email already exists
	var existingUser db.User
	err := collection.FindOne(ctx, bson.M{"email": creds.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(creds.Password)
	if err != nil {
		http.Error(w, "Password hashing failed", http.StatusInternalServerError)
		return
	}

	// Assign UUID and hash password
	creds.ID = uuid.New().String()
	creds.Password = hashedPassword

	// Insert new user into MongoDB
	_, err = collection.InsertOne(ctx, creds)
	if err != nil {
		http.Error(w, "Failed to insert user data", http.StatusInternalServerError)
		log.Printf("Failed to insert user data into the database: %v", err)
		return
	}

	// Generate and send OTP
	otp := GenerateOTP()
	err = SendOTPViaEmail(creds.Email, otp)
	if err != nil {
		http.Error(w, "Failed to send OTP", http.StatusInternalServerError)
		return
	}

	// Cache OTP or store in MongoDB
	storeOTP(creds.Email, otp)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// // handleGoogleSignup manages Google OAuth signup
func handleGoogleSignup(w http.ResponseWriter, r *http.Request) {
	user, err := VerifyGoogleToken(r)
	if err != nil {
		http.Error(w, "Google signup failed", http.StatusUnauthorized)
		return
	}

	collection := db.Client.Database("astromatch").Collection("users")

	// Check if user already exists by email
	var existingUser db.User
	err = collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "User already exists with this email", http.StatusConflict)
		return
	}

	user.ID = uuid.New().String()
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Google user registered successfully"})
}

// handleFacebookSignup manages Facebook OAuth signup
func handleFacebookSignup(w http.ResponseWriter, r *http.Request) {
	user, err := VerifyFacebookToken(r)
	if err != nil {
		http.Error(w, "Facebook signup failed", http.StatusUnauthorized)
		return
	}

	collection := db.Client.Database("astromatch").Collection("users")

	// Check if user already exists by email
	var existingUser db.User
	err = collection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "User already exists with this email", http.StatusConflict)
		return
	}

	user.ID = uuid.New().String()
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		http.Error(w, "User creation failed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Facebook user registered successfully"})
}

// handlePhoneSignup handles phone-based signup
func handlePhoneSignup(w http.ResponseWriter, creds db.User) {
	collection := db.Client.Database("astromatch").Collection("users")
	ctx := context.TODO()

	// Check if phone already exists
	var existingUser db.User
	err := collection.FindOne(ctx, bson.M{"phone": creds.Phone}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "Phone number already exists", http.StatusConflict)
		log.Printf("Phone number already exists: %s", creds.Phone)
		return
	} else if err != mongo.ErrNoDocuments {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		log.Printf("Error querying phone number: %v", err)
		return
	}

	creds.ID = uuid.New().String()

	_, err = collection.InsertOne(ctx, creds)
	if err != nil {
		// Log full error details if insertion fails
		log.Printf("Failed to insert phone user - Full Error: %+v", err)

		// Handle duplicate key error
		if mongo.IsDuplicateKeyError(err) {
			http.Error(w, "Phone number already exists", http.StatusConflict)
			log.Printf("Duplicate key error during insertion: %v", err)
		} else {
			http.Error(w, "Failed to insert phone user", http.StatusInternalServerError)
		}
		return
	}
	// Generate and send OTP
	otp := GenerateOTP()
	err = SendOTPViaPhone(creds.Phone, otp)
	if err != nil {
		http.Error(w, "Failed to send OTP", http.StatusInternalServerError)
		return
	}

	// Cache OTP or store in MongoDB
	storeOTP(creds.Phone, otp)

	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent. Verify to complete signup."})
}

// storeOTP caches the OTP or saves it to MongoDB for verification
func storeOTP(identifier, otp string) {
	collection := db.Client.Database("astromatch").Collection("user_otp")
	ctx := context.TODO()

	// Store OTP with expiration in the database
	_, err := collection.InsertOne(ctx, bson.M{
		"identifier": identifier,
		"otp":        otp,
		"expires_at": time.Now().Add(5 * time.Minute),
	})

	if err != nil {
		log.Printf("Failed to store OTP for %s: %v", identifier, err)
	}
}
