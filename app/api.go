package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	pubsub "github.com/libp2p/go-libp2p-pubsub" // for message broadcasting
	"github.com/libp2p/go-libp2p/core/host"
)

func HandleUserInput(ctx context.Context, port int, host host.Host, deploymentTopic *pubsub.Topic) {

	router := gin.Default()

	// attach cors middleware
	router.Use(CorsMiddleware())

	// REST API for deployment requests
	router.POST("/deploy", func(c *gin.Context) {

		var request ApiDeployRequest
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var addrs []string
		for _, addr := range host.Addrs() {
			addrs = append(addrs, addr.String())
		}
		requestBytes, err := json.Marshal(DeployRequest{
			SourcePeerID: host.ID(),
			SourceAddrs:  addrs,
			Program:      request.Program,
			Arguments:    request.Arguments,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshalling deployment request", "details": err.Error()})
			return
		}

		fmt.Printf("Received deployment request: %s %s\n", request.Program, strings.Join(request.Arguments, " "))

		// Publish deployment request to pubsub topic
		if err := deploymentTopic.Publish(ctx, requestBytes); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error publishing deployment request", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "Deployment request sent"})
	})

	// Start listening for incoming connections
	fmt.Println("Listening for deployment requests...")
retry:
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		if strings.Contains(err.Error(), "already in use") {
			port = port + 1
			fmt.Println("Port already in use, trying to listen on port ", port)
			goto retry
		}
		println("Error starting server:", err.Error())
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
