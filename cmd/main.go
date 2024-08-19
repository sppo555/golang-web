package main

import (
	"log"
	"myproject/internal/handlers"
	"myproject/middleware"
	"net/http"
)

func main() {
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/balance", middleware.JWTMiddleware(handlers.BalanceHandler))
	http.HandleFunc("/price", handlers.PriceHandler)

	log.Println("Server starting on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
