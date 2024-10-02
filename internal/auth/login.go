package auth

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var user User
	err := db.DB.QueryRow(
		"SELECT user_id, username, password_hash, role, created_at FROM users WHERE username = $1",
		creds.Username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := GenerateJWT(user.Username)
	if err != nil {
		log.Printf("Failed to generate JWT: %v\n", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
