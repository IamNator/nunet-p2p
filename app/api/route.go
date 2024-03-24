package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin" // for message broadcasting
	"github.com/libp2p/go-libp2p/core/peer"

	"nunet/app/shared"
	"nunet/pkg"
)

// PeerOperations defines the functionalities for peer management
type PeerOperations interface {
	AddPeer(ctx context.Context, addr string) error
	DiscoverPeers(ctx context.Context, topicName string) error
	ListAddresses() ([]string, error)
	PeerID() peer.ID
}

// JobOperations defines the functionalities for job management
type JobOperations interface {
	PublishDeploymentRequest(ctx context.Context, request shared.DeployRequest) error
	HandleDeploymentRequest(ctx context.Context)
	HandleDeploymentResponse(ctx context.Context)
	ListPeers() []peer.ID
}

// api struct holds references to PeerOperations and JobOperations services
type api struct {
	P2P PeerOperations
	Job JobOperations
}

// NewApi creates a new instance of the api struct
func NewApi(p2p PeerOperations, job JobOperations) *api {
	return &api{
		P2P: p2p,
		Job: job,
	}
}

// Run starts the api server and listens for incoming connections
func (a *api) Run(port int) error {
	router := gin.Default()
	router.Use(pkg.CorsMiddleware()) // attach cors middleware

	router.GET("/health", a.handleHealthRequest)
	router.POST("/peer", a.handleAddPeerRequest)
	router.POST("/deploy", a.handleDeploymentRequest)

	// Start listening for incoming connections with port handling logic
	fmt.Println("Listening for deployment requests...")
retry:
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		if strings.Contains(err.Error(), "already in use") {
			port++
			fmt.Printf("Port %d already in use, retrying with port %d\n", port-1, port)
			goto retry
		}
		return err
	}

	return nil
}
