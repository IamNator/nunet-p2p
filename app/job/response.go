package job

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nunet/app/shared"
)

// SendDeploymentResponse sends a response to the deployment request
func (j *Job) sendDeploymentResponse(
	ctx context.Context,
	request shared.DeployRequest,
	pid int,
	output []string,
	err error,
) error {
	response := shared.DeployResponse{
		Err:          err.Error(),
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

		if response.SourcePeerID != j.Host.ID().String() { // Ignore messages not meant for this peer
			fmt.Println("Received deployment response for another peer")
			continue
		}

		if strings.TrimSpace(response.Err) == "" {
			fmt.Printf("Deployment successful. PID: %d, %v \n", response.PID, strings.Join(response.Outputs, ","))
		} else {
			fmt.Printf("Deployment failed: %s\n", response.Err)
		}
	}
}
