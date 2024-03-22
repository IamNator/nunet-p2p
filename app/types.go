package app

type ApiDeployRequest struct {
	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`
}

type ApiAddPeerRequest struct {
	Address string `json:"address"`
}

type DeployRequest struct {
	SourcePeerID string   `json:"source_peer_id"`
	SourceAddrs  []string `json:"source_addrs"`
	TargetPeerID string   `json:"target_peer_id"`

	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`
}

type DeployResponse struct {
	Success      bool     `json:"success"`
	SourcePeerID string   `json:"source_peer_id"`
	SourceAddrs  []string `json:"source_addrs"`

	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`

	PID          int      `json:"pid"`
	TargetPeerID string   `json:"target_peer_id"`
	TargetAddrs  []string `json:"target_addrs"`
}
