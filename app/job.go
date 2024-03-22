package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Job struct {
	Host                    host.Host
	DeploymentTopic         *pubsub.Topic
	DeploymentSub           *pubsub.Subscription
	DeploymentResponseTopic *pubsub.Topic
	DeploymentResponseSub   *pubsub.Subscription
}

// NewJob creates a new Job instance
func NewJob(
	h host.Host,
	deploymentTopic *pubsub.Topic,
	deploymentSub *pubsub.Subscription,
	deploymentResponseTopic *pubsub.Topic,
	deploymentResponseSub *pubsub.Subscription,
) *Job {
	return &Job{
		Host:                    h,
		DeploymentTopic:         deploymentTopic,
		DeploymentSub:           deploymentSub,
		DeploymentResponseTopic: deploymentResponseTopic,
		DeploymentResponseSub:   deploymentResponseSub,
	}
}

func (j *Job) ListPeers() []peer.ID {
	return j.Host.Peerstore().Peers()
}

func (j *Job) PublishDeploymentRequest(ctx context.Context, request DeployRequest) error {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshalling deployment request: %w", err)
	}

	if err := j.DeploymentTopic.Publish(ctx, requestBytes); err != nil {
		return fmt.Errorf("error publishing deployment request: %w", err)
	}

	fmt.Println("Deployment request sent")
	return nil
}

// HandleDeploymentRequest processes incoming deployment requests
func (j *Job) HandleDeploymentRequest(ctx context.Context) {
	for {
		msg, err := j.DeploymentSub.Next(ctx)
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}
		if j.Host.ID() == msg.GetFrom() { // Ignore messages from self
			continue
		}

		var request DeployRequest
		if err := json.Unmarshal(msg.GetData(), &request); err != nil {
			fmt.Println("Error unmarshalling request:", err)
			continue
		}

		if request.TargetPeerID != j.Host.ID().String() {
			fmt.Println("Received deployment request for another peer")
			continue
		}

		output, pid, err := runCmd(request.Program, request.Arguments...)
		if err != nil {
			fmt.Println("Error processing deployment request:", err)
		}

		if err := j.sendDeploymentResponse(ctx, request, pid, output, err); err != nil {
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
				outputs = append(outputs, "Info: "+strings.TrimSpace(string(buf[:n])))
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
				outputs = append(outputs, "Error: "+strings.TrimSpace(string(buf[:n])))
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

	select {
	case <-time.After(time.Minute / 2):
		cmd.Process.Kill()
	default:
		wg.Wait()
	}

	fmt.Println("Command executed successfully")
	return outputs, cmd.ProcessState.Pid(), nil
}

// SendDeploymentResponse sends a response to the deployment request
func (j *Job) sendDeploymentResponse(
	ctx context.Context,
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

	if err := j.DeploymentResponseTopic.Publish(ctx, responseBytes); err != nil {
		return fmt.Errorf("error publishing deployment response: %w", err)
	}

	fmt.Println("Deployment response sent")
	return nil
}

func (j *Job) HandleDeploymentResponse(ctx context.Context) {
	for {
		msg, err := j.DeploymentResponseSub.Next(ctx)
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}
		if j.Host.ID() == msg.GetFrom() { // Ignore messages from self
			continue
		}

		var response DeployResponse
		if err := json.Unmarshal(msg.GetData(), &response); err != nil {
			fmt.Println("Error unmarshalling response:", err)
			continue
		}

		if response.SourcePeerID != j.Host.ID().String() {
			fmt.Println("Received deployment response for another peer")
			continue
		}

		if response.Success {
			fmt.Printf("Deployment successful. PID: %d, %v \n", response.PID, strings.Join(response.Outputs, ","))
		} else {
			fmt.Println("Deployment failed")
		}
	}
}
