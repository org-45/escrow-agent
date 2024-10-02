package auth

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int       `db:"user_id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var jwtKey = []byte("my_secret_key")

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Retrieved credentials: %+v\n", creds)

	var storedCreds User
	err := db.DB.Get(&storedCreds, "SELECT user_id, username, password_hash, role FROM users WHERE username = $1", creds.Username)
	if err != nil {

		log.Printf("Error fetching user from DB for username %s: %v\n", creds.Username, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedCreds.PasswordHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials, password missmatch", http.StatusUnauthorized)
		return
	}

	claims := Claims{
		UserID:   storedCreds.ID,
		Username: storedCreds.Username,
		Role:     storedCreds.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
