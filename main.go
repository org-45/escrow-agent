package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/org-45/escrow-agent/internal/db"
	"github.com/org-45/escrow-agent/internal/escrow"
	"github.com/rs/cors"
)

func main() {
	db.InitDB()

	r := mux.NewRouter()

	r.HandleFunc("/escrow", escrow.CreateEscrowHandler).Methods("POST")
	r.HandleFunc("/escrow/{id}/release", escrow.ReleaseFundsHandler).Methods("POST")
	r.HandleFunc("/escrow/{id}/dispute", escrow.DisputeEscrowHandler).Methods("POST")

	r.HandleFunc("/escrow/pending", escrow.GetAllPendingEscrowsHandler).Methods("GET")
	r.HandleFunc("/escrow/disputed", escrow.GetAllDisputedEscrowsHandler).Methods("GET")

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// wrap the router with the CORS middleware
	handler := c.Handler(r)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
