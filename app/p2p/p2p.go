package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht" // for peer discovery
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type RoutingDiscovery interface {
	Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error)
	FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error)
}

type P2P struct {
	Host             host.Host
	routingDiscovery RoutingDiscovery
}

func New(h host.Host) (*P2P, error) {
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

	peerInfo := peer.AddrInfo{
		ID:    p.Host.ID(),
		Addrs: p.Host.Addrs(),
	}

	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
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
