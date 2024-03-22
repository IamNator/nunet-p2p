package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht" // for peer discovery
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
)

func AddPeer(ctx context.Context, addr string, h host.Host) error {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("error creating multiaddr: %s", err)
	}

	fmt.Printf("Adding Peer address: %s\n", ma)
	peerinfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return fmt.Errorf("error getting peerinfo: %s", err)
	}

	if err := h.Connect(ctx, *peerinfo); err != nil {
		return fmt.Errorf("error connecting to peer: %s", err)
	}

	fmt.Printf("Connected to peer: %s\n", peerinfo.ID)
	return nil
}

func DiscoverPeers(ctx context.Context, h host.Host, topicName string) error {
	kademliaDHT, err := initDHT(ctx, h)
	if err != nil {
		return fmt.Errorf("error initializing dht: %s", err)
	}
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, topicName) // Advertise the host's address
	go refreshPeers(routingDiscovery, h, topicName)   // Refresh peers periodically
	return nil
}

func refreshPeers(routingDiscovery *drouting.RoutingDiscovery, h host.Host, topicName string) {

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
			} else {
				fmt.Println("Connected to:", peer.ID)
				anyConnected = true
			}
		}

		time.Sleep(time.Second)

		if anyConnected {
			fmt.Println("Peer discovery complete")
			time.Sleep(time.Minute * 10) // Adjust peer discovery interval as needed
			anyConnected = false
		}
	}

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
