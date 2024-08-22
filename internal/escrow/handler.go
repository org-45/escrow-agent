package escrow

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func CreateEscrowHandler(w http.ResponseWriter, r *http.Request) {
    var escrowRequest struct {
        BuyerID     string  `json:"buyer_id"`
        SellerID    string  `json:"seller_id"`
        Amount      float64 `json:"amount"`
        Description string  `json:"description"`
    }

    if err := json.NewDecoder(r.Body).Decode(&escrowRequest); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    escrow, err := CreateEscrow(escrowRequest.BuyerID, escrowRequest.SellerID, escrowRequest.Amount, escrowRequest.Description)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(escrow)
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


// GetAllPendingEscrowsHandler returns all escrows that are currently in the pending state.
func GetAllPendingEscrowsHandler(w http.ResponseWriter, r *http.Request) {
    pendingEscrows := GetAllPendingEscrows()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(pendingEscrows)
}


