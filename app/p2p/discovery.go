package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	multiaddr "github.com/multiformats/go-multiaddr"
)

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
