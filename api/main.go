package main

import (
	"fmt"
	"log"
	"net/http"

	"kishanhitk/overengineered/database"
	"kishanhitk/overengineered/routes"
)

func main() {
	db := database.InitDB()
	defer db.Close()

	mux := routes.SetupRoutes(db)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
