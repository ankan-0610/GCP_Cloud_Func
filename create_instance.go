package main

import (
	"context"
	"fmt"
	"io"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/protobuf/proto"
        "google.golang.org/grpc/status"
        "google.golang.org/grpc/codes"
)

func firewallRuleExists(ctx context.Context, projectID, firewallName string) (bool, error) {
	firewallClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return false, fmt.Errorf("NewFirewallsRESTClient: %w", err)
	}
	defer firewallClient.Close()

	req := &computepb.GetFirewallRequest{
		Project: projectID,
		Firewall: firewallName,
	}

	_, err = firewallClient.Get(ctx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, fmt.Errorf("unable to get firewall rule: %w", err)
	}

	return true, nil
}

// createInstance sends an instance creation request to the Compute Engine API and waits for it to complete.
func createInstance(w io.Writer) error {
        projectID := "cloudsec-390404"
		zone := "us-east4-c" // Change this to your desired zone
		instanceName := "test-vm-inst-2"
		machineType := "n1-standard-1" // Change this to your desired machine type
        sourceImage := "projects/cloudsec-390404/global/images/image-1"
        networkName := "global/networks/default"

        ctx := context.Background()
        instancesClient, err := compute.NewInstancesRESTClient(ctx)
        if err != nil {
                return fmt.Errorf("NewInstancesRESTClient: %w", err)
        }
        defer instancesClient.Close()

        // Create a firewall rule to allow ssh traffic
        firewallClient, err := compute.NewFirewallsRESTClient(ctx)
        if err != nil {
        	return fmt.Errorf("NewFirewallsRESTClient: %w", err)
        }
        defer firewallClient.Close()

        // Check if the firewall rule already exists
	firewallExists, err := firewallRuleExists(ctx, projectID, "allow-ssh")
	if err != nil {
		return fmt.Errorf("unable to check if firewall rule exists: %w", err)
	}

        if !firewallExists{
                // Create a firewall rule to allow ssh traffic
		firewallReq := &computepb.InsertFirewallRequest{
			Project: projectID,
			FirewallResource: &computepb.Firewall{
				Name:         proto.String("allow-ssh"),
				Network:      proto.String(networkName),
				SourceRanges: []string{"0.0.0.0/0"},
				Allowed: []*computepb.Allowed{
					{
						IPProtocol: proto.String("tcp"),
						Ports:      []string{"22"},
					},
				},
			},
		}

		firewallOp, err := firewallClient.Insert(ctx, firewallReq)
		if err != nil {
			return fmt.Errorf("unable to create firewall rule: %w", err)
		}

		if err = firewallOp.Wait(ctx); err != nil {
			return fmt.Errorf("unable to wait for the firewall operation: %w", err)
		}
        } else{
                firewallName := "allow-ssh"
                // If the firewall rule already exists, update it
                updatedFirewall := &computepb.Firewall{
                        Name:         proto.String(firewallName),
                        Network:      proto.String(networkName),
                        SourceRanges: []string{"0.0.0.0/0"},
                        Allowed: []*computepb.Allowed{
                                {
                                        IPProtocol: proto.String("tcp"),
                                        Ports:      []string{"22"},
                                },
                        },
                }
                op, err := firewallClient.Update(ctx, &computepb.UpdateFirewallRequest{
                        Project: projectID,
                        Firewall: firewallName,
                        FirewallResource: updatedFirewall,
                })
                if err != nil {
                        return fmt.Errorf("unable to update firewall rule: %w", err)
                }

                if err = op.Wait(ctx); err != nil {
                        return fmt.Errorf("unable to wait for the firewall operation: %w", err)
                }
        }

        req := &computepb.InsertInstanceRequest{
                Project: projectID,
                Zone:    zone,
                InstanceResource: &computepb.Instance{
                        Name: proto.String(instanceName),
                        Disks: []*computepb.AttachedDisk{
                                {
                                        InitializeParams: &computepb.AttachedDiskInitializeParams{
                                                DiskSizeGb:  proto.Int64(10),
                                                SourceImage: proto.String(sourceImage),
                                        },
                                        AutoDelete: proto.Bool(true),
                                        Boot:       proto.Bool(true),
                                        Type:       proto.String(computepb.AttachedDisk_PERSISTENT.String()),
                                },
                        },
                        MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", zone, machineType)),
                        NetworkInterfaces: []*computepb.NetworkInterface{
                                {
                                        Name: proto.String(networkName),
                                        AccessConfigs: []*computepb.AccessConfig{
                                                {
                                                        Name: proto.String("External NAT"),
                                                        Type: proto.String(computepb.AccessConfig_ONE_TO_ONE_NAT.String()),
                                                },
                                        },
                                },
                        },
                },
        }

        op, err := instancesClient.Insert(ctx, req)
        if err != nil {
                return fmt.Errorf("unable to create instance: %w", err)
        }

        if err = op.Wait(ctx); err != nil {
                return fmt.Errorf("unable to wait for the operation: %w", err)
        }

        fmt.Fprintf(w, "Instance created successfully \n")

        return nil
}

func main(){
	if err := createInstance(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating instance: %v\n", err)
		os.Exit(1)
	}
}