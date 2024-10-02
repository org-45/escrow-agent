package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"escrow-agent/internal/auth"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/internal/profile"

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
