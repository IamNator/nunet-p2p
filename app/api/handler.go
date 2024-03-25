package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin" // for message broadcasting

	"nunet/app/shared"
	"nunet/pkg"
)

// handleHealthRequest returns health information about the node
func (a *api) handleHealthRequest(c *gin.Context) {
	availableCompute, err := pkg.GetComputeAvailable()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error getting compute availability",
			"details": err.Error(),
		})
		return
	}

	addrs, err := a.P2P.ListAddresses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error getting addresses",
			"details": err.Error(),
		})
		return
	}

	connectedPeers := a.Job.ListPeers()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Healthy",
		"data": gin.H{
			"id":        a.P2P.PeerID(),
			"addresses": addrs,
			"peers":     connectedPeers,
			"num_peers": len(connectedPeers),
			"network":   "libp2p",
			"cpu":       availableCompute.FreeCPUCores,
			"ram":       availableCompute.FreeRAM,
			"total_cpu": availableCompute.TotalCPUCores,
			"total_ram": availableCompute.TotalRAM,
			"cpu_model": availableCompute.TotalCPUModel,
			"cpu_ghz":   availableCompute.ToalCPUGhz,
		},
	})
}

// handleDeploymentRequest handles incoming deployment requests
func (a *api) handleDeploymentRequest(c *gin.Context) {
	var request shared.ApiDeployRequest

	// Decode request body and handle bad request
	if err := json.NewDecoder(c.Request.Body).Decode(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// validate request
	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Check for available peers and handle no peers scenario
	peers := a.Job.ListPeers()
	if len(peers) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "No peers available to deploy program",
			"details": "Ensure there are other peers listening on the deployment topic",
		})
		return
	}

	fmt.Printf("Received api request: %s %s\n", request.Program, strings.Join(request.Arguments, " "))

	addrs, err := a.P2P.ListAddresses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error getting addresses",
			"details": err.Error(),
		})
		return
	}

	// Publish deployment request to pubsub topic
	if err := a.Job.PublishDeploymentRequest(context.Background(), shared.DeployRequest{
		SourcePeerID: a.P2P.PeerID().String(),
		SourceAddrs:  addrs,
		Program:      request.Program,
		Arguments:    request.Arguments,
		TargetPeerID: peers[0].String(), // send to first peer
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   "Error publishing request",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Job request sent",
	})
}

func (a api) handleAddPeerRequest(c *gin.Context) {
	var request shared.ApiAddPeerRequest
	// Decode request body and handle bad request
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// validate request
	if err := request.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Add peer to the network
	if err := a.P2P.AddPeer(context.Background(), request.Address); err != nil {
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
