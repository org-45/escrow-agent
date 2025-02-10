// +build dispute_buyer_wins

// Dispute - Damaged Goods (Resolved in Buyer’s Favor)
// 1. Buyer (Georgia) buys from Seller (Harry) for $300.
// 2. Georgia funds escrow (funded).
// 3. Harry ships product, but it arrives damaged.
// 4. Georgia opens a dispute (open) for “Damaged Goods.”
// 5. Admin reviews and decides in Georgia’s favor -> dispute resolved with refund to buyer.
// 6. Escrow set to cancelled, released back to buyer, transaction cancelled.
// Key Points: Dispute arises, resolved in buyer’s favor, buyer gets the funds back.

package dispute_buyer_wins_test

import (
	"database/sql"
	"os"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

func TestDisputeBuyerWinsFlow(t *testing.T) {
	db := connectDB(t)
	defer db.Close()

	// Ensure a clean slate
	cleanTables(t, db)

	// Create users
	buyerID := createUser(t, db, "Georgia", "some_hashed_password", "buyer")
	sellerID := createUser(t, db, "Harry", "some_hashed_password", "seller")
	adminID := createUser(t, db, "AdminUser", "some_hashed_password", "admin") // Added an admin user
	t.Logf("Users created buyer=%s, seller=%s, admin=%s", buyerID, sellerID, adminID)

	// Georgia initiates a transaction
	transactionID := createTransaction(t, db, buyerID, sellerID, 300.00)
	t.Logf("Transaction created (pending): %s", transactionID)

	// Create the initial payment record
	paymentID := createPayment(t, db, transactionID, buyerID, 300.00)
	t.Logf("Payment created: %s", paymentID)

	// Update transaction with payment id
	updateTransactionWithPayment(t, db, transactionID, paymentID)
	t.Logf("Transaction %s updated with payment %s", transactionID, paymentID)

	// Georgia funds escrow
	fundEscrow(t, db, transactionID, 300.00)
	t.Logf("Escrow funded with $300 for transaction: %s", transactionID)

	//Harry ships product
	addTransactionLog(t, db, transactionID, "SHIPPING", "Harry shipped the item.")
	t.Log("Shipping log added.")

	// Simulate item arriving damaged
	addTransactionLog(t, db, transactionID, "DELIVERY", "Item delivered, but reported as damaged by buyer.")
	t.Log("Delivery and damage report log added.")

	//Georgia opens dispute
	disputeID := openDispute(t, db, transactionID, buyerID, "Damaged Goods Received") // Georgia opens dispute
	t.Logf("Dispute opened with ID: %s", disputeID)

	//Admin resolves dispute in buyer favor
	resolveDispute(t, db, disputeID, adminID, "Resolved in Buyer's Favor")

	//Cancel transaction and escrow + refund
	cancelTransactionAndRefund(t, db, transactionID, buyerID) //Refactored refund process
	t.Logf("Transaction %s cancelled and refund initiated", transactionID)

	// Verify the final state of all the processes
	verifyDisputeResolution(t, db, transactionID, buyerID, disputeID, t)

	t.Log("Dispute resolved in buyer's favor test completed successfully!")
}

func connectDB(t *testing.T) *sql.DB {
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "postgres")
	pass := getEnv("TEST_DB_PASS", "postgres")
	name := getEnv("TEST_DB_NAME", "test_escrow_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping DB: %v", err)
	}

	return db
}

func openDispute(t *testing.T, db *sql.DB, transactionID, raisedBy, reason string) string {
	var disputeID string
	query := `
		INSERT INTO disputes (transaction_id, raised_by, reason, dispute_status)
		VALUES ($1, $2, $3, 'open')
		RETURNING dispute_id;
	`
	if err := db.QueryRow(query, transactionID, raisedBy, reason).Scan(&disputeID); err != nil {
		t.Fatalf("openDispute failed: %v", err)
	}

	queryUpdateTransaction := `
		UPDATE transactions
		SET dispute_id = $1
		WHERE transaction_id = $2
	`
	if _, err := db.Exec(queryUpdateTransaction, disputeID, transactionID); err != nil {
		t.Fatalf("updateTransactionWithDispute failed: %v", err)
	}

	return disputeID
}

