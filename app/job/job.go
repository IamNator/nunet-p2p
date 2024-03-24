package job

import (
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Job struct {
	Host                    host.Host
	DeploymentTopic         *pubsub.Topic
	DeploymentSub           *pubsub.Subscription
	DeploymentResponseTopic *pubsub.Topic
	DeploymentResponseSub   *pubsub.Subscription
}

// New creates a new Job instance
func New(
	h host.Host,
	deploymentTopic *pubsub.Topic,
	deploymentSub *pubsub.Subscription,
	deploymentResponseTopic *pubsub.Topic,
	deploymentResponseSub *pubsub.Subscription,
) *Job {
	return &Job{
		Host:                    h,
		DeploymentTopic:         deploymentTopic,
		DeploymentSub:           deploymentSub,
		DeploymentResponseTopic: deploymentResponseTopic,
		DeploymentResponseSub:   deploymentResponseSub,
	}
}

func (j *Job) ListPeers() []peer.ID {
	return j.DeploymentTopic.ListPeers()
}
