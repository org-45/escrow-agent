package escrow

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/org-45/escrow-agent/internal/db"
	"github.com/org-45/escrow-agent/pkg/models"
)

type Customer struct {
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name,omitempty"`
	LastName    string `json:"last_name"`
	Line1       string `json:"line1"`
	Line2       string `json:"line2,omitempty"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	PostCode    string `json:"post_code"`
	PhoneNumber string `json:"phone_number"`
}

func CreateCustomer(customer Customer) error {
	query := `
		INSERT INTO customers (email, first_name, middle_name, last_name, line1, line2, city, state, country, post_code, phone_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := db.DB.Exec(query, customer.Email, customer.FirstName, customer.MiddleName, customer.LastName, customer.Line1, customer.Line2, customer.City, customer.State, customer.Country, customer.PostCode, customer.PhoneNumber)
	return err
}

func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var customer Customer
	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = CreateCustomer(customer)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Customer could not be created", http.StatusInternalServerError)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Customer created successfully"})
}

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
