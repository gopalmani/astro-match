package api

import (
	"astromatch/auth"
	"astromatch/matchmaking"
	"astromatch/user"
	"net/http"
)

func SetupRoutes() {
	//unprotected routes with middleware
	http.HandleFunc("/api/auth/login", CORSMiddleware(auth.LoginHandler))
	http.HandleFunc("/api/auth/signup", CORSMiddleware(auth.SignupHandler))
	http.HandleFunc("/api/auth/verify-otp", CORSMiddleware(auth.VerifyUserHandler))

	//protected routes with auth
	http.HandleFunc("/api/match/find", AuthMiddleware(matchmaking.MatchHandler))
	http.HandleFunc("/api/v1/users", AuthMiddleware(user.GetOrUpdateProfile))

}
