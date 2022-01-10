package npm

import (
	"context"

	npmconfig "github.com/Azure/azure-container-networking/npm/config"
	"github.com/Azure/azure-container-networking/npm/pkg/transport"
)

type NetworkPolicyDataplane struct {
	ctx    context.Context
	config npmconfig.Config
	client *transport.DataplaneEventsClient
}

func NewNetworkPolicyDataplane(
	ctx context.Context,
	config npmconfig.Config,
) *NetworkPolicyDataplane {
	return &NetworkPolicyDataplane{
		config: config,
	}
}

func (n *NetworkPolicyDataplane) Start(config npmconfig.Config, stopCh <-chan struct{}) error {
	_ = n.client.EventsChannel()
	// TODO (nitishm): Start the goalstate processor which reads events from the
	// channel and sends them to the goalstate manager.

	return n.client.Start(n.ctx, stopCh)
}
