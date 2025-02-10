// +build escrow_extended

// Escrow Extended Due to Shipping Delays
// 1. Buyer (Oliver) orders $350 from Seller (Paula).
// 2. Oliver funds the escrow (funded).
// 3. Shipping is delayed beyond the expected date.
// 4. The system or buyer extends the escrow period (updating some expiration field in escrow_accounts).
// 5. Paula eventually ships. Buyer confirms receipt.
// 6. Escrow released, transaction completed.
// Key Points: Showcases updating an “expiry_date” or extending escrow status due to real-life delays.

package escrow_extended_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestEscrowExtendedFlow(t *testing.T) {
	db := connectDB(t)
	defer db.Close()

	// Ensure a clean slate
	cleanTables(t, db)

	// Create users
	buyerID := createUser(t, db, "Oliver", "some_hashed_password", "buyer")
	sellerID := createUser(t, db, "Paula", "some_hashed_password", "seller")
	t.Logf("Users created buyer=%s, seller=%s", buyerID, sellerID)

	// Oliver initiates a transaction
	transactionID := createTransaction(t, db, buyerID, sellerID, 350.00)
	t.Logf("Transaction created (pending): %s", transactionID)

	// Create the initial payment record
	paymentID := createPayment(t, db, transactionID, buyerID, 350.00)
	t.Logf("Payment created: %s", paymentID)

	// Update transaction with payment id
	updateTransactionWithPayment(t, db, transactionID, paymentID)
	t.Logf("Transaction %s updated with payment %s", transactionID, paymentID)

	// Oliver funds escrow
	fundEscrow(t, db, transactionID, 350.00)
	t.Logf("Escrow funded with $350 for transaction: %s", transactionID)

	// Simulate expected delivery date and delay
	expectedDelivery := time.Now().Add(7 * 24 * time.Hour) // 7 days from now
	t.Logf("Expected delivery date: %s", expectedDelivery)

	extendEscrowExpiry(t, db, transactionID, expectedDelivery.Add(5*24*time.Hour)) // Extends by 5 days
	t.Logf("Escrow expiry extended by 5 days.")

	// Paula ships the item after the initial expected delivery date
	addTransactionLog(t, db, transactionID, "SHIPPING", "Paula shipped the item (delayed).")
	t.Log("Shipping log added (delayed).")

	// Oliver confirms receipt
	addTransactionLog(t, db, transactionID, "DELIVERY", "Oliver confirmed receipt.")
	t.Log("Delivery log added.")

	// Release escrow and complete transaction
	releaseEscrowAndCompleteTransaction(t, db, transactionID)
	t.Logf("Escrow released and transaction completed: %s", transactionID)

	// Verify final state
	verifyTransactionCompleted(t, db, transactionID)

	t.Log("Escrow extended due to shipping delays test completed successfully!")
}

func extendEscrowExpiry(t *testing.T, db *sql.DB, transactionID string, newExpiryDate time.Time) {
	query := `
		UPDATE escrow_accounts
		SET expiry_date = $2
		WHERE transaction_id = $1
	`
	if _, err := db.Exec(query, transactionID, newExpiryDate); err != nil {
		t.Fatalf("extendEscrowExpiry failed: %v", err)
	}
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

	addTransactionLog(t, db, transactionID, "RESOLUTION", "Escrow released to seller and transaction completed after extended escrow period.")
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
		INSERT INTO escrow_accounts (transaction_id, escrowed_amount, escrow_status, expiry_date)
		VALUES ($1, $2, 'funded', $3)
		ON CONFLICT (transaction_id) DO UPDATE
		  SET escrowed_amount = EXCLUDED.escrowed_amount,
		      escrow_status    = 'funded',
		      expiry_date = EXCLUDED.expiry_date;
	`
	expectedDelivery := time.Now().Add(7 * 24 * time.Hour) // 7 days from now
	if _, err := db.Exec(query, transactionID, amount, expectedDelivery); err != nil {
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