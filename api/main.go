package main

import (
	"fmt"
	"kishanhitk/overengineered/database"
	"kishanhitk/overengineered/routes"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
)

func main() {
	db := database.InitDB()
	defer db.Close()

	// Get Redis URL from environment variable
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:7379" // Default value if not set
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	defer rdb.Close()

	mux := routes.SetupRoutes(db, rdb)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
