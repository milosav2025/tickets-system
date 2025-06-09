package main

import (
	"log"
	"net/http"

	"goticketsistem/db"
	"goticketsistem/handlers"
)

func main() {
	dsn := "user=postgres password=misa dbname=tickets&system sslmode=disable"
	dbManager, err := db.NewDBManager(dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbManager.Close()

	handler := handlers.NewTicketHandler(dbManager)
	mux := http.NewServeMux()                       // Kreiraj novi ServeMux
	mux.HandleFunc("/ticket", handler.HandleTicket) // Registrovani handler
	log.Println("Server starting on :8080 at 02:30 PM CEST, June 07, 2025...")
	if err := http.ListenAndServe(":8080", mux); err != nil { // Koristi mux
		log.Fatal("Server failed:", err)
	}
}
