package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"

	"nunet/app/api"
	"nunet/app/job"
	"nunet/app/p2p"
)

const (
	defaultTopicName = "container-deployment-12223-nnddd" // Topic for deployment messages
	defaultPort      = 8080                               // REST API port
)

func main() {
	// Create a new context
	ctx := context.Background()

	// Read environment variables for configuration
	topicName := getEnvOrDefault("TOPIC_NAME", defaultTopicName)
	port, err := getPortFromEnv(defaultPort)
	if err != nil {
		log.Fatal("failed to parse port: %w", err)
	}

	// Create a new libp2p host
	node, err := libp2p.New(
		libp2p.FallbackDefaults,
	)
	if err != nil {
		log.Fatal("failed to create libp2p host: %w", err)
	}
	defer node.Close()

	// Print host information
	printHostInfo(node)

	// Create a new P2P instance
	P2P, err := p2p.NewP2P(node)
	if err != nil {
		log.Fatal("failed to create P2P instance: %w", err)
	}

	// Discover peers for communication
	if err := P2P.DiscoverPeers(ctx, topicName); err != nil {
		log.Fatal("failed to discover peers: %w", err)
	}

	// Create pubsub instance and join topic
	pubSub, err := pubsub.NewGossipSub(ctx, node)
	if err != nil {
		log.Fatal("failed to create pubsub: %w", err)
	}

	deploymentTopic, err := pubSub.Join(topicName)
	if err != nil {
		log.Fatal("failed to join deployment topic: %w", err)
	}

	// Subscribe to deployment topic and handle requests concurrently
	deploymentSub, err := deploymentTopic.Subscribe()
	if err != nil {
		log.Fatal("failed to subscribe to deployment topic: %w", err)
	}

	responseTopic, err := pubSub.Join(topicName + "-response")
	if err != nil {
		log.Fatal("failed to join deployment response topic: %w", err)
	}

	// Subscribe to deployment response topic
	responseSub, err := responseTopic.Subscribe()
	if err != nil {
		log.Fatal("failed to subscribe to deployment response topic: %w", err)
	}

	jobs := job.NewJob(node, deploymentTopic, deploymentSub, responseTopic, responseSub)
	go jobs.HandleDeploymentRequest(ctx)
	go jobs.HandleDeploymentResponse(ctx)

	// Create and run the REST API
	API := api.NewApi(P2P, jobs)
	log.Fatal(API.Run(port))
}

func getEnvOrDefault(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getPortFromEnv(defaultValue int) (int, error) {
	if value := os.Getenv("PORT"); value != "" {
		return strconv.Atoi(value)
	}
	return defaultValue, nil
}

func printHostInfo(host host.Host) {
	fmt.Println("Host ID:", host.ID())
	for _, addr := range host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, host.ID())
	}
}
