package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"kishanhitk/overengineered/models"

	"github.com/go-redis/redis/v8"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Welcome to the Greetings API!"))
}

func GreetHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req models.NameRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = db.Exec("INSERT INTO greetings (name) VALUES (?)", req.Name)
		if err != nil {
			http.Error(w, "Failed to store greeting", http.StatusInternalServerError)
			return
		}

		// Increment the count in Redis
		ctx := context.Background()
		_, err = rdb.Incr(ctx, "greetings_count").Result()
		if err != nil {
			// If Redis fails, log the error but don't stop the operation
			fmt.Printf("Failed to increment Redis count: %v\n", err)
		}

		response := models.GreetingResponse{
			Message: fmt.Sprintf("Hello, %s!", req.Name),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetGreetingsCountHandler(db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := context.Background()

		// Get data from Redis
		start := time.Now()
		redisCount, err := rdb.Get(ctx, "greetings_count").Int()
		redisElapsed := time.Since(start)
		fmt.Printf("Redis count: %d, error: %v, elapsed: %v\n", redisCount, err, redisElapsed)

		// Get data from DB
		start = time.Now()
		var dbCount int
		err = db.QueryRow("SELECT COUNT(*) FROM greetings").Scan(&dbCount)
		dbElapsed := time.Since(start)
		fmt.Printf("DB count: %d, error: %v, elapsed: %v\n", dbCount, err, dbElapsed)

		if err != nil {
			http.Error(w, "Failed to get greetings count", http.StatusInternalServerError)
			return
		}

		response := struct {
			GreetingsCount int `json:"current_greetings_count"`
		}{
			GreetingsCount: dbCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
