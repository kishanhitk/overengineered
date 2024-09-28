package routes

import (
	"database/sql"
	"net/http"

	"kishanhitk/overengineered/handlers"
	"kishanhitk/overengineered/middleware"

	"github.com/go-redis/redis/v8"
)

func SetupRoutes(db *sql.DB, rdb *redis.Client) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", middleware.EnableCORS(handlers.HomeHandler))
	mux.HandleFunc("/greetings", middleware.EnableCORS(handlers.GreetHandler(db, rdb)))
	mux.HandleFunc("/greetings/count", middleware.EnableCORS(handlers.GetGreetingsCountHandler(db, rdb)))

	return mux
}
