package app

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

type ApiDeployRequest struct {
	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`
}

type ApiAddPeerRequest struct {
	Address string `json:"address"`
}

type DeployRequest struct {
	SourcePeerID peer.ID  `json:"sourcePeerID"`
	SourceAddrs  []string `json:"sourceAddrs"`

	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`
}
