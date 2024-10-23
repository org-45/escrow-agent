package auth_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"escrow-agent/internal/auth"
	"escrow-agent/internal/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler_Success(t *testing.T) {

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock DB: %v", err)
	}
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db.DB = sqlxDB

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mock.ExpectQuery("SELECT user_id, username, password_hash, role FROM users").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "password_hash", "role"}).
			AddRow(1, "testuser", passwordHash, "buyer"))

	loginReq := auth.UserCredentials{
		Username: "testuser",
		Password: "password123",
	}
	payload, _ := json.Marshal(loginReq)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, response["token"])
}

func TestLoginHandler_InvalidPassword(t *testing.T) {

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock DB: %v", err)
	}
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db.DB = sqlxDB

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	mock.ExpectQuery("SELECT user_id, username, password_hash, role FROM users").
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "username", "password_hash", "role"}).
			AddRow(1, "testuser", passwordHash, "buyer"))

	loginReq := auth.UserCredentials{
		Username: "testuser",
		Password: "wrongpassword",
	}
	payload, _ := json.Marshal(loginReq)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestLoginHandler_InvalidUsername(t *testing.T) {

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock DB: %v", err)
	}
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	db.DB = sqlxDB

	mock.ExpectQuery("SELECT user_id, username, password_hash, role FROM users").
		WithArgs("testuser").
		WillReturnError(sql.ErrNoRows)

	loginReq := auth.UserCredentials{
		Username: "testuser",
		Password: "password123",
	}
	payload, _ := json.Marshal(loginReq)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.LoginHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
