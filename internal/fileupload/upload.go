package fileupload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"escrow-agent/internal/db"
	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type File struct {
	ID            uuid.UUID       `json:"id"`
	TransactionID uuid.UUID       `json:"transaction_id"`
	FileName      string    `json:"file_name"`
	FilePath      string    `json:"file_path"`
	UploadedAt    time.Time `json:"uploaded_at"`
}

var minioClient *minio.Client

func init() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln("Error initializing MinIO client:", err)
	}

	err = setupMinIO()
	if err != nil {
		log.Fatalln("Error setting up MinIO:", err)
	}
}

func setupMinIO() error {
	ctx := context.Background()
	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	log.Printf("Setting up MinIO bucket '%s'...\n", bucketName)

	var err error
	for i := 0; i < 5; i++ {
		exists, err := minioClient.BucketExists(ctx, bucketName)
		if err == nil {
			if !exists {
				err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: os.Getenv("MINIO_REGION")})
				if err == nil {
					log.Printf("Bucket '%s' created successfully\n", bucketName)
					break
				}
			} else {
				log.Printf("Bucket '%s' already exists\n", bucketName)
				break
			}
		}

		log.Printf("Retrying MinIO setup (attempt %d): %v\n", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to create bucket after retries: %v", err)
	}

	return nil
}

func saveFileToDB(transactionID, fileName, filePath string) error {
	query := `
        INSERT INTO files (transaction_id, file_name, file_path)
        VALUES ($1, $2, $3)
    `
	_, err := db.DB.Exec(query, transactionID, fileName, filePath)
	return err
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (10 MB max file size)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Retrieve file from request
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	transactionID := r.FormValue("transactionID")
	if transactionID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	objectName := fmt.Sprintf("transactions/%s/%s", transactionID, header.Filename)
	contentType := header.Header.Get("Content-Type")

	metadata := map[string]string{
		"transactionID": transactionID,
		"originalName":  header.Filename,
	}

	ctx := context.Background()
	info, err := minioClient.PutObject(ctx, bucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
	})
	if err != nil {
		log.Printf("Failed to upload object to MinIO: %v\n", err)
		http.Error(w, "Error uploading file to MinIO", http.StatusInternalServerError)
		return
	}

	filepath := objectName
	err = saveFileToDB(transactionID, header.Filename, filepath)
	if err != nil {
		http.Error(w, "Error saving file metadata to database", http.StatusInternalServerError)
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s (%d bytes)\n", objectName, info.Size)
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Extract transactionID from the URL path
	vars := mux.Vars(r)
	transactionIDStr := vars["transactionID"]

	// Convert transactionID to an integer
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	// Query the database for files associated with the transactionID
	query := `
        SELECT id, transaction_id, file_name, file_path, uploaded_at
        FROM files
        WHERE transaction_id = $1
    `
	rows, err := db.DB.Query(query, transactionID)
	if err != nil {
		log.Printf("Error querying database: %v\n", err)
		http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate over the rows and build a list of files
	var files []File
	for rows.Next() {
		var file File
		err := rows.Scan(&file.ID, &file.TransactionID, &file.FileName, &file.FilePath, &file.UploadedAt)
		if err != nil {
			log.Printf("Error scanning row: %v\n", err)
			http.Error(w, "Failed to read file data", http.StatusInternalServerError)
			return
		}
		files = append(files, file)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v\n", err)
		http.Error(w, "Failed to retrieve files", http.StatusInternalServerError)
		return
	}

	// Return a 404 if no files are found
	if len(files) == 0 {
		http.Error(w, "No files found for the given transaction ID", http.StatusNotFound)
		return
	}

	// Return the list of files as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
}
