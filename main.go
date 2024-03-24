package main

import (
	"context"
	"log"

	"nunet/app"
	"nunet/pkg"
)

const (
	defaultTopicName = "container-deployment-12223-nnddd" // Topic for deployment messages
	defaultPort      = 8080                               // REST API port
)

func main() {
	// Create a new context
	ctx := context.Background()

	// Read environment variables for configuration
	topicName := pkg.GetEnvOrDefault("TOPIC_NAME", defaultTopicName)
	port := pkg.GetEnvOrDefaultInt("PORT", defaultPort)

	// Run the application
	if err := app.Run(ctx, topicName, port); err != nil {
		log.Fatal("failed to run application: %w", err)
	}
}
