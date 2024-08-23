package models

import "time"

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
