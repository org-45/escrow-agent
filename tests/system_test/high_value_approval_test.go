// +build high_value_approval

// High-Value Transaction + Admin Approval
// 1. Buyer (Quinn) places a $5,000 transaction from Seller (Rita). Payment method might be bank_transfer.
// 2. Quinn funds escrow (funded).
// 3. Because it’s high-value, an admin manually reviews the transaction before shipment.
// 4. Approval given, Rita ships the item. Buyer confirms -> escrow released.
// 5. Transaction ends completed.
// Key Points: Demonstrates a case where manual or admin checks happen before completion due to high value.

package high_value_approval_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const highValueThreshold = 1000 // Adjust to whatever your threshold is

func TestHighValueTransactionFlow(t *testing.T) {
	db := connectDB(t)
	defer db.Close()

	// Ensure a clean slate
	cleanTables(t, db)

	// Create users
	buyerID := createUser(t, db, "Quinn", "some_hashed_password", "buyer")
	sellerID := createUser(t, db, "Rita", "some_hashed_password", "seller")
	adminID := createUser(t, db, "AdminUser", "some_hashed_password", "admin")
	t.Logf("Users created buyer=%s, seller=%s, admin=%s", buyerID, sellerID, adminID)

	// Quinn initiates a transaction
	transactionID := createTransaction(t, db, buyerID, sellerID, 5000.00)
	t.Logf("Transaction created (pending): %s", transactionID)

	// Create the initial payment record
	paymentID := createPayment(t, db, transactionID, buyerID, 5000.00, "bank_transfer") // Using bank transfer
	t.Logf("Payment created: %s", paymentID)

	// Update transaction with payment id
	updateTransactionWithPayment(t, db, transactionID, paymentID)
	t.Logf("Transaction %s updated with payment %s", transactionID, paymentID)

	// Quinn funds escrow
	fundEscrow(t, db, transactionID, 5000.00)
	t.Logf("Escrow funded with $5000 for transaction: %s", transactionID)

	// Admin reviews the transaction and approves it
	approveTransaction(t, db, transactionID, adminID)
	t.Logf("Admin approved transaction: %s", transactionID)

	// Rita ships the item
	addTransactionLog(t, db, transactionID, "SHIPPING", "Rita shipped the item after admin approval.")
	t.Log("Shipping log added.")

	// Buyer confirms receipt
	addTransactionLog(t, db, transactionID, "DELIVERY", "Quinn confirmed receipt.")
	t.Log("Delivery log added.")

	// Release escrow and complete transaction
	releaseEscrowAndCompleteTransaction(t, db, transactionID)
	t.Logf("Escrow released and transaction completed: %s", transactionID)

	// Verify the final state
	verifyTransactionCompleted(t, db, transactionID)

	t.Log("High-value transaction with admin approval test completed successfully!")
}

func approveTransaction(t *testing.T, db *sql.DB, transactionID, adminID string) {
	// Simulates admin approval process. In a real system, this might involve updating a status flag
	// in a transactions table or a separate approval table.  For simplicity, we're just adding a log.

	//Check first transaction
	var transactionStatus string
	err := db.QueryRow("SELECT transaction_status FROM transactions WHERE transaction_id = $1", transactionID).Scan(&transactionStatus)
	if err != nil {
		t.Fatalf("Error getting transaction status: %v", err)
	}

	addTransactionLog(t, db, transactionID, "ADMIN_APPROVAL", fmt.Sprintf("Transaction approved by admin %s", adminID))
}

func releaseEscrowAndCompleteTransaction(t *testing.T, db *sql.DB, transactionID string) {
	//Update escrow account as released
	queryEscrow := `
		UPDATE escrow_accounts
		SET escrow_status = 'released',
		    released_at = NOW()
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(queryEscrow, transactionID); err != nil {
		t.Fatalf("releaseEscrow failed: %v", err)
	}

	//Complete the transaction
	queryTransaction := `
		UPDATE transactions
		SET transaction_status = 'completed',
		    updated_at = NOW()
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(queryTransaction, transactionID); err != nil {
		t.Fatalf("completeTransaction failed: %v", err)
	}

	addTransactionLog(t, db, transactionID, "RESOLUTION", "Escrow released to seller and transaction completed after admin approval.")
}

func verifyTransactionCompleted(t *testing.T, db *sql.DB, transactionID string) {
	var transactionStatus string
	err := db.QueryRow("SELECT transaction_status FROM transactions WHERE transaction_id = $1", transactionID).Scan(&transactionStatus)
	if err != nil {
		t.Fatalf("Error getting transaction status: %v", err)
	}
	if transactionStatus != "completed" {
		t.Errorf("Expected transaction status to be 'completed', but got '%s'", transactionStatus)
	}

	var escrowStatus string
	err = db.QueryRow("SELECT escrow_status FROM escrow_accounts WHERE transaction_id = $1", transactionID).Scan(&escrowStatus)
	if err != nil {
		t.Fatalf("Error getting escrow status: %v", err)
	}
	if escrowStatus != "released" {
		t.Errorf("Expected escrow status to be 'released', but got '%s'", escrowStatus)
	}
}

func createPayment(t *testing.T, db *sql.DB, transactionID string, buyerID string, amount float64, method string) string {
	var paymentID string
	query := `
		INSERT INTO payments (transaction_id, amount, method, payment_status, encrypted_details)
		VALUES ($1, $2, $3, 'completed', $4)
		RETURNING payment_id;
	`
	dummyEncryptedDetails := []byte("dummy_encrypted_details")
	if err := db.QueryRow(query, transactionID, amount, method, dummyEncryptedDetails).Scan(&paymentID); err != nil {
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