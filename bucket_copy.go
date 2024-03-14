package main

import (
	"fmt"
	"os/exec"
	"log"
	"strings"
	"context"

	"cloud.google.com/go/storage"
)

func main() {
	// Set the project ID and the region or multi-region where the function is deployed
	projectID := "cloudsec-390404"
	region := "us-central1"
	functionName := "function-2"

	// Run the gcloud command to get the equivalent REST response
	cmd := exec.Command("gcloud", "functions", "describe", functionName, "--region", region, "--project", projectID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to run gcloud command: %v, output: %s", err, output)
	}

	fmt.Println(string(output))
	// Parse the output to extract the "bucket" and "object" for the source code
	// Note that the output format may change depending on the function's configuration
	bucketName := getValue(string(output), "bucket:")
	objectName := getValue(string(output), "object:")
	fmt.Println("Bucket:", bucketName)
	fmt.Println("Object:", objectName)

	ctx := context.Background()

	// Create a storage client.
    client, err := storage.NewClient(ctx)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

	// Get a reference to the bucket.
	bucket := client.Bucket(bucketName)

	// Define the IAM policy bindings.
	policy, err := bucket.IAM().Policy(ctx)
	if err != nil {
		log.Fatalf("Failed to get bucket IAM policy: %v", err)
	}

	// Add a new binding to the IAM policy.
	policy.Add("roles/storage.objectViewer", "allUsers")

	// Set the updated IAM policy for the bucket.
	if err := bucket.IAM().SetPolicy(ctx, policy); err != nil {
		log.Fatalf("Failed to set bucket IAM policy: %v", err)
	}

	fmt.Println("Bucket permissions updated successfully.")
}

func getValue(input string, key string) string {
	start := strings.Index(input, key)
	if start == -1 {
		return ""
	}
	start += len(key)
	end := strings.Index(input[start:], "\n")
	if end == -1 {
		end = len(input)
	} else {
		end += start
	}
	return input[start:end]
}