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

	"nunet/pkg"
)

type api struct {
	Host            host.Host
	DeploymentTopic *pubsub.Topic
}

func NewApi(host host.Host, deploymentTopic *pubsub.Topic) *api {
	return &api{
		Host:            host,
		DeploymentTopic: deploymentTopic,
	}
}

func (a api) Run(port int) error {

	router := gin.Default()
	router.Use(pkg.CorsMiddleware()) // attach cors middleware

	router.GET("/health", a.handleHealthRequest)
	router.POST("/peer", a.handleAddPeerRequest)
	router.POST("/deploy", a.handleDeploymentRequest)

	// Start listening for incoming connections
	fmt.Println("Listening for deployment requests...")
retry:
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		if strings.Contains(err.Error(), "already in use") {
			port = port + 1
			fmt.Printf("Port %d already in use, retrying with port %d\n", port-1, port)
			goto retry
		}
		return err
	}

	return nil
}

func (a api) handleHealthRequest(c *gin.Context) {
	cpuAvailable, ramAvailable, err := pkg.GetComputeAvailable()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error getting compute availability",
			"details": err.Error(),
		})
		return
	}

	var address []string
	for _, addr := range a.Host.Addrs() {
		address = append(address, fmt.Sprintf("%s/p2p/%s", addr, a.Host.ID().String()))
	}

	connectedPeers := a.Host.Network().Peers()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Healthy",
		"data": gin.H{
			"id":        a.Host.ID().String(),
			"addresses": address,
			"peers":     connectedPeers,
			"num_peers": len(connectedPeers),
			"network":   "libp2p",
			"cpu":       cpuAvailable,
			"ram":       ramAvailable,
		},
	})
}

func (a api) handleDeploymentRequest(c *gin.Context) {

	var request ApiDeployRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	var addrs []string
	for _, addr := range a.Host.Addrs() {
		addrs = append(addrs, addr.String())
	}

	//select a random peer that is listening on thesame topic to deploy the program
	peers := a.DeploymentTopic.ListPeers()
	if len(peers) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "No peers available to deploy program",
			"details": "Ensure there are other peers listening on the deployment topic",
		})
		return
	}

	requestBytes, err := json.Marshal(DeployRequest{
		SourcePeerID: a.Host.ID().String(),
		SourceAddrs:  addrs,
		Program:      request.Program,
		Arguments:    request.Arguments,
		TargetPeerID: peers[0].String(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error marshalling deployment request",
			"details": err.Error(),
		})
		return
	}

	fmt.Printf("Received deployment request: %s %s\n", request.Program, strings.Join(request.Arguments, " "))

	// Publish deployment request to pubsub topic
	if err := a.DeploymentTopic.Publish(context.Background(), requestBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error publishing deployment request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Deployment request sent",
	})
}

func (a api) handleAddPeerRequest(c *gin.Context) {
	var request ApiAddPeerRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	if err := AddPeer(context.Background(), request.Address, a.Host); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error adding peer",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Peer added",
	})
}
