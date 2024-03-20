package main

import (
	"context"
	"fmt"
	"nunet/app"

	"github.com/libp2p/go-libp2p"               // for peer discovery
	pubsub "github.com/libp2p/go-libp2p-pubsub" // for message broadcasting
)

const (
	topicName string = "container-deployment-dafiiiid121" // Topic for deployment messages
)

var (
	port = 8080 // REST API port
)

var (
	PeerAvailability = make(map[string]map[string]any) // map of peer availability (cpu, ram)
)

func main() {
	ctx := context.Background() // Create a libp2p host
	host, err := libp2p.New(libp2p.FallbackDefaults)
	if err != nil {
		panic(err)
	}
	defer host.Close()
	
	fmt.Println("Host ID:", host.ID())

	app.DiscoverPeers(ctx, host, topicName)

	// Advertise the topic
	pubSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	// Create a new deploymentTopic
	deploymentTopic, err := pubSub.Join(topicName)
	if err != nil {
		panic(err)
	}

	// Subscribe to the topic
	sub, err := deploymentTopic.Subscribe()
	if err != nil {
		panic(err)
	}

	// Handle deployment requests
	go app.HandleDeploymentRequest(ctx, host, sub)

	// Handle user input
	app.HandleUserInput(ctx, port, host, deploymentTopic)
}
