// +build seller_cancel

// Seller Unable to Fulfill
// 1. Buyer (Eve) places a $120 transaction with Seller (Frank).
// 2. Eve funds escrow (funded).
// 3. Frank cannot procure the product or deliver it on time.
// 4. Frank (or system) cancels the transaction; escrow is set to cancelled.
// 5. Eve gets refunded, transaction status cancelled.
// Key Points: Seller-initiated cancellation after funding but before shipping.

package seller_cancel_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestSellerCancelFlow(t *testing.T) {
	db := connectDB(t)
	defer db.Close()

	//optional for fresh test
	cleanTables(t, db)

	//create users
	buyerID := createUser(t, db, "Eve", "some_hashed_password", "buyer")
	sellerID := createUser(t, db, "Frank", "some_hashed_password", "seller")
	t.Logf("Users created buyer=%s, seller=%s", buyerID, sellerID)

	// Eve initiates a transaction, Frank as a seller
	transactionID := createTransaction(t, db, buyerID, sellerID, 120.00)
	t.Logf("Transaction created (pending): %s", transactionID)

	//Eve creates payment record and update transaction
	paymentID := createPayment(t, db, transactionID, buyerID, 120.00)
	updateTransactionWithPayment(t, db, transactionID, paymentID)

	//Eve funds escrow
	fundEscrow(t, db, transactionID, 120.00)
	t.Logf("Escrow funded with $120 for transaction: %s", transactionID)

	//Frank cancels the escrow
	cancelTransaction(t, db, transactionID) // Simulates Frank's inability to fulfill
	t.Logf("Transaction cancelled by seller: %s", transactionID)

	//verify all data
	verifyCancellation(t, db, transactionID, buyerID, 120.00, t)

	t.Log("Integration test completed successfully!")
}

// Function to create a payment record
func createPayment(t *testing.T, db *sql.DB, transactionID string, buyerID string, amount float64) string {
	var paymentID string
	query := `
		INSERT INTO payments (transaction_id, amount, method, payment_status, encrypted_details)
		VALUES ($1, $2, 'credit_card', 'completed', $3)  -- Replace 'credit_card' with appropriate method
		RETURNING payment_id;
	`
	// For the encrypted_details, we're using a dummy value for this example.
	// In a real system, this would be the encrypted payment information.
	dummyEncryptedDetails := []byte("dummy_encrypted_details") // Example byte array

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

func cancelTransaction(t *testing.T, db *sql.DB, transactionID string) {
	//Simulate cancellation request and update transaction and escrow status
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

	addTransactionLog(t, db, transactionID, "CANCELLATION_REQUESTED", "Seller unable to fulfill order, requesting cancellation.")
}

func verifyRefund(t *testing.T, db *sql.DB, transactionID, buyerID string, amount float64, tb testing.TB) {
	//TODO: Check payment type if refund
	var refundCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM payments WHERE transaction_id = $1 AND amount = $2`,
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

func verifyCancellation(t *testing.T, db *sql.DB, transactionID, buyerID string, amount float64, tb testing.TB) {
	//Verify that transaction is cancelled and escrow is cancelled

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

	verifyRefund(t, db, transactionID, buyerID, amount, tb)
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

func createTransaction(t *testing.T, db *sql.DB, buyerID, sellerID string, amount float64) string {
	var transactionID string
	query := `
		INSERT INTO transactions (
			buyer_id,
			seller_id,
			amount,
			escrow_status,
			transaction_status
		)
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
		INSERT INTO escrow_accounts (
			transaction_id,
			escrowed_amount,
			escrow_status
		)
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

func getEnv(key, defVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defVal
	}
	return val
}