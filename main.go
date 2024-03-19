package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"   // for peer discovery
	pubsub "github.com/libp2p/go-libp2p-pubsub" // for message broadcasting
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	topicName string = "container-deployment-1" // Topic for deployment messages
)

var (
	port = 8080 // REST API port
)

type deployRequest struct {
	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`
}

var (
	PeerAvailability = make(map[string]map[string]any) // map of peer availability (cpu, ram)
)

func main() {
	// Create a libp2p host
	ctx := context.Background()
	host, err := libp2p.New(libp2p.FallbackDefaults)
	if err != nil {
		panic(err)
	}
	defer host.Close()

	// Advertise the host's address
	fmt.Println("Host ID:", host.ID())

	kademliaDHT, err := initDHT(ctx, host)
	if err != nil {
		fmt.Println("Error initializing DHT:", err)
		panic(err)
	}
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, topicName)

	go discoverPeers(routingDiscovery, host, topicName)

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

	// publish the host cpu and ram availability
	// go publishAvailability(ctx, host, pubSub)

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
			processDeploymentRequest(msg.GetFrom().String(), msg.GetData())
		}
	}()

	// Handle user input
	handleUserInput(ctx, deploymentTopic)
}

func processDeploymentRequest(id string, data []byte) {
	var request deployRequest
	err := json.Unmarshal(data, &request)
	if err != nil {
		fmt.Println("Error unmarshalling request:", err)
		return
	}
	// ... (container deployment logic using program and arguments)
	fmt.Printf("Deploying container: %s %s from %s\n", request.Program, strings.Join(request.Arguments, " "), id)
}

func isSender(_ context.Context, host host.Host, p peer.ID) bool {
	return host.ID() == p
}

func getComputeAvailable() (cpuAvailable int, ramAvailable float64, err error) {
	// Get CPU information
	cpuInfo, err := cpu.Info()
	if err != nil {
		fmt.Println("Error getting CPU information:", err)
		return
	}

	// Print number of logical cores
	fmt.Printf("Logical cores: %d\n", cpuInfo[0].Cores)

	// Get memory information
	vmem, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error getting memory information:", err)
		return
	}

	// Print total RAM in Gigabytes
	totalRAM := float64(vmem.Total) / 1024 / 1024 / 1024
	fmt.Printf("Total RAM: %.2f GB\n", totalRAM)

	return int(cpuInfo[0].Cores), totalRAM, nil
}

func initDHT(ctx context.Context, h host.Host) (*dht.IpfsDHT, error) {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	if err != nil {
		return nil, err
	}
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				fmt.Println("Bootstrap warning:", err)
			} else {
				fmt.Println("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	return kademliaDHT, nil
}

func discoverPeers(routingDiscovery *drouting.RoutingDiscovery, h host.Host, topicName string) {

	// Look for others who have announced and attempt to connect to them
	anyConnected := false
	ctx := context.Background()
	for {
		fmt.Println("Searching for peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, topicName)
		if err != nil {
			fmt.Println("Error finding peers:", err)
			continue
		}

		for peer := range peerChan {
			if peer.ID == h.ID() {
				continue // No self connection
			}
			if err := h.Connect(ctx, peer); err != nil {
				fmt.Printf("Failed connecting to %s, error: %s\n;\n", peer.ID, err.Error())
				//remove the peer from the list of available peers and peer store
				delete(PeerAvailability, peer.ID.String())
				h.Peerstore().ClearAddrs(peer.ID)
			} else {
				fmt.Println("Connected to:", peer.ID)
				anyConnected = true
			}
		}

		if anyConnected {
			fmt.Println("Peer discovery complete")
			time.Sleep(time.Minute * 10) // Adjust peer discovery interval as needed
			anyConnected = false
		}
	}

}

// func publishAvailability(ctx context.Context, host host.Host, pubSub *pubsub.PubSub) {

// 	// Create a new deploymentTopic
// 	availabilityTopic, err := pubSub.Join("availability")
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Subscribe to the topic
// 	sub, err := availabilityTopic.Subscribe()
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Publish availability
// 	go func() {
// 		for {
// 			cpuAvailable, ramAvailable, err := getComputeAvailable()
// 			if err != nil {
// 				continue
// 			}
// 			availability := map[string]interface{}{
// 				"cpu": cpuAvailable,
// 				"ram": ramAvailable,
// 			}
// 			availabilityBytes, err := json.Marshal(availability)
// 			if err != nil {
// 				continue
// 			}
// 			if err := availabilityTopic.Publish(ctx, availabilityBytes); err != nil {
// 				fmt.Println("Error publishing availability:", err)
// 				continue
// 			}
// 			time.Sleep(time.Minute) // Adjust availability update interval as needed
// 		}
// 	}()

// 	// update the availability of the peer

// 	for {
// 		msg, err := sub.Next(ctx)
// 		if err != nil {
// 			fmt.Println("Error reading message:", err)
// 			continue
// 		}

// 		if isSender(ctx, host, msg.GetFrom()) {
// 			continue
// 		}
// 		var availability map[string]interface{}
// 		err = json.Unmarshal(msg.GetData(), &availability)
// 		if err != nil {
// 			fmt.Println("Error unmarshalling availability:", err)
// 			continue
// 		}
// 		PeerAvailability[msg.GetFrom().String()] = availability
// 		fmt.Printf("Peer %s availability: %v\n", msg.GetFrom().String(), availability)
// 	}

// }

func handleUserInput(ctx context.Context, deploymentTopic *pubsub.Topic) {
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

		fmt.Printf("Received deployment request: %s %s\n", request.Program, strings.Join(request.Arguments, " "))

		// Publish deployment request to pubsub topic
		if err := deploymentTopic.Publish(ctx, requestBytes); err != nil {
			fmt.Println("Error publishing deployment request:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})

	// Start listening for incoming connections
	fmt.Println("Listening for deployment requests...")
retry:
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		if strings.Contains(err.Error(), "already in use") {
			port = port + 1
			fmt.Println("Port already in use, trying to listen on port ", port)
			goto retry
		}
		println("Error starting server:", err.Error())
	}
}
