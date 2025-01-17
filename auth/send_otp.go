package auth

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/go-resty/resty/v2"
	gomail "gopkg.in/gomail.v2"
)

// OTP store (replace with Redis or MongoDB)
var otpStore = make(map[string]string)

// Generate a random 6-digit OTP
func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return strconv.Itoa(rand.Intn(999999))
}

// Send OTP via email
func SendOTPViaEmail(email, otp string) error {
	smtpHost := "smtp-relay.sendinblue.com"
	smtpPort := 587
	smtpUsername := "astromatch.business@gmail.com"
	smtpPassword := "xsmtpsib-246e9bb954eaf1c2167000bb32705fd43a502f77c504ab170af4dd0682677560-QTYc3GrvaXCKAEnM"

	msg := gomail.NewMessage()
	msg.SetHeader("From", smtpUsername)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Astromatch Signup OTP")
	msg.SetBody("text/plain", "Your OTP is: "+otp)

	mailer := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	if err := mailer.DialAndSend(msg); err != nil {
		return err
	}
	log.Printf("OTP %s sent successfully to %s", otp, email)
	return nil
}

// Store OTP temporarily (expires after 5 minutes)
func CacheOTP(email, otp string) {
	otpStore[email] = otp
	go func() {
		time.Sleep(15 * time.Minute)
		delete(otpStore, email)
	}()
}

// Validate OTP (Check if valid and matches)
func ValidateOTP(email, otp string) bool {
	if storedOTP, exists := otpStore[email]; exists {
		return storedOTP == otp
	}
	return false
}

func SendOTPViaPhone(phoneNumber, otp string) error {
	client := resty.New()

	response, err := client.R().
		SetHeader("accept", "application/json").
		SetHeader("api-key", "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").
		SetHeader("content-type", "application/json").
		SetBody(map[string]interface{}{
			"sender":    "AstroMatch",
			"recipient": phoneNumber,
			"content":   "Your Astromatch OTP is: " + otp,
		}).
		Post("https://api.brevo.com/v3/transactionalSMS/sms")

	if err != nil {
		log.Printf("Failed to send OTP to %s: %v", phoneNumber, err)
		return err
	}

	log.Printf("OTP %s sent successfully to %s. Response: %s", otp, phoneNumber, response.String())
	return nil
}

func VerifyOTPHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) {
	var req struct {
		Email  string `json:"email,omitempty"`
		Mobile string `json:"mobile,omitempty"`
		OTP    string `json:"otp"`
	}

	// Decode the incoming request
	json.NewDecoder(r.Body).Decode(&req)

	// Determine if it's email or phone verification
	var identifier string
	var filter bson.M

	if req.Email != "" {
		identifier = req.Email
		filter = bson.M{"email": req.Email}
	} else if req.Mobile != "" {
		identifier = req.Mobile
		filter = bson.M{"mobile": req.Mobile}
	} else {
		http.Error(w, "Invalid request. Email or mobile is required.", http.StatusBadRequest)
		return
	}

	// Validate the OTP
	if ValidateOTP(identifier, req.OTP) {
		collection := client.Database("astromatch").Collection("users")
		_, err := collection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"isVerified": true}})
		if err != nil {
			http.Error(w, "Failed to verify user", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "User verified successfully"})
	} else {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
	}
}
