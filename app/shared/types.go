package shared

import "fmt"

type ApiDeployRequest struct {
	Program   string   `json:"program"`
	Arguments []string `json:"arguments"`
}

func (a ApiDeployRequest) Validate() error {
	if a.Program == "" {
		return fmt.Errorf("program is required")
	}
	return nil
}

type ApiAddPeerRequest struct {
	Address string `json:"address"`
}

func (a ApiAddPeerRequest) Validate() error {
	if a.Address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
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

	Outputs []string `json:"outputs"`
}
