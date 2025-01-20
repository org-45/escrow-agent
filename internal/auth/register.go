package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"escrow-agent/internal/db"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterResponse struct {
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func validateInput(req RegisterRequest) error {

	if req.Username == "" || req.Password == "" || req.Role == "" {
		return fmt.Errorf("all required parameters not passed")
	}

	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if req.Role != "buyer" && req.Role != "seller" && req.Role != "admin" {
		return fmt.Errorf("invalid role")
	}

	return nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("RegisterHandler called")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validateInput(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	log.Printf("Registering user: %v", req.Username)

	var createdAt time.Time
	err = db.DB.QueryRow(
		"INSERT INTO users (username, password_hash, role, created_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP) RETURNING created_at",
		req.Username, hashedPassword, req.Role,
	).Scan(&createdAt)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgerrcode.UniqueViolation {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		Message:   "User registered successfully",
		CreatedAt: createdAt,
	})
}
