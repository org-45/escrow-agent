package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"

	"escrow-agent/internal/auth"
	"escrow-agent/internal/db"
	"escrow-agent/internal/escrow"
	"escrow-agent/internal/middleware"
	"escrow-agent/internal/profile"
	"escrow-agent/internal/transactions"

	"github.com/google/wire"
)

func InitializeDB() (*sqlx.DB, func(), error) {
	db.InitDB()
	cleanup := func() {
		db.DB.Close()
	}
	return db.DB, cleanup, nil
}

func InitializeRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	r.HandleFunc("/register", auth.RegisterHandler).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.JWTAuthMiddleware)

	api.HandleFunc("/profile", profile.ProfileHandler).Methods("GET")
	api.HandleFunc("/profile", profile.ProfileUpdateHandler).Methods("PUT")
	api.HandleFunc("/transactions", transactions.CreateTransactionHandler).Methods("POST")
	api.HandleFunc("/transactions", transactions.GetTransactionsHandler).Methods("GET")
	api.HandleFunc("/transactions/{id}", transactions.GetTransactionHandler).Methods("GET")
	api.HandleFunc("/transactions/{id}/fulfill", transactions.FulfillTransactionHandler).Methods("PUT")
	api.HandleFunc("/transactions/{id}/confirm", transactions.ConfirmDeliveryHandler).Methods("PUT")
	api.HandleFunc("/escrow/{id}/deposit", escrow.DepositEscrowHandler).Methods("POST")

	return r
}

func InitializeCORSHandler(r *mux.Router) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})
	return c.Handler(r)
}

func InitializeServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

var WireSet = wire.NewSet(InitializeDB, InitializeRouter, InitializeCORSHandler, InitializeServer)
