package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"kishanhitk/overengineered/models"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Write([]byte("Welcome to the Greetings API!"))
}

func GreetHandler(db *sql.DB) http.HandlerFunc {
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

		response := models.GreetingResponse{
			Message: fmt.Sprintf("Hello, %s!", req.Name),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetGreetingsCountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM greetings").Scan(&count)
		if err != nil {
			http.Error(w, "Failed to get greetings count", http.StatusInternalServerError)
			return
		}

		response := struct {
			GreetingsCount int `json:"greetings_count"`
		}{
			GreetingsCount: count,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
