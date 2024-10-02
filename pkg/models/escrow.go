package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type EscrowStatus string

const (
	StatusPending  EscrowStatus = "pending"
	StatusReleased EscrowStatus = "released"
	StatusDisputed EscrowStatus = "disputed"
)

type Escrow struct {
	ID          string       `db:"id"`
	BuyerID     string       `db:"buyer_id"`
	SellerID    string       `db:"seller_id"`
	Amount      float64      `db:"amount"`
	Status      EscrowStatus `db:"status"`
	CreatedAt   time.Time    `db:"created_at"`
	ReleasedAt  *time.Time   `db:"released_at"`
	DisputedAt  *time.Time   `db:"disputed_at"`
	Description string       `db:"description"`
}

type User struct {
	ID        int       `db:"user_id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Password  string    `db:"password_hash" json:"-"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Transaction struct {
	ID        int       `db:"transaction_id" json:"id"`
	BuyerID   int       `db:"buyer_id" json:"buyer_id"`
	SellerID  int       `db:"seller_id" json:"seller_id"`
	Amount    float64   `db:"amount" json:"amount"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type EscrowAccount struct {
	ID            int       `db:"escrow_id" json:"id"`
	TransactionID int       `db:"transaction_id" json:"transaction_id"`
	Amount        float64   `db:"escrowed_amount" json:"escrowed_amount"`
	Status        string    `db:"status" json:"status"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
