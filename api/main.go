package main

import (
	"fmt"
	"kishanhitk/overengineered/database"
	"kishanhitk/overengineered/handlers"
	"kishanhitk/overengineered/middleware"
	"log"
	"net/http"
)

func main() {
	db := database.InitDB()
	defer db.Close()

	http.HandleFunc("/", middleware.EnableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!1234"))
	}))
	http.HandleFunc("/greet", middleware.EnableCORS(handlers.GreetHandler(db)))
	http.HandleFunc("/total-greetings", middleware.EnableCORS(handlers.GetTotalGreetingsHandler(db)))

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
