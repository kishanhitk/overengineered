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
	w.Write([]byte("Welcome to the Greetings API on docker!"))
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
		count, err := rdb.Get(ctx, "greetings_count").Int()
		// log count
		fmt.Println("Current greetings count from Redis:", count, err)
		if err == redis.Nil {
			// Key doesn't exist in Redis, fetch from DB and set in Redis
			err = db.QueryRow("SELECT COUNT(*) FROM greetings").Scan(&count)
			if err != nil {
				http.Error(w, "Failed to get greetings count", http.StatusInternalServerError)
				return
			}
			// Set the count in Redis with an expiration time (e.g., 1 hour)
			err = rdb.Set(ctx, "greetings_count", count, time.Hour).Err()
			if err != nil {
				fmt.Printf("Failed to set Redis count: %v\n", err)
			}
		} else if err != nil {
			// If Redis fails, fallback to DB
			err = db.QueryRow("SELECT COUNT(*) FROM greetings").Scan(&count)
			if err != nil {
				http.Error(w, "Failed to get greetings count", http.StatusInternalServerError)
				return
			}
		}

		response := struct {
			GreetingsCount int `json:"current_greetings_count"`
		}{
			GreetingsCount: count,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
