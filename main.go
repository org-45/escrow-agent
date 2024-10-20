package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"escrow-agent/internal/admin"
	"escrow-agent/internal/auth"
	"escrow-agent/internal/db"
	"escrow-agent/internal/escrow"
	"escrow-agent/internal/logs"
	"escrow-agent/internal/middleware"
	"escrow-agent/internal/profile"
	"escrow-agent/internal/transactions"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	db.InitDB()

	defer db.DB.Close()

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
	api.HandleFunc("/escrow/{id}/release", escrow.ReleaseEscrowHandler).Methods("PUT")

	api.HandleFunc("/admin/users", admin.GetUsersHandler).Methods("GET")
	api.HandleFunc("/admin/users/{id}", admin.GetUserByIDHandler).Methods("GET")
	api.HandleFunc("/admin/transactions", admin.GetTransactionsHandler).Methods("GET")

	api.HandleFunc("/logs/{transaction_id}", logs.GetTransactionLogsHandler).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	handler := c.Handler(r)

	srv := &http.Server{
		Handler:      handler,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)

	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
