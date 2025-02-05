package transactions

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/internal/middleware"
	"escrow-agent/pkg/models"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CreateTransactionRequest struct {
	SellerID int     `json:"seller_id"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"transaction_status,omitempty"`
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

	//validation
	if req.SellerID == 0 || req.Amount <= 0 {
		http.Error(w, "Seller ID and valid amount are required", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = "pending"
	}

	query := `
		INSERT INTO transactions (buyer_id, seller_id, amount, transaction_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING transaction_id, buyer_id, seller_id, amount, transaction_status, created_at, updated_at
	`
	var transaction models.Transaction
	err := db.DB.QueryRowx(query, claims.UserID, req.SellerID, req.Amount, req.Status).StructScan(&transaction)
	if err != nil {
		log.Printf("[ERROR] Failed to create transaction: %v", err)
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}

	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal transaction to JSON: %v", err)
		transactionJSON = []byte("{}")
	}

	logQuery := `
		INSERT INTO transaction_logs (transaction_id, event_type, event_details, created_at)
		VALUES ($1, 'TransactionCreated', $2, NOW())
	`
	eventDetails := fmt.Sprintf("Transaction created by buyer: %s", string(transactionJSON))
	_, err = db.DB.Exec(logQuery, transaction.TransactionID, eventDetails)
	if err != nil {
		log.Printf("[ERROR] Failed to insert log for transaction ID %d: %v", transaction.TransactionID, err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		log.Printf("[ERROR] Unauthorized access attempt - missing or invalid claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var transactions []models.Transaction
	query := `
		SELECT transaction_id, buyer_id, seller_id, amount, transaction_status, created_at, updated_at
		FROM transactions
		WHERE buyer_id = $1 OR seller_id = $1
		ORDER BY created_at DESC
	`
	err := db.DB.Select(&transactions, query, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch transactions for userID %d: %v", claims.UserID, err)
		http.Error(w, "Failed to fetch transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		log.Printf("[ERROR] Error encoding transactions response: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func GetTransactionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok {
		log.Printf("[ERROR] Unauthorized access attempt - missing or invalid claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction
	query := `
		SELECT transaction_id, buyer_id, seller_id, amount, transaction_status, created_at, updated_at
		FROM transactions
		WHERE transaction_id = $1
	`
	err = db.DB.Get(&transaction, query, transactionID)
	if err != nil {
		log.Printf("[ERROR] Transaction not found with ID %d: %v", transactionID, err)
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if transaction.BuyerID != claims.UserID && transaction.SellerID != claims.UserID {
		log.Printf("[ERROR] Unauthorized access to transaction by userID %d", claims.UserID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(transaction); err != nil {
		log.Printf("[ERROR] Error encoding transaction response: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func FulfillTransactionHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*middleware.Claims)
	if !ok || claims.Role != "seller" {
		log.Printf("[ERROR] Unauthorized access attempt - invalid role or missing claims")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	transactionID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction
	query := `
		SELECT transaction_id, buyer_id, seller_id, transaction_status
		FROM transactions
		WHERE transaction_id = $1
	`
	err = db.DB.Get(&transaction, query, transactionID)
	if err != nil {
		log.Printf("[ERROR] Transaction not found with ID %d: %v", transactionID, err)
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if transaction.SellerID != claims.UserID {
		log.Printf("[ERROR] Unauthorized access to transaction by userID %d", claims.UserID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if transaction.Status != "pending" {
		http.Error(w, "Transaction cannot be fulfilled in its current status", http.StatusBadRequest)
		return
	}

	updateQuery := `
		UPDATE transactions
		SET transaction_status = 'deposited', updated_at = NOW()
		WHERE transaction_id = $1
	`
	_, err = db.DB.Exec(updateQuery, transactionID)
	if err != nil {
		log.Printf("[ERROR] Failed to update transaction status for ID %d: %v", transactionID, err)
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal transaction to JSON: %v", err)
		transactionJSON = []byte("{}")
	}

	logQuery := `
		INSERT INTO transaction_logs (transaction_id, event_type, event_details, created_at)
		VALUES ($1, 'TransactionFulfilled', $2, NOW())
	`
	eventDetails := fmt.Sprintf("Transaction fulfilled by seller: %s", string(transactionJSON))
	_, err = db.DB.Exec(logQuery, transactionID, eventDetails)
	if err != nil {
		log.Printf("[ERROR] Failed to insert log for transaction ID %d: %v", transactionID, err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Transaction marked as fulfilled"})
}

func ConfirmDeliveryHandler(w http.ResponseWriter, r *http.Request) {
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

	var transaction models.Transaction
	query := `
		SELECT transaction_id, buyer_id, seller_id, transaction_status
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

	if transaction.Status != "deposited" {
		http.Error(w, "Transaction cannot be confirmed in its current status", http.StatusBadRequest)
		return
	}

	updateQuery := `
		UPDATE transactions
		SET transaction_status = 'completed', updated_at = NOW()
		WHERE transaction_id = $1
	`
	_, err = db.DB.Exec(updateQuery, transactionID)
	if err != nil {
		log.Printf("[ERROR] Failed to update transaction status for ID %d: %v", transactionID, err)
		http.Error(w, "Failed to update transaction", http.StatusInternalServerError)
		return
	}

	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal transaction to JSON: %v", err)
		transactionJSON = []byte("{}")
	}

	logQuery := `
		INSERT INTO transaction_logs (transaction_id, event_type, event_details, created_at)
		VALUES ($1, 'TransactionConfirmed', $2, NOW())
	`
	eventDetails := fmt.Sprintf("Transaction confirmed by buyer: %s", string(transactionJSON))
	_, err = db.DB.Exec(logQuery, transactionID, eventDetails)
	if err != nil {
		log.Printf("[ERROR] Failed to insert log for transaction ID %d: %v", transactionID, err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Transaction confirmed by buyer"})
}
