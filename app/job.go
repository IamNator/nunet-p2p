package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

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

		output, pid, err := runCmd(request.Program, request.Arguments...)
		if err != nil {
			fmt.Println("Error processing deployment request:", err)
		}

		if err := SendDeploymentResponse(ctx, host, topic, request, pid, output, err); err != nil {
			fmt.Println("Error responding to deployment request:", err)
		}
	}
}

// runCmd executes the given command with the provided arguments
func runCmd(name string, args ...string) ([]string, int, error) {

	fmt.Printf("Executing command: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)

	// get the outputs
	var outputs []string

	// Attach the stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, 0, fmt.Errorf("error attaching stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, 0, fmt.Errorf("error attaching stderr pipe: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	// get the outputs
	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)
			if n > 0 {
				outputs = append(outputs, "Info :"+strings.TrimSpace(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)
			if n > 0 {
				outputs = append(outputs, "Error :"+strings.TrimSpace(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, 0, fmt.Errorf("error starting command: %w", err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return outputs, 0, fmt.Errorf("error waiting for command to finish: %w", err)
	}

	wg.Wait()

	fmt.Println("Command executed successfully")
	return outputs, cmd.ProcessState.Pid(), nil
}

// SendDeploymentResponse sends a response to the deployment request
func SendDeploymentResponse(
	ctx context.Context,
	host host.Host,
	deploymentTopic *pubsub.Topic,
	request DeployRequest,
	pid int,
	output []string,
	err error,
) error {
	response := DeployResponse{
		Success:      err == nil,
		SourcePeerID: request.SourcePeerID,
		SourceAddrs:  request.SourceAddrs,
		Program:      request.Program,
		Arguments:    request.Arguments,
		PID:          pid,
		TargetPeerID: request.TargetPeerID,
		TargetAddrs:  request.SourceAddrs, // Check if this should be SourceAddrs or TargetAddrs
		Outputs:      output,
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
			fmt.Printf("Deployment successful. PID: %d\n, outputs: %v \n", response.PID, strings.Join(response.Outputs, ","))
		} else {
			fmt.Println("Deployment failed")
		}
	}
}
