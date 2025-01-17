package main

import (
	"astromatch/api"
	"astromatch/db"
	"fmt"
	"log"
	"net/http"
)

func main() {

	//initialise db connection
	db.InitDB()

	// Setup API routes
	api.SetupRoutes()

	// Root endpoint to confirm service is running
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, " 💫 🪐 🔮 ✨ Welcome to ⋆｡ﾟ☁︎｡⋆｡ ﾟ☾ ﾟ｡⋆ ASTROTALK ⋆｡ﾟ☁︎｡⋆｡ ﾟ☾ ﾟ｡⋆ ✨ 🪐 🔮  💫 ")
	})

	// Log that services are running
	log.Println("🌟 AstroMatch services are running on port 8080 🌟")

	// Start the server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
