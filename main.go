package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/org-45/escrow-agent/internal/db"
	"github.com/org-45/escrow-agent/internal/escrow"
)

func main() {
	db.InitDB()

	r := mux.NewRouter()

	r.HandleFunc("/escrow", escrow.CreateEscrowHandler).Methods("POST")
	r.HandleFunc("/escrow/{id}/release", escrow.ReleaseFundsHandler).Methods("POST")
	r.HandleFunc("/escrow/{id}/dispute", escrow.DisputeEscrowHandler).Methods("POST")

	r.HandleFunc("/escrow/pending", escrow.GetAllPendingEscrowsHandler).Methods("GET")
	r.HandleFunc("/escrow/disputed", escrow.GetAllDisputedEscrowsHandler).Methods("GET")

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
