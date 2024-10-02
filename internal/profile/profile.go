package profile

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/pkg/models"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ProfileHandler has been called")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		log.Printf("[ERROR] Unauthorized access attempt - missing or invalid claims by kd")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := getUserByID(db.DB, claims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("[ERROR] Error encoding profile response: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func getUserByID(db *sqlx.DB, userID int) (*models.User, error) {
	var user models.User
	err := db.Get(&user, "SELECT user_id, username, role, created_at FROM users WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
