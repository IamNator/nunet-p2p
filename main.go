package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	kad "github.com/libp2p/go-libp2p-kad-dht"   // for peer discovery
	pubsub "github.com/libp2p/go-libp2p-pubsub" // for message broadcasting
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp" // example transport
)

const (
	topicName = "container-deployment" // Topic for deployment messages
)

type deployRequest struct {
	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`
}

func main() {
	// Create a libp2p host
	ctx := context.Background()
	bb, _ := os.ReadFile("host.key")
	privKey, err := crypto.UnmarshalPrivateKey(bb) // Replace with key generation
	if err != nil {
		panic(err)
	}
	opts := libp2p.FallbackDefaults // Adjust transport options as needed
	// Add TCP transport for example
	opts = libp2p.ChainOptions(
		opts,
		libp2p.Identity(privKey),
		libp2p.Transport(tcp.NewTCPTransport),
	)

	host, err := libp2p.New(opts)
	if err != nil {
		panic(err)
	}
	defer host.Close()

	// Advertise the topic

	pubSub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	// Create a new topic
	topic, err := pubSub.Join(topicName)
	if err != nil {
		panic(err)
	}

	// Subscribe to the topic
	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}

	// Process incoming messages
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				fmt.Println("Error reading message:", err)
				continue
			}
			if isSender(ctx, host, msg.GetFrom()) {
				continue
			}
			processDeploymentRequest(msg.GetData())
		}
	}()

	// Peer discovery (optional)
	kademliaDHT, err := kad.New(ctx, host)
	if err != nil {
		panic(err)
	}

	if err := kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	go func() {
		for {
			peerID, err := kademliaDHT.GetClosestPeers(ctx, topicName)
			if err != nil {
				fmt.Println("Error finding peers:", err)
				continue
			}
			for _, p := range peerID {
				if p == host.ID() {
					continue
				}
				fmt.Println("Connecting to peer:", p)
				peer, err := kademliaDHT.FindPeer(ctx, p)
				if err != nil {
					fmt.Println("Error finding peer:", err)
					continue
				}
				host.Connect(ctx, peer)
			}
			time.Sleep(time.Minute) // Adjust discovery interval as needed
		}
	}()

	// Advertise the host's address
	fmt.Println("Host ID:", host.ID())
	fmt.Println("Host address:", host.Addrs())

	// Advertise the host's address

	// REST API for deployment requests
	http.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request deployRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		requestBytes, err := json.Marshal(request)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Publish deployment request to pubsub topic
		if err := topic.Publish(ctx, requestBytes); err != nil {
			fmt.Println("Error publishing deployment request:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})

	// Start listening for incoming connections
	fmt.Println("Listening for deployment requests...")
	select {}
}

func processDeploymentRequest(data []byte) {
	var request deployRequest
	err := json.Unmarshal(data, &request)
	if err != nil {
		fmt.Println("Error unmarshalling request:", err)
		return
	}
	// ... (container deployment logic using program and arguments)
	fmt.Printf("Deploying container: %s %s\n", request.Program, strings.Join(request.Arguments, " "))
}

func isSender(ctx context.Context, host host.Host, p peer.ID) bool {
	// Implement logic to check if the peer is the sender (e.g., compare host keys)
	return false // Replace with actual implementation
}
