package main

import (
	"context"
	"fmt"
	"nunet/app"
	"os"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	defaultTopicName = "container-deployment" // Topic for deployment messages
	port             = 8080                   // REST API port
)

func main() {
	// Create a new context
	ctx := context.Background()

	var TopicName string = defaultTopicName
	if topic := os.Getenv("TOPIC_NAME"); topic != "" {
		TopicName = topic
	}

	// Create a new libp2p host
	host, err := libp2p.New(libp2p.FallbackDefaults)
	if err != nil {
		panic(fmt.Errorf("failed to create libp2p host: %w", err))
	}
	defer host.Close()

	// Print host ID and listening addresses
	fmt.Println("Host ID:", host.ID())
	for _, addr := range host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, host.ID())
	}

	// Discover peers for peer-to-peer communication
	if err := app.DiscoverPeers(ctx, host, TopicName); err != nil {
		panic(fmt.Errorf("failed to discover peers: %w", err))
	}

	// Advertise the topic for message broadcasting
	pubSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(fmt.Errorf("failed to create pubsub: %w", err))
	}

	// Join the deployment topic
	deploymentTopic, err := pubSub.Join(TopicName)
	if err != nil {
		panic(fmt.Errorf("failed to join deployment topic: %w", err))
	}

	// Subscribe to the deployment topic
	sub, err := deploymentTopic.Subscribe()
	if err != nil {
		panic(fmt.Errorf("failed to subscribe to deployment topic: %w", err))
	}

	// Handle deployment requests in a separate goroutine
	go app.HandleDeploymentRequest(ctx, host, sub)

	// Create a new REST API and run it
	api := app.NewApi(host, deploymentTopic)
	api.Run(port)
}
