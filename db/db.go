package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User struct represents user data
type User struct {
	ID           string `bson:"id,omitempty" json:"id,omitempty"`
	Name         string `bson:"name" json:"name"`
	Email        string `bson:"email" json:"email"`
	Password     string `bson:"password" json:"password"`
	Phone        string `bson:"phone" json:"phone"`
	Birthdate    string `bson:"birthdate,omitempty" json:"birthdate,omitempty"`
	ZodiacSign   string `bson:"zodiac_sign,omitempty" json:"zodiacSign,omitempty"`
	ProfilePic   string `bson:"profile_pic,omitempty" json:"profilePic,omitempty"`
	SignupMethod string `bson:"signup_method" json:"signupMethod"`
	IsVerified   bool   `bson:"is_verified" json:"isVerified"`
}

// UserOTP struct represents OTP data for verification
type UserOTP struct {
	Email     string    `bson:"email,omitempty" json:"email,omitempty"`
	Phone     string    `bson:"phone,omitempty" json:"phone,omitempty"`
	OTP       string    `bson:"otp" json:"otp"`
	ExpiresAt time.Time `bson:"expires_at"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	UserID        string   `bson:"user_id"`
	PreferredSign string   `bson:"preferred_sign"`
	MaxDistance   int      `bson:"max_distance"`
	Interests     []string `bson:"interests"`
}

// MongoDB client
var Client *mongo.Client

// Initialize MongoDB connection
func InitDB() {
	connectionString := "mongodb+srv://astro:astromatch@cluster0.9qerzmd.mongodb.net/?retryWrites=true&w=majority"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(connectionString)

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB not reachable:", err)
	}

	fmt.Println("Connected to MongoDB")
}

// GetUserByEmail fetches a user by their email from MongoDB
func GetUserByEmail(email string) (User, error) {
	collection := Client.Database("astroMatch").Collection("users")
	ctx := context.TODO()

	var user User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, errors.New("user not found")
		}
		log.Printf("Error fetching user by email: %v", err)
		return User{}, err
	}
	return user, nil
}

// GetUserByID fetches a user by their ID
func GetUserByID(userID string) (User, error) {
	collection := Client.Database("astroMatch").Collection("users")
	ctx := context.TODO()

	var user User
	err := collection.FindOne(ctx, bson.M{"id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, errors.New("user not found")
		}
		log.Printf("Error fetching user by ID: %v", err)
		return User{}, err
	}
	return user, nil
}

// CreateUser inserts a new user into the database
func CreateUser(user User) error {
	collection := Client.Database("astroMatch").Collection("users")
	ctx := context.TODO()

	user.ID = uuid.New().String() // Generate UUID for user ID
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return err
	}
	return nil
}

// GetAllUsers retrieves all users from the database
func GetAllUsers() ([]User, error) {
	collection := Client.Database("astroMatch").Collection("users")
	ctx := context.TODO()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Printf("Failed to fetch users: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []User
	for cursor.Next(ctx) {
		var user User
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode user: %v", err)
			return nil, err
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return nil, errors.New("no users found")
	}
	return users, nil
}

// UpdateUserProfile updates user information
func UpdateUserProfile(userID string, updatedUser User) error {
	collection := Client.Database("astroMatch").Collection("users")
	ctx := context.TODO()

	filter := bson.M{"id": userID}
	update := bson.M{"$set": updatedUser}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to update user profile: %v", err)
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}

// UpdateUserPreferences updates preferences for a user
func UpdateUserPreferences(userID string, newPrefs UserPreferences) error {
	collection := Client.Database("astromatch").Collection("preferences")
	ctx := context.TODO()

	filter := bson.M{"user_id": userID}
	update := bson.M{"$set": newPrefs}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to update user preferences: %v", err)
		return err
	}

	if result.MatchedCount == 0 {
		_, err := collection.InsertOne(ctx, newPrefs)
		if err != nil {
			log.Printf("Failed to insert new preferences: %v", err)
			return err
		}
	}
	return nil
}
