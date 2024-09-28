package main

import (
	"fmt"
	"kishanhitk/overengineered/database"
	"kishanhitk/overengineered/routes"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

func main() {
	db := database.InitDB()
	defer db.Close()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:7379",
	})
	defer rdb.Close()

	mux := routes.SetupRoutes(db, rdb)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
