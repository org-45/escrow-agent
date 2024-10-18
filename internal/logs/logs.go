package logs

import (
	"encoding/json"
	"escrow-agent/internal/db"
	"escrow-agent/pkg/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetTransactionLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionIDStr := vars["transaction_id"]
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var logs []models.TransactionLog
	query := "SELECT log_id, transaction_id, event_type, event_details, created_at FROM transaction_logs WHERE transaction_id = $1"
	err = db.DB.Select(&logs, query, transactionID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch logs for transaction ID %d: %v", transactionID, err)
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(logs)
}
