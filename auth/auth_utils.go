package auth

import (
	"astromatch/db"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/huandu/facebook/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

var jwtKey = []byte("supersecretkey")

// Claims struct for JWT
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a JWT token for user authentication
func GenerateJWT(userID string) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateJWT checks if the JWT token is valid
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}

// HashPassword securely hashes user passwords
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a hashed password with a plain text one
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// VerifyGoogleToken verifies Google OAuth token
func VerifyGoogleToken(r *http.Request) (db.User, error) {
	var req struct {
		Token string `json:"token"`
	}

	// json.NewDecoder(r.Body).Decode(&req)
	// err := json.NewDecoder(r.Body).Decode(&req)
	// if err != nil {
	// 	log.Printf("Failed to decode request body: %v", err)
	// 	return db.User{}, errors.New("invalid request body")
	// }

	// if req.Token == "" {
	// 	log.Println("Missing Google ID token in request")
	// 	return db.User{}, errors.New("missing Google token")
	// }
	// Read body only once
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		return db.User{}, errors.New("unable to read request body")
	}

	// Log the raw request body for debugging
	log.Printf("Raw Request Body: %s", string(body))

	// Unmarshal JSON payload manually
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Printf("Failed to decode request body: %v", err)
		return db.User{}, errors.New("invalid request body format")
	}

	// Get Google Client ID from environment variables
	clientID := "905964697703-a3smbmrtsnvcgpssqmdoh2stoietokc0.apps.googleusercontent.com"
	if clientID == "" {
		log.Println("Google Client ID is not set")
		return db.User{}, errors.New("Google Client ID is missing")
	}

	// Validate the token
	payload, err := idtoken.Validate(r.Context(), req.Token, clientID)
	if err != nil {
		log.Printf("Google token validation failed: %v", err)
		return db.User{}, errors.New("invalid Google token")
	}

	// Extract email, name, and profile pic from the payload
	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	return db.User{
		Email:        email,
		Name:         name,
		ProfilePic:   picture,
		SignupMethod: "google",
	}, nil
}

// VerifyFacebookToken verifies Facebook OAuth token
func VerifyFacebookToken(r *http.Request) (db.User, error) {
	var req struct {
		Token string `json:"token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	res, err := facebook.Get("/me", facebook.Params{
		"fields":       "id,name,email",
		"access_token": req.Token,
	})
	if err != nil {
		return db.User{}, errors.New("invalid Facebook token")
	}

	return db.User{
		Email: res.Get("email").(string),
		Name:  res.Get("name").(string),
	}, nil
}

// VerifyPhoneNumber simulates phone number verification
func VerifyPhoneNumber(phone string) bool {
	return len(phone) >= 10
}
