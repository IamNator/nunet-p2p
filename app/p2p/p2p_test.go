package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	multiaddr "github.com/multiformats/go-multiaddr"

	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListAddresses(t *testing.T) {
	// mockHost := &mockHost{}
	// mockHost.On("ID").Return("12D3KooWBHCqYQ3CQQrmTMXDLgxiR5paj18pjBiTkzn8ZVGXMrd7")

	// addrs, err := multiaddr.NewMultiaddr("/ip4/172.31.10.0/tcp/43047/p2p/12D3KooWBHCqYQ3CQQrmTMXDLgxiR5paj18pjBiTkzn8ZVGXMrd7")
	// if !assert.NoError(t, err) {
	// 	t.FailNow()
	// }
	// mockHost.On("Addrs").Return([]multiaddr.Multiaddr{addrs})

	// p2pInstance := &P2P{Host: mockHost}

	// addresses, err := p2pInstance.ListAddresses()
	// if !assert.NoError(t, err) {
	// 	t.FailNow()
	// }
	// assert.Len(t, addresses, 1)
	// assert.Equal(t, "/ip4/172.31.10.0/tcp/43047/p2p/12D3KooWBHCqYQ3CQQrmTMXDLgxiR5paj18pjBiTkzn8ZVGXMrd7", addresses[0])
}

type mockHost struct {
	host.Host
	mock.Mock
}

func (m *mockHost) ID() peer.ID {
	args := m.Called()
	return peer.ID(args.String(0))
}

func (m *mockHost) Addrs() []multiaddr.Multiaddr {
	args := m.Called()
	return args.Get(0).([]multiaddr.Multiaddr)
}

func (m *mockHost) Connect(ctx context.Context, pi peer.AddrInfo) error {
	args := m.Called(pi)
	return args.Error(0)
}

type mockRoutingDiscovery struct {
	*drouting.RoutingDiscovery
	mock.Mock
}

func (m *mockRoutingDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	args := m.Called(ns, opts)
	return args.Get(0).(time.Duration), args.Error(1)
}
