package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

func main() {
	ctx := context.Background()

	// Get the cloud function name and other details from environment variables
	cfName := os.Getenv("CF_NAME")
	bucketName := os.Getenv("BUCKET_NAME")
	objectName := os.Getenv("OBJECT_NAME")
	vmFilePath := os.Getenv("VM_FILE_PATH")

	if cfName == "" || bucketName == "" || objectName == "" || vmFilePath == "" {
		log.Fatal("Missing required environment variables.")
	}

	// Create a client
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Initialize bucket and object within the bucket
	bkt := client.Bucket(bucketName)
	obj := bkt.Object(objectName)

	// Create a file in the vm to store the downloaded object
	file, err := os.Create(vmFilePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Create a reader to read the object from the bucket
	rc, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer rc.Close()

	// Copy the object from the bucket to the local file system
	if _, err = io.Copy(file, rc); err != nil {
		log.Fatalf("Failed to copy file: %v", err)
	}

	fmt.Printf("File downloaded from GCS. Cloud function name: %s\n", cfName)
}