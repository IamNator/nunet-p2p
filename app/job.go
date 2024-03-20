package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pubsub "github.com/libp2p/go-libp2p-pubsub" // for message broadcasting
	"github.com/libp2p/go-libp2p/core/host"
)

func HandleDeploymentRequest(ctx context.Context, host host.Host, sub *pubsub.Subscription) {
	// Process incoming messages
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}
		if host.ID() == msg.GetFrom() { // Ignore messages from self
			continue
		}
		processDeploymentRequest(msg.GetData())
	}
}

func processDeploymentRequest(data []byte) {
	var request DeployRequest
	err := json.Unmarshal(data, &request)
	if err != nil {
		fmt.Println("Error unmarshalling request:", err)
		return
	}
	// ... (container deployment logic using program and arguments)
	fmt.Printf("Deploying container: %s %s from %s\n", request.Program, strings.Join(request.Arguments, " "), request.SourcePeerID)

}
