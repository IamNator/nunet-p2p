package app

import (
	"context"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"nunet/app/api"
	"nunet/app/job"
	"nunet/app/p2p"
	"nunet/pkg"
)

func Run(ctx context.Context, topicName string, port int) error {

	// Create a new libp2p host
	node, err := libp2p.New(
		libp2p.FallbackDefaults,
	)
	if err != nil {
		return fmt.Errorf("failed to create libp2p host: %w", err)
	}
	defer node.Close()

	// Print host information
	pkg.PrintHostInfo(node)

	// Create a new P2P instance
	P2P, err := p2p.New(node)
	if err != nil {
		return fmt.Errorf("failed to create P2P instance: %w", err)
	}

	// Discover peers for communication
	if err := P2P.DiscoverPeers(ctx, topicName); err != nil {
		return fmt.Errorf("failed to discover peers: %w", err)
	}

	// Create pubsub instance and join topic
	pubSub, err := pubsub.NewGossipSub(ctx, node)
	if err != nil {
		return fmt.Errorf("failed to create pubsub: %w", err)
	}

	deploymentTopic, err := pubSub.Join(topicName)
	if err != nil {
		return fmt.Errorf("failed to join deployment topic: %w", err)
	}

	// Subscribe to deployment topic and handle requests concurrently
	deploymentSub, err := deploymentTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to deployment topic: %w", err)
	}

	responseTopic, err := pubSub.Join(topicName + "-response")
	if err != nil {
		return fmt.Errorf("failed to join deployment response topic: %w", err)
	}

	// Subscribe to deployment response topic
	responseSub, err := responseTopic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to deployment response topic: %w", err)
	}

	jobs := job.New(
		node,
		deploymentTopic,
		deploymentSub,
		responseTopic,
		responseSub,
	)
	go jobs.HandleDeploymentRequest(ctx)
	go jobs.HandleDeploymentResponse(ctx)

	// Create and run the REST API
	API := api.NewApi(P2P, jobs)
	return API.Run(port)
}