func resolveDispute(t *testing.T, db *sql.DB, disputeID, resolvedBy, resolution string) {
	query := `
		UPDATE disputes
		SET dispute_status = 'resolved',
		    resolution = $2,
		    resolved_by = $3,
		    resolved_at = NOW()
		WHERE dispute_id = $1
	`
	if _, err := db.Exec(query, disputeID, resolution, resolvedBy); err != nil {
		t.Fatalf("resolveDispute failed: %v", err)
	}
}

func cancelTransactionAndRefund(t *testing.T, db *sql.DB, transactionID, buyerID string) {
	//First, update transaction and escrow account
	queryTransaction := `
		UPDATE transactions
		SET transaction_status = 'cancelled',
		    updated_at = NOW()
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(queryTransaction, transactionID); err != nil {
		t.Fatalf("cancelTransaction (transaction status) failed: %v", err)
	}

	queryEscrow := `
		UPDATE escrow_accounts
		SET escrow_status = 'cancelled'
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(queryEscrow, transactionID); err != nil {
		t.Fatalf("cancelTransaction (escrow status) failed: %v", err)
	}

	// Now refund by updating the exiting payment
	queryPayment := `
		UPDATE payments
		SET payment_status = 'refunded',
		    processed_at = NOW()
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(queryPayment, transactionID); err != nil {
		t.Fatalf("cancelTransaction (payment status) failed: %v", err)
	}

	addTransactionLog(t, db, transactionID, "CANCELLATION_REQUESTED", "Transaction cancelled and refund initiated after dispute resolution.")
}

func verifyDisputeResolution(t *testing.T, db *sql.DB, transactionID, buyerID, disputeID string, tb testing.TB) {
	//Verify that dispute is resolved, transaction is cancelled and escrow is cancelled

	var transactionStatus string
	err := db.QueryRow("SELECT transaction_status FROM transactions WHERE transaction_id = $1", transactionID).Scan(&transactionStatus)
	if err != nil {
		t.Fatalf("Error getting transaction status: %v", err)
	}
	if transactionStatus != "cancelled" {
		t.Errorf("Expected transaction status to be 'cancelled', but got '%s'", transactionStatus)
	}

	var escrowStatus string
	err = db.QueryRow("SELECT escrow_status FROM escrow_accounts WHERE transaction_id = $1", transactionID).Scan(&escrowStatus)
	if err != nil {
		t.Fatalf("Error getting escrow status: %v", err)
	}
	if escrowStatus != "cancelled" {
		t.Errorf("Expected escrow status to be 'cancelled', but got '%s'", escrowStatus)
	}

	var disputeStatus string
	err = db.QueryRow("SELECT dispute_status FROM disputes WHERE dispute_id = $1", disputeID).Scan(&disputeStatus)
	if err != nil {
		t.Fatalf("Error getting dispute status: %v", err)
	}
	if disputeStatus != "resolved" {
		t.Errorf("Expected dispute status to be 'resolved', but got '%s'", disputeStatus)
	}

	verifyRefund(t, db, transactionID, buyerID, 300.00, tb) // Added verifyRefund call
}

func createPayment(t *testing.T, db *sql.DB, transactionID string, buyerID string, amount float64) string {
	var paymentID string
	query := `
		INSERT INTO payments (transaction_id, amount, method, payment_status, encrypted_details)
		VALUES ($1, $2, 'credit_card', 'completed', $3)
		RETURNING payment_id;
	`
	dummyEncryptedDetails := []byte("dummy_encrypted_details")
	if err := db.QueryRow(query, transactionID, amount, dummyEncryptedDetails).Scan(&paymentID); err != nil {
		t.Fatalf("createPayment failed: %v", err)
	}
	return paymentID
}

func updateTransactionWithPayment(t *testing.T, db *sql.DB, transactionID string, paymentID string) {
	query := `
		UPDATE transactions
		SET payment_id = $1
		WHERE transaction_id = $2
	`
	if _, err := db.Exec(query, paymentID, transactionID); err != nil {
		t.Fatalf("updateTransactionWithPayment failed: %v", err)
	}
}

func verifyRefund(t *testing.T, db *sql.DB, transactionID, buyerID string, amount float64, tb testing.TB) {
	var refundCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM payments WHERE transaction_id = $1 AND amount = $2 AND payment_status='refunded'`,
		transactionID, amount).Scan(&refundCount)
	if err != nil {
		t.Fatalf("Error checking for refund: %v", err)
	}

	if refundCount == 0 {
		t.Errorf("No completed refund found for transaction %s, buyer %s, amount %f", transactionID, buyerID, amount)
	} else {
		t.Logf("Verified refund of %f for transaction %s", amount, transactionID)
	}
}

