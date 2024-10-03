package transactions

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/pkg/models"
	"log"
	"net/http"
)

type CreateTransactionRequest struct {
	SellerID int     `json:"seller_id"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"status,omitempty"`
}

func CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "buyer" {
		log.Printf("[ERROR] Unauthorized access attempt - invalid role or missing claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SellerID == 0 || req.Amount <= 0 {
		http.Error(w, "Seller ID and valid amount are required", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = "pending"
	}

	query := `
		INSERT INTO transactions (buyer_id, seller_id, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING transaction_id, buyer_id, seller_id, amount, status, created_at, updated_at
	`
	var transaction models.Transaction
	err := db.DB.QueryRowx(query, claims.UserID, req.SellerID, req.Amount, req.Status).StructScan(&transaction)
	if err != nil {
		log.Printf("[ERROR] Failed to create transaction: %v", err)
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}