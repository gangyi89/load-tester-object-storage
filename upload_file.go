//write a go program that upload all files in a specifc directory to the object storage using aws s3 go library
//allow me to specify the upload rate, example 100 concurrent uploads at a time
// allow me to specify the aws key, secret and endpoint url

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Custom AWS logger that writes to our log file
type customLogger struct {
	logger *log.Logger
}

func (l *customLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}

func main() {
	// Create log file with timestamp
	logFileName := fmt.Sprintf("upload_%s.log", time.Now().Format("2006-01-02_15-04-05"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	// Set log output to file
	log.SetOutput(logFile)

	// Create custom AWS logger
	awsLogger := &customLogger{
		logger: log.New(logFile, "AWS: ", log.LstdFlags),
	}

	// Generate folder name once at start
	folderPath := time.Now().Format("2006-01-02_15-04-05")

	// Log start of upload process
	log.Printf("Starting upload process to folder: %s", folderPath)

	// Parse command line arguments
	uploadRate := flag.Int("rate", 100, "number of concurrent uploads")
	dir := flag.String("dir", "load_test_files", "directory containing files to upload")
	bucket := flag.String("bucket", "my-bucket", "bucket name")
	endpoint := flag.String("endpoint", "", "S3 endpoint URL")
	accessKey := flag.String("access-key", "", "AWS access key")
	secretKey := flag.String("secret-key", "", "AWS secret key")

	flag.Parse()

	log.Printf("Starting upload with rate: %d, directory: %s, bucket: %s, endpoint: %s\n",
		*uploadRate, *dir, *bucket, *endpoint)
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(*endpoint),
		Region:           aws.String("us-east-1"),
		Credentials:      credentials.NewStaticCredentials(*accessKey, *secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
		Logger:           awsLogger,
		LogLevel:         aws.LogLevel(aws.LogDebugWithSigning),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create S3 client
	client := s3.New(sess)

	// Get a list of files to upload
	files, err := filepath.Glob(filepath.Join(*dir, "*"))
	if err != nil {
		log.Fatal("Error reading directory:", err)
	}
	log.Printf("Found %d files to upload\n", len(files))
	if len(files) == 0 {
		log.Fatal("No files found in directory")
	}

	// Create a channel to limit the number of concurrent uploads
	uploadChan := make(chan string, *uploadRate)

	// Create a WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Start a fixed number of upload goroutines
	for i := 0; i < *uploadRate; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range uploadChan {
				log.Printf("Starting upload of %s\n", file)
				if err := uploadFile(client, file, *bucket, folderPath); err != nil {
					log.Printf("Error processing %s: %v", file, err)
				}
			}
		}()
	}

	// Add files to the upload channel
	for _, file := range files {
		uploadChan <- file
	}
	log.Printf("All files queued for upload\n")
	close(uploadChan)

	// Wait for all uploads to complete
	wg.Wait()
	log.Printf("Upload complete\n")
}

func uploadFile(client *s3.S3, filePath, bucket, folderPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file %s: %v\n", filePath, err)
		return err
	}
	defer file.Close()

	// Enable debug logging for the S3 client
	client.Config.LogLevel = aws.LogLevel(aws.LogDebugWithSigning)

	// Create an uploader with the same session configuration as the client
	uploader := s3manager.NewUploaderWithClient(client)

	// Modify objectKey to include the new folder path
	objectKey := fmt.Sprintf("%s/%s", folderPath, filepath.Base(filePath))

	// Upload the file
	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   file,
	}

	_, err = uploader.Upload(uploadInput)
	if err != nil {
		// Check if it's a 403 error
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "AccessDenied" {
			log.Printf("403 Access Denied error for file %s. Request details:", filePath)
			log.Printf("Bucket: %s", bucket)
			log.Printf("Object Key: %s", objectKey)
			log.Printf("Error Message: %v", err)
		}
		log.Printf("Failed to upload file %s: %v", filePath, err)
		return err
	}

	log.Printf("Successfully uploaded %s to %s/%s", filePath, bucket, objectKey)
	return nil
}
