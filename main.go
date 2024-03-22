package main

import (
	"context"
	"fmt"
	"log"
	"nunet/app"
	"os"
	"strconv"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
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
		log.Fatal(fmt.Errorf("failed to parse port: %w", err))
	}

	// Create a new libp2p host
	host, err := libp2p.New(libp2p.FallbackDefaults)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create libp2p host: %w", err))
	}
	defer host.Close()

	// Print host information
	printHostInfo(host)

	// Discover peers for communication
	if err := app.DiscoverPeers(ctx, host, topicName); err != nil {
		log.Fatal(fmt.Errorf("failed to discover peers: %w", err))
	}

	// Create pubsub instance and join topic
	pubSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create pubsub: %w", err))
	}
	deploymentTopic, err := pubSub.Join(topicName)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to join deployment topic: %w", err))
	}

	// Subscribe to deployment topic and handle requests concurrently
	sub, err := deploymentTopic.Subscribe()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to subscribe to deployment topic: %w", err))
	}
	go app.HandleDeploymentRequest(ctx, host, sub)

	// Create and run the REST API
	api := app.NewApi(host, deploymentTopic)
	log.Fatal(api.Run(port)) // Use port number
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
