package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht" // for peer discovery
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
)

type P2P struct {
	Host             host.Host
	routingDiscovery *drouting.RoutingDiscovery
}

func NewP2P(h host.Host) (*P2P, error) {
	kademliaDHT, err := initDHT(context.Background(), h)
	if err != nil {
		return nil, fmt.Errorf("error initializing dht: %s", err)
	}

	return &P2P{
		Host:             h,
		routingDiscovery: drouting.NewRoutingDiscovery(kademliaDHT),
	}, nil
}

func (p *P2P) PeerID() peer.ID {
	return p.Host.ID()
}

func (p *P2P) ListAddresses() ([]string, error) {

	peerInfo := peerstore.AddrInfo{
		ID:    p.Host.ID(),
		Addrs: p.Host.Addrs(),
	}

	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		fmt.Println("Error getting addresses:", err)
		return nil, err
	}

	var results []string
	for _, addr := range addrs {
		results = append(results, addr.String())
	}
	return results, nil
}

func (p *P2P) AddPeer(ctx context.Context, addr string) error {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return fmt.Errorf("error creating multiaddr: %s", err)
	}

	fmt.Printf("Adding Peer address: %s\n", ma)
	peerinfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		return fmt.Errorf("error getting peerinfo: %s", err)
	}

	if err := p.Host.Connect(ctx, *peerinfo); err != nil {
		return fmt.Errorf("error connecting to peer: %s", err)
	}

	fmt.Printf("Connected to peer: %s\n", peerinfo.ID)
	return nil
}

func (p *P2P) DiscoverPeers(ctx context.Context, topicName string) error {
	dutil.Advertise(ctx, p.routingDiscovery, topicName) // Advertise the host's address
	go p.refreshPeers(topicName)                        // Refresh peers periodically
	return nil
}

func (p *P2P) refreshPeers(topicName string) {

	// Look for others who have announced and attempt to connect to them
	anyConnected := false
	ctx := context.Background()
	for {
		fmt.Println("Searching for peers...")
		peerChan, err := p.routingDiscovery.FindPeers(ctx, topicName)
		if err != nil {
			fmt.Println("Error finding peers:", err)
			continue
		}

		for peer := range peerChan {
			if peer.ID == p.Host.ID() {
				continue // No self connection
			}
			if err := p.Host.Connect(ctx, peer); err != nil {
				fmt.Printf("Failed connecting to %s, error: %s\n;\n", peer.ID, err.Error())
			} else {
				fmt.Println("Connected to:", peer.ID)
				anyConnected = true
			}
		}

		time.Sleep(time.Minute / 3)

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
