package escrow

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/org-45/escrow-agent/pkg/models"
)

// In-memory store for simplicity
var escrowStore = make(map[string]*models.Escrow)

func CreateEscrow(buyerID, sellerID string, amount float64, description string) (*models.Escrow, error) {
    escrow := &models.Escrow{
        ID:          generateID(),
        BuyerID:     buyerID,
        SellerID:    sellerID,
        Amount:      amount,
        Status:      models.StatusPending,
        CreatedAt:   time.Now(),
        Description: description,
    }

    escrowStore[escrow.ID] = escrow
    return escrow, nil
}

func generateID() string {
    return uuid.New().String()
}

func ReleaseFunds(escrowID string) error {
    escrow, exists := escrowStore[escrowID]
    if !exists {
        return errors.New("escrow not found")
    }

    if escrow.Status != models.StatusPending {
        return errors.New("escrow not in pending state")
    }

    now := time.Now()
    escrow.Status = models.StatusReleased
    escrow.ReleasedAt = &now

    // Implement payment gateway integration here to transfer funds to the seller

    return nil
}

func DisputeEscrow(escrowID string) error {
    escrow, exists := escrowStore[escrowID]
    if !exists {
        return errors.New("escrow not found")
    }

    if escrow.Status != models.StatusPending {
        return errors.New("escrow not in pending state")
    }

    now := time.Now()
    escrow.Status = models.StatusDisputed
    escrow.DisputedAt = &now

    // Implement logic to handle dispute

    return nil
}



// GetAllPendingEscrows returns a list of all escrows that are currently in the pending state.
func GetAllPendingEscrows() []*models.Escrow {
    var pendingEscrows []*models.Escrow
    for _, escrow := range escrowStore {
        if escrow.Status == models.StatusPending {
            pendingEscrows = append(pendingEscrows, escrow)
        }
    }
    return pendingEscrows
}

