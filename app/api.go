package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin" // for message broadcasting
	"github.com/libp2p/go-libp2p/core/peer"

	"nunet/pkg"
)

type PeerOperations interface {
	AddPeer(ctx context.Context, addr string) error
	DiscoverPeers(ctx context.Context, topicName string) error
	ListAddresses() []string
	PeerID() peer.ID
}

type JobOperations interface {
	PublishDeploymentRequest(ctx context.Context, request DeployRequest) error
	HandleDeploymentRequest(ctx context.Context)
	HandleDeploymentResponse(ctx context.Context)
	ListPeers() []peer.ID
}

type api struct {
	P2P PeerOperations
	Job JobOperations
}

func NewApi(p2p PeerOperations, job JobOperations) *api {
	return &api{
		P2P: p2p,
		Job: job,
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

	connectedPeers := a.Job.ListPeers()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Healthy",
		"data": gin.H{
			"id":        a.P2P.PeerID(),
			"addresses": a.P2P.ListAddresses(),
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

	//select a random peer that is listening on thesame topic to deploy the program
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

	// Publish deployment request to pubsub topic
	if err := a.Job.PublishDeploymentRequest(context.Background(), DeployRequest{
		SourcePeerID: a.P2P.PeerID().String(),
		SourceAddrs:  a.P2P.ListAddresses(),
		Program:      request.Program,
		Arguments:    request.Arguments,
		TargetPeerID: peers[0].String(),
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
