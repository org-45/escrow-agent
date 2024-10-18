package admin

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/pkg/models"
	"log"
	"net/http"
)

// GetUsersHandler returns a list of all users (admin-only)
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is an admin
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "admin" {
		log.Printf("[ERROR] Unauthorized access attempt by userID %d with role %s", claims.UserID, claims.Role)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch all users from the database
	var users []models.User
	err := db.DB.Select(&users, "SELECT user_id, username, role, created_at FROM users")
	if err != nil {
		log.Printf("[ERROR] Failed to fetch users: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	// Return the list of users in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
