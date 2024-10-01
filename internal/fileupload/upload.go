package fileupload

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIO client setup
var minioClient *minio.Client

func init() {
    endpoint := "localhost:9000"
    accessKeyID := "minioadmin"
    secretAccessKey := "minioadmin"
    useSSL := false

    var err error
    minioClient, err = minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        log.Fatalln(err)
    }
}

// UploadHandler handles file uploads
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

    // Upload file to MinIO
    bucketName := "escrow-documents"
    objectName := header.Filename
    contentType := header.Header.Get("Content-Type")

    ctx := context.Background()
    err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
    if err != nil {
        exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
        if errBucketExists == nil && exists {
            log.Printf("Bucket %s already exists\n", bucketName)
        } else {
            http.Error(w, "Error with MinIO bucket", http.StatusInternalServerError)
            return
        }
    }

    // Upload file to the bucket
    info, err := minioClient.PutObject(ctx, bucketName, objectName, file, header.Size, minio.PutObjectOptions{ContentType: contentType})
    if err != nil {
        http.Error(w, "Error uploading file to MinIO", http.StatusInternalServerError)
        return
    }

    // Success response
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "File uploaded successfully: %s (%d bytes)\n", objectName, info.Size)
}
