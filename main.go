package main

import (
	"log"
	"net/http"

	"escrow-agent/internal/auth"
	"escrow-agent/internal/db"
	"escrow-agent/internal/escrow"
	"escrow-agent/internal/fileupload"
	"escrow-agent/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	db.InitDB()

	r := mux.NewRouter()

	//public routes
	r.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	r.HandleFunc("/register", auth.RegisterHandler).Methods("POST")

	//protected routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.JWTAuthMiddleware)
	api.HandleFunc("/escrow", escrow.CreateEscrowHandler).Methods("POST")
	api.HandleFunc("/escrow/{id}/release", escrow.ReleaseFundsHandler).Methods("POST")
	api.HandleFunc("/escrow/{id}/dispute", escrow.DisputeEscrowHandler).Methods("POST")

	api.HandleFunc("/escrow/pending", escrow.GetAllPendingEscrowsHandler).Methods("GET")
	api.HandleFunc("/escrow/disputed", escrow.GetAllDisputedEscrowsHandler).Methods("GET")

	api.HandleFunc("/upload", fileupload.UploadHandler).Methods("POST")

	api.HandleFunc("/customer", escrow.CreateCustomerHandler).Methods("POST")

	// setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// wrap the router with the CORS middleware
	handler := c.Handler(r)

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
