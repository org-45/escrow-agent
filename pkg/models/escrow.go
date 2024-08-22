package models

import "time"

type EscrowStatus string

const (
    StatusPending   EscrowStatus = "pending"
    StatusReleased  EscrowStatus = "released"
    StatusDisputed  EscrowStatus = "disputed"
)

type Escrow struct {
    ID          string       `json:"id"`
    BuyerID     string       `json:"buyer_id"`
    SellerID    string       `json:"seller_id"`
    Amount      float64      `json:"amount"`
    Status      EscrowStatus `json:"status"`
    CreatedAt   time.Time    `json:"created_at"`
    ReleasedAt  *time.Time   `json:"released_at,omitempty"`
    DisputedAt  *time.Time   `json:"disputed_at,omitempty"`
    Description string       `json:"description"`
}
