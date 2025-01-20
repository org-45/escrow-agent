package fileupload

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

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
	// Set CORS rules
	// corsConfig := &cors.Config{
	// 	CORSRules: []cors.Rule{
	// 		{
	// 			AllowedOrigin: []string{"*"},
	// 			AllowedMethod: []string{"GET", "PUT", "POST", "DELETE"},
	// 			AllowedHeader: []string{"*"},
	// 			ExposeHeader:  []string{"ETag"},
	// 			MaxAgeSeconds: 3000,
	// 		},
	// 	},
	// }
	// err = minioClient.SetBucketCors(ctx, bucketName, corsConfig)
	// if err != nil {
	// 	return fmt.Errorf("error setting CORS rules: %v", err)
	// }
	// log.Println("CORS rules set successfully")

	return nil
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

	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	objectName := header.Filename
	contentType := header.Header.Get("Content-Type")

	ctx := context.Background()
	info, err := minioClient.PutObject(ctx, bucketName, objectName, file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		http.Error(w, "Error uploading file to MinIO", http.StatusInternalServerError)
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s (%d bytes)\n", objectName, info.Size)
}
