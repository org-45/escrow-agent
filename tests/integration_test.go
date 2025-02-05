// +build integration

package integration_test

import(
	"database/sql"
	"fmt"
	"os"
	"testing"
	_ "github.com/lib/pq"
)

func TestIntegrationFlow(t *testing.T){

	db := connectDB(t)
	defer db.Close()

	//optional for fresh test
	cleanTables(t,db)

	//create users

	buyerID := createUser(t,db,"bd","some_hashed_password","buyer")
	sellerID := createUser(t,db,"sd","some_hashed_password","seller")
	adminID := createUser(t,db,"ad","some_hashed_password","admin")

	t.Logf("Users created -> buyer=%s, seller=%s, admin=%s", buyerID, sellerID, adminID)

	verifyAllData(t, db)
	t.Log("Integration test completed successfully!")


}

func connectDB(t *testing.T) *sql.DB {
	// adjust these to match your setup or use environment variables
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

func cleanTables(t *testing.T, db *sql.DB){

	tables := []string{
		"escrow_accounts",
		"transaction_logs",
		"files",
		"disputes",
		"payments",
		"transactions",
		"users",
	}

	for _,table := range tables{
		_, error := db.Exec("TRUNCATE " + table + " CASCADE")
		if error != nil{
			t.Logf("Could not truncate table %s: %v", table, error)
		}
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

func verifyAllData(t *testing.T, db *sql.DB) {
	t.Log("Verifying data in all tables...")

	checkRowCount(t, db, "users")

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