package p2p

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddPeer(t *testing.T) {
	mockHost := &mockHost{}
	mockHost.On("Connect", mock.Anything).Return(nil)

	p2pInstance := &P2P{Host: mockHost}

	err := p2pInstance.AddPeer(context.Background(), "/ip4/127.0.0.1/tcp/4002/p2p/12D3KooWBHCqYQ3CQQrmTMXDLgxiR5paj18pjBiTkzn8ZVGXMrd7")

	assert.NoError(t, err)
}

func TestDiscoverPeers(t *testing.T) {
	mockRoutingDiscovery := &mockRoutingDiscovery{}
	mockRoutingDiscovery.On("Advertise", "topicName").Return(nil)

	p2pInstance := &P2P{routingDiscovery: mockRoutingDiscovery}

	err := p2pInstance.DiscoverPeers(context.Background(), "topicName")

	assert.NoError(t, err)
}
