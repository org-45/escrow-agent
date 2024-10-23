package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"escrow-agent/internal/auth"
	"escrow-agent/internal/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock DB: %v", err)
	}
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")

	db.DB = sqlxDB

	createdAt := time.Now()
	mock.ExpectQuery("INSERT INTO users").
		WithArgs("testuser", sqlmock.AnyArg(), "buyer").
		WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(createdAt))

	registerReq := auth.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Role:     "buyer",
	}
	payload, _ := json.Marshal(registerReq)

	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.RegisterHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response auth.RegisterResponse
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "User registered successfully", response.Message)
	assert.WithinDuration(t, time.Now(), response.CreatedAt, time.Second)
}

func TestRegisterHandler_InvalidInput(t *testing.T) {
	invalidReq := auth.RegisterRequest{
		Username: "",
		Password: "short",
		Role:     "invalidrole",
	}
	payload, _ := json.Marshal(invalidReq)

	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.RegisterHandler)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
