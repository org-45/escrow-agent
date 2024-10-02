package profile

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/pkg/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
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

type UpdateProfileRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
}

func ProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ProfileUpdateHandler has been called")

	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		log.Printf("[ERROR] Unauthorized access attempt - missing or invalid claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var updateReq UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	query := "UPDATE users SET "
	var fields []string
	var args []interface{}
	argCount := 1

	if updateReq.Username != "" {
		fields = append(fields, "username = $"+strconv.Itoa(argCount))
		args = append(args, updateReq.Username)
		argCount++
	}

	if updateReq.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateReq.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[ERROR] Failed to hash password: %v", err)
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		fields = append(fields, "password_hash = $"+strconv.Itoa(argCount))
		args = append(args, hashedPassword)
		argCount++
	}

	if updateReq.Role != "" {
		fields = append(fields, "role = $"+strconv.Itoa(argCount))
		args = append(args, updateReq.Role)
		argCount++
	}

	if len(fields) == 0 {
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	query += strings.Join(fields, ", ")

	query += " WHERE user_id = $" + strconv.Itoa(argCount)
	args = append(args, claims.UserID)

	log.Printf("Executing update query: %s with args: %+v", query, args)
	_, err := db.DB.Exec(query, args...)
	if err != nil {
		log.Printf("[ERROR] Failed to update user profile for userID %d: %v", claims.UserID, err)
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Profile updated successfully"}`))
}
