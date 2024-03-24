package job

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"nunet/app/shared"
	"nunet/pkg"
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
	return j.DeploymentTopic.ListPeers()
}

func (j *Job) PublishDeploymentRequest(ctx context.Context, request shared.DeployRequest) error {
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

		var request shared.DeployRequest
		if err := json.Unmarshal(msg.GetData(), &request); err != nil {
			fmt.Println("Error unmarshalling request:", err)
			continue
		}

		if request.TargetPeerID != j.Host.ID().String() {
			fmt.Println("Received deployment request for another peer")
			continue
		}

		output, pid, err := pkg.RunCmd(request.Program, request.Arguments...)
		if err != nil {
			fmt.Println("Error processing deployment request:", err)
		}

		if err := j.sendDeploymentResponse(ctx, request, pid, output, err); err != nil {
			fmt.Println("Error responding to deployment request:", err)
		}
	}
}

// SendDeploymentResponse sends a response to the deployment request
func (j *Job) sendDeploymentResponse(
	ctx context.Context,
	request shared.DeployRequest,
	pid int,
	output []string,
	err error,
) error {
	response := shared.DeployResponse{
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

		var response shared.DeployResponse
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
