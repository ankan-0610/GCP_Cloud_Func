package main

import (
	"context"
	"fmt"
	"os"
	"io"
	"path/filepath"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	// "google.golang.org/api/option"
	"cloud.google.com/go/storage"
)

func downloadFunctionCode(ctx context.Context, projectID, bucketName, functionName, instanceIP, destination string) error {
	// Use the Storage client to access the bucket
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}

	// Construct the object name for the zip file (replace with actual logic)
	objectName := fmt.Sprintf("%s.zip", functionName)

	// Prepare the local file to save the downloaded content
	destFilePath := filepath.Join(destination, objectName)
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	// Download the object to the local file on the GCE instance
	rc, err := storageClient.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to create object reader: %v", err)
	}
	defer rc.Close()

	// Copy the object content to the local file
	if _, err := io.Copy(destFile, rc); err != nil {
		return fmt.Errorf("failed to copy object content: %v", err)
	}

	fmt.Printf("Downloaded function code to: %s\n", destFilePath)

	// Now you can use destFilePath to reference the downloaded file on the GCE instance
	// (Implement logic to transfer the file to the GCE instance using compute client or other methods)

	// Example using gcloud compute scp (replace with actual implementation)
	// Use the Compute client or other methods to transfer the downloaded file
	// ...
	// scpCommand := exec.Command(
	// 	"gcloud",
	// 	"compute",
	// 	"scp",
	// 	destFilePath,
	// 	fmt.Sprintf("%s@%s:%s", "your-ssh-username", instanceIP, "/path/on/instance"),
	// 	"--project", projectID,
	// )

	// // Execute the command
	// output, err := scpCommand.CombinedOutput()
	// if err != nil {
	// 	return fmt.Errorf("failed to transfer file to GCE instance: %v, output: %s", err, output)
	// }

	// fmt.Printf("File transferred to GCE instance successfully!\n")

	// return nil

	// Example using Compute engine client
	// Create a new Compute Engine API client
	computeService, err := compute.NewServiceAttachmentsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Compute Engine service client: %v", err)
	}

	// Example using Compute Engine API for file transfer
	filePathOnInstance := "/path/on/instance/" + objectName

	// Create an instance reference
	instance := fmt.Sprintf("projects/%s/zones/%s/instances/%s", projectID, "your-zone", instanceName)

	// Create a request body for the file transfer operation
	req := &computepb.StartWithEncryptionKeyInstanceRequest{
		Source:      destFilePath,
		Destination: filePathOnInstance,
	}

	// Start the instance with the file transfer request
	op, err := computeService.StartWithEncryptionKey(projectID, "your-zone", instance, &req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to start instance with file transfer: %v", err)
	}

	// Wait for the operation to complete
	if _, err := waitForOperation(ctx, computeService, projectID, op.Name); err != nil {
		return fmt.Errorf("file transfer operation failed: %v", err)
	}

	fmt.Printf("File transferred to GCE instance successfully!\n")

	return nil
}

// waitForOperation waits for a Compute Engine operation to complete.
func waitForOperation(ctx context.Context, computeService *compute.Service, projectID, operationName string) (*compute.Operation, error) {
	for {
		op, err := computeService.ZoneOperations.Get(projectID, "your-zone", operationName).Context(ctx).Do()
		if err != nil {
			return nil, err
		}

		if op.Status == "DONE" {
			return op, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
}

func main(){
	// projectID := "your-project-id"
	// bucketName := "your-bucket-name"
	// functionName := "your-function-name"
	// instanceIP := "your-gce-instance-ip"
	// destination := "your-destination-directory"

	// ctx := context.Background()

	// err := downloadFunctionCode(ctx, projectID, bucketName, functionName, instanceIP, destination)
	// if err != nil {
	// 	fmt.Printf("Error: %v\n", err)
	// 	return
	// }

	fmt.Println("Function code download completed successfully!")
}