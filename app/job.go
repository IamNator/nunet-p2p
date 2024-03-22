package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

func HandleDeploymentResponse(ctx context.Context, host host.Host, sub *pubsub.Subscription) {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}
		if host.ID() == msg.GetFrom() { // Ignore messages from self
			continue
		}

		var response DeployResponse
		if err := json.Unmarshal(msg.GetData(), &response); err != nil {
			fmt.Println("Error unmarshalling response:", err)
			continue
		}

		if response.SourcePeerID != host.ID().String() {
			fmt.Println("Received deployment response for another peer")
			continue
		}

		if response.Success {
			fmt.Printf("Deployment successful. PID: %d\n", response.PID)
		} else {
			fmt.Println("Deployment failed")
		}
	}
}

// HandleDeploymentRequest processes incoming deployment requests
func HandleDeploymentRequest(ctx context.Context, host host.Host, sub *pubsub.Subscription, topic *pubsub.Topic) {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}
		if host.ID() == msg.GetFrom() { // Ignore messages from self
			continue
		}

		var request DeployRequest
		if err := json.Unmarshal(msg.GetData(), &request); err != nil {
			fmt.Println("Error unmarshalling request:", err)
			continue
		}

		if request.TargetPeerID != host.ID().String() {
			fmt.Println("Received deployment request for another peer")
			continue
		}

		pid, err := processCMDRequest(request)
		if err != nil {
			fmt.Println("Error processing deployment request:", err)
		}

		if err := SendDeploymentResponse(ctx, host, topic, request, pid, err); err != nil {
			fmt.Println("Error responding to deployment request:", err)
		}
	}
}

// processCMDRequest executes the command described in the deployment request
func processCMDRequest(request DeployRequest) (int, error) {
	fmt.Printf("Executing command: %s %s\n", request.Program, strings.Join(request.Arguments, " "))
	return runCmd(request.Program, request.Arguments...)
}

// runCmd executes the given command with the provided arguments
func runCmd(name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...)

	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("error starting command: %w", err)
	}

	// Get the PID of the process
	pid := cmd.Process.Pid
	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return 0, fmt.Errorf("error waiting for command to finish: %w", err)
	}

	fmt.Println("Command executed successfully")
	return pid, nil
}

// SendDeploymentResponse sends a response to the deployment request
func SendDeploymentResponse(ctx context.Context, host host.Host, deploymentTopic *pubsub.Topic, request DeployRequest, pid int, err error) error {
	response := DeployResponse{
		Success:      err == nil,
		SourcePeerID: request.SourcePeerID,
		SourceAddrs:  request.SourceAddrs,
		Program:      request.Program,
		Arguments:    request.Arguments,
		PID:          pid,
		TargetPeerID: request.TargetPeerID,
		TargetAddrs:  request.SourceAddrs, // Check if this should be SourceAddrs or TargetAddrs
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshalling deployment response: %w", err)
	}

	if err := deploymentTopic.Publish(ctx, responseBytes); err != nil {
		return fmt.Errorf("error publishing deployment response: %w", err)
	}

	fmt.Println("Deployment response sent")
	return nil
}