func createTransaction(t *testing.T, db *sql.DB, buyerID, sellerID string, amount float64) string {
	var transactionID string
	query := `
		INSERT INTO transactions (buyer_id, seller_id, amount, escrow_status, transaction_status)
		VALUES ($1, $2, $3, 'pending', 'pending')
		RETURNING transaction_id;
	`
	if err := db.QueryRow(query, buyerID, sellerID, amount).Scan(&transactionID); err != nil {
		t.Fatalf("createTransaction failed: %v", err)
	}
	return transactionID
}

func fundEscrow(t *testing.T, db *sql.DB, transactionID string, amount float64) {
	query := `
		INSERT INTO escrow_accounts (transaction_id, escrowed_amount, escrow_status)
		VALUES ($1, $2, 'funded')
		ON CONFLICT (transaction_id) DO UPDATE
		  SET escrowed_amount = EXCLUDED.escrowed_amount,
		      escrow_status    = 'funded';
	`
	if _, err := db.Exec(query, transactionID, amount); err != nil {
		t.Fatalf("fundEscrow failed: %v", err)
	}
}

func addTransactionLog(t *testing.T, db *sql.DB, transactionID, eventType, eventDetails string) {
	query := `
		INSERT INTO transaction_logs (transaction_id, event_type, event_details)
		VALUES ($1, $2, $3);
	`
	if _, err := db.Exec(query, transactionID, eventType, eventDetails); err != nil {
		t.Fatalf("addTransactionLog failed: %v", err)
	}
}

func cleanTables(t *testing.T, db *sql.DB) {
	query := `
        TRUNCATE escrow_accounts,
                  transaction_logs,
                  files,
                  disputes,
                  payments,
                  transactions,
                  users
        RESTART IDENTITY CASCADE;
    `
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("Could not truncate tables: %v", err)
	}
	t.Log("All tables truncated for a clean test scenario.")
}

func createUser(t *testing.T, db *sql.DB, username, pwdHash, role string) string {
	var userID string
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING user_id;
	`

	if err := db.QueryRow(query, username, pwdHash, role).Scan(&userID); err != nil {
		t.Fatalf("createUser failed: %v", err)
	}
	return userID
}

func getEnv(key, defVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defVal
	}
	return val
}

func releaseEscrow(t *testing.T, db *sql.DB, transactionID string) {
	query := `
		UPDATE escrow_accounts
		SET escrow_status = 'released'
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(query, transactionID); err != nil {
		t.Fatalf("releaseEscrow failed: %v", err)
	}
}

func completeTransaction(t *testing.T, db *sql.DB, transactionID string) {
	query := `
		UPDATE transactions
		SET transaction_status = 'completed',
		    updated_at = NOW()
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(query, transactionID); err != nil {
		t.Fatalf("completeTransaction failed: %v", err)
	}
}