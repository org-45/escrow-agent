package admin

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/pkg/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "admin" {
		log.Printf("[ERROR] Unauthorized access attempt by userID %d with role %s", claims.UserID, claims.Role)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var users []models.User
	err := db.DB.Select(&users, "SELECT user_id, username, role, created_at FROM users")
	if err != nil {
		log.Printf("[ERROR] Failed to fetch users: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func GetUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "admin" {
		log.Printf("[ERROR] Unauthorized access attempt by userID %d with role %s", claims.UserID, claims.Role)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	err = db.DB.Get(&user, "SELECT user_id, username, role, created_at FROM users WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch user with ID %d: %v", userID, err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "admin" {
		log.Printf("[ERROR] Unauthorized access attempt by userID %d with role %s", claims.UserID, claims.Role)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var transactions []models.Transaction
	err := db.DB.Select(&transactions, "SELECT transaction_id, buyer_id, seller_id, amount, status, created_at FROM transactions")
	if err != nil {
		log.Printf("[ERROR] Failed to fetch transactions: %v", err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}
