package router

import (
	"escrow-agent/internal/admin"
	"escrow-agent/internal/auth"
	"escrow-agent/internal/escrow"
	"escrow-agent/internal/fileupload"
	"escrow-agent/internal/logs"
	"escrow-agent/internal/middleware"
	"escrow-agent/internal/profile"
	"escrow-agent/internal/transactions"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// public routes
	r.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	r.HandleFunc("/register", auth.RegisterHandler).Methods("POST")

	// protected routes with JWT middleware
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.JWTAuthMiddleware)

	api.HandleFunc("/profile", profile.ProfileHandler).Methods("GET")
	api.HandleFunc("/profile", profile.ProfileUpdateHandler).Methods("PUT")

	api.HandleFunc("/transactions", transactions.CreateTransactionHandler).Methods("POST")
	api.HandleFunc("/transactions", transactions.GetTransactionsHandler).Methods("GET")
	api.HandleFunc("/transactions/{id}", transactions.GetTransactionHandler).Methods("GET")
	api.HandleFunc("/transactions/{id}/fulfill", transactions.FulfillTransactionHandler).Methods("PUT")
	api.HandleFunc("/transactions/{id}/confirm", transactions.ConfirmDeliveryHandler).Methods("PUT")

	api.HandleFunc("/escrow/{id}/deposit", escrow.DepositEscrowHandler).Methods("POST")
	api.HandleFunc("/escrow/{id}/release", escrow.ReleaseEscrowHandler).Methods("PUT")

	api.HandleFunc("/admin/users", admin.GetUsersHandler).Methods("GET")
	api.HandleFunc("/admin/users/{id}", admin.GetUserByIDHandler).Methods("GET")
	api.HandleFunc("/admin/transactions", admin.GetTransactionsHandler).Methods("GET")

	api.HandleFunc("/logs/{transaction_id}", logs.GetTransactionLogsHandler).Methods("GET")

	api.HandleFunc("/upload", fileupload.UploadHandler).Methods("POST")
	api.HandleFunc("/transactions/{transactionID}/files", fileupload.ListFilesHandler).Methods("GET")

	return r
}
