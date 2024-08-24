package escrow

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/org-45/escrow-agent/internal/db"
	"github.com/org-45/escrow-agent/pkg/models"
)

// create an escrow
func CreateEscrow(buyerID string, sellerID string, amount float64, description string) (*models.Escrow, error) {

	escrow := &models.Escrow{
		ID:          generateID(),
		BuyerID:     buyerID,
		SellerID:    sellerID,
		Amount:      amount,
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
		Description: description,
	}

	log.Printf("Received Escrow Data: %+v\n", escrow)

	query := `INSERT INTO escrows (id, buyer_id, seller_id, amount, status, created_at, description)
							VALUES ($1, $2, $3, $4, $5, $6, $7)	`
	_, err := db.DB.Exec(query, escrow.ID, escrow.BuyerID, escrow.SellerID, escrow.Amount, escrow.Status, escrow.CreatedAt, escrow.Description)

	if err != nil {
		return nil, err
	}
	return escrow, nil
}

// release funds from escrow
func ReleaseFunds(escrowID string) error {
	escrow, err := getEscrowByID(escrowID)

	if err != nil {
		return err
	}

	if escrow.Status != models.StatusPending {
		return errors.New("escrow not in pending state")
	}

	now := time.Now()
	escrow.Status = models.StatusReleased
	escrow.ReleasedAt = &now

	// update the escrow status in the database
	query := `UPDATE escrows SET status = $1, released_at = $2 WHERE id = $3`
	_, err = db.DB.Exec(query, escrow.Status, escrow.ReleasedAt, escrowID)
	if err != nil {
		return err
	}

	// implement payment gateway integration here to transfer funds to the seller

	return nil
}

// dispute escrow
func DisputeEscrow(escrowID string) error {
	// check if the escrow exists and is in the pending state
	escrow, err := getEscrowByID(escrowID)
	if err != nil {
		return err
	}

	if escrow.Status != models.StatusPending {
		return errors.New("escrow not in pending state")
	}

	now := time.Now()
	escrow.Status = models.StatusDisputed
	escrow.DisputedAt = &now

	// update the escrow status in the database
	query := `UPDATE escrows SET status = $1, disputed_at = $2 WHERE id = $3`
	_, err = db.DB.Exec(query, escrow.Status, escrow.DisputedAt, escrowID)
	if err != nil {
		return err
	}

	return nil
}

// GetAllPendingEscrows returns a list of all escrows that are currently in the pending state.
func GetAllPendingEscrows() ([]*models.Escrow, error) {
	var pendingEscrows []*models.Escrow
	query := `SELECT * FROM escrows WHERE status = $1`
	err := db.DB.Select(&pendingEscrows, query, models.StatusPending)

	if err != nil {
		return nil, err
	}
	return pendingEscrows, nil
}

// GetAllDisputedEscrows returns a list of all escrows that are currently in the disputed state.
func GetAllDisputedEscrows() ([]*models.Escrow, error) {
	var disputedEscrows []*models.Escrow
	query := `SELECT * FROM escrows WHERE status = $1`
	err := db.DB.Select(&disputedEscrows, query, models.StatusDisputed)

	if err != nil {
		return nil, err
	}
	return disputedEscrows, nil
}

func generateID() string {
	return uuid.New().String()
}

func getEscrowByID(escrowID string) (*models.Escrow, error) {
	var escrow models.Escrow
	query := `SELECT * FROM escrows WHERE id = $1`
	err := db.DB.Get(&escrow, query, escrowID)

	if err != nil {
		return nil, errors.New("escrow not found")
	}
	return &escrow, nil
}
