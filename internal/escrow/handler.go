package escrow

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func CreateEscrowHandler(w http.ResponseWriter, r *http.Request) {
	var escrowRequest struct {
		BuyerID     string  `json:"BuyerID"`
		SellerID    string  `json:"SellerID"`
		Amount      float64 `json:"Amount"`
		Description string  `json:"Description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&escrowRequest); err != nil {
		log.Printf("Failed to decode request body: %v\n", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if escrowRequest.BuyerID == "" || escrowRequest.SellerID == "" || escrowRequest.Amount <= 0 {
		log.Printf("Validation error: missing required fields")
		http.Error(w, "Missing required fields: buyer_id, seller_id, or amount", http.StatusBadRequest)
		return
	}

	log.Printf("Received create escrow request: %+v\n", escrowRequest)

	escrow, err := CreateEscrow(escrowRequest.BuyerID, escrowRequest.SellerID, escrowRequest.Amount, escrowRequest.Description)
	if err != nil {
		log.Printf("Failed to create escrow: %v\n", err)
		http.Error(w, "Failed to create escrow", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(escrow); err != nil {
		log.Printf("Failed to encode response: %v\n", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Escrow created successfully: %+v\n", escrow)
}

func ReleaseFundsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	escrowID := vars["id"]

	if err := ReleaseFunds(escrowID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DisputeEscrowHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	escrowID := vars["id"]

	if err := DisputeEscrow(escrowID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetAllPendingEscrowsHandler(w http.ResponseWriter, r *http.Request) {
	pendingEscrows, err := GetAllPendingEscrows()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pendingEscrows)
}

func GetAllDisputedEscrowsHandler(w http.ResponseWriter, r *http.Request) {
	disputedEscrows, err := GetAllDisputedEscrows()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disputedEscrows)
}
