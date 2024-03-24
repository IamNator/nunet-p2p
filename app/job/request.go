package job

import (
	"context"
	"encoding/json"
	"fmt"

	"nunet/app/shared"
	"nunet/pkg"
)

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

		if request.TargetPeerID != j.Host.ID().String() { // Ignore messages not meant for this peer
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
