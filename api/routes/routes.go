package routes

import (
	"database/sql"
	"net/http"

	"kishanhitk/overengineered/handlers"
	"kishanhitk/overengineered/middleware"
)

func SetupRoutes(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", middleware.EnableCORS(handlers.HomeHandler))
	mux.HandleFunc("/greetings", middleware.EnableCORS(handlers.GreetHandler(db)))
	mux.HandleFunc("/greetings/count", middleware.EnableCORS(handlers.GetGreetingsCountHandler(db)))

	return mux
}
``