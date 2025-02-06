// +build simple_purchase


// Simple Purchase
// 1. Buyer (Alice) wants to buy an item from Seller (Bob) for $50.
// 2. Alice creates a transaction in “pending” status.
// 3. Alice funds the escrow account with $50 (escrow: funded).
// 4. Bob ships/delivers the item.
// 5. Alice confirms receipt; escrow is updated to released.
// 6. Transaction final status: completed.
// Key Points: Straightforward flow from pending → funded → released → completed.


package simple_purchase_test

import(
	"database/sql"
	"fmt"
	"os"
	"testing"
	_ "github.com/lib/pq"
)

func TestSimplePurchaseFlow(t *testing.T){
	db := connectDB(t)
	defer db.Close()

	//optional for fresh test
	cleanTables(t,db)

	//create users

	buyerID := createUser(t,db,"Alice","some_hashed_password","buyer")
	sellerID := createUser(t,db,"Bob","some_hashed_password","seller")
	t.Logf("Users created buyer=%s, seller=%s", buyerID, sellerID)

	transactionID := createTransaction(t, db, buyerID, sellerID, 50.00)
	t.Logf("Transaction created (pending): %s", transactionID)

	fundEscrow(t, db, transactionID, 50.00)
	t.Logf("Escrow funded with $50 for transaction: %s", transactionID)

	addTransactionLog(t, db, transactionID, "SHIPPING", "Bob shipped the item.")
	t.Log("Shipping log added.")

	releaseEscrow(t, db, transactionID)
	t.Logf("Escrow released for transaction: %s", transactionID)

	completeTransaction(t, db, transactionID)
	t.Logf("Transaction completed: %s", transactionID)

	verifyAllData(t, db)

	t.Log("Integration test completed successfully!")

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


func createUser(t *testing.T, db *sql.DB, username,pwdHash, role string) string{
	var userID string
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING user_id;
	`

	if err := db.QueryRow(query, username, pwdHash, role).Scan(&userID);  err!= nil{
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

func verifyAllData(t *testing.T, db *sql.DB) {
	t.Log("Verifying data in all tables...")

	checkRowCount(t, db, "users")
	checkRowCount(t, db, "transactions")


	t.Log("Verification complete.")
}


func checkRowCount(t *testing.T, db *sql.DB, tableName string) {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		t.Fatalf("Error counting rows in %s: %v", tableName, err)
	}
	t.Logf("Table '%s' row count = %d", tableName, count)
}

func getEnv(key, defVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defVal
	}
	return val
}