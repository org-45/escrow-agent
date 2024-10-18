package escrow

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/pkg/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type DepositEscrowRequest struct {
	Amount float64 `json:"amount"`
}

func DepositEscrowHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "buyer" {
		log.Printf("[ERROR] Unauthorized access attempt - missing claims or incorrect role")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var req DepositEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction
	query := `
		SELECT transaction_id, buyer_id, seller_id, amount, status
		FROM transactions
		WHERE transaction_id = $1
	`
	err = db.DB.Get(&transaction, query, transactionID)
	if err != nil {
		log.Printf("[ERROR] Transaction not found with ID %d: %v", transactionID, err)
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if transaction.BuyerID != claims.UserID {
		log.Printf("[ERROR] Unauthorized access to transaction by userID %d", claims.UserID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Transaction %v", transaction)

	validStatuses := map[string]bool{"pending": true, "deposited": true, "in_progress": true}
	transactionStatus := strings.ToLower(strings.TrimSpace(transaction.Status))

	if !validStatuses[transactionStatus] {
		http.Error(w, "Transaction cannot be deposited into escrow in its current status", http.StatusBadRequest)
		return
	}
	insertQuery := `
		INSERT INTO escrow_accounts (transaction_id, escrowed_amount, status, created_at)
		VALUES ($1, $2, 'held', NOW())
		RETURNING escrow_id
	`
	var escrowID int
	err = db.DB.QueryRow(insertQuery, transactionID, req.Amount).Scan(&escrowID)
	if err != nil {
		log.Printf("[ERROR] Failed to deposit escrow for transaction ID %d: %v", transactionID, err)
		http.Error(w, "Failed to deposit escrow", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Escrow deposit successful",
		"escrow_id": escrowID,
	})
}

func ReleaseEscrowHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "buyer" {
		log.Printf("[ERROR] Unauthorized access attempt by userID %d with role %s", claims.UserID, claims.Role)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionIDStr := vars["id"]
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction
	err = db.DB.Get(&transaction, "SELECT * FROM transactions WHERE transaction_id = $1", transactionID)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if transaction.Status != "in_progress" && transaction.Status != "pending" {
		http.Error(w, "Cannot release funds for this transaction", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec("UPDATE escrow_accounts SET status = 'released' WHERE transaction_id = $1", transactionID)
	if err != nil {
		http.Error(w, "Failed to release funds from escrow", http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("UPDATE transactions SET status = 'completed', updated_at = NOW() WHERE transaction_id = $1", transactionID)
	if err != nil {
		http.Error(w, "Failed to update transaction status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Funds successfully released to the seller",
	})
}
