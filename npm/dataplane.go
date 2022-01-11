package npm

import (
	"context"
	"os"
	"strconv"

	npmconfig "github.com/Azure/azure-container-networking/npm/config"
	"github.com/Azure/azure-container-networking/npm/pkg/transport"
)

const (
	POD_NAME_ENV  = "POD_NAME"
	NODE_NAME_ENV = "NODE_NAME"
)

type NetworkPolicyDataplane struct {
	ctx    context.Context
	config npmconfig.Config
	client *transport.DataplaneEventsClient
}

func NewNetworkPolicyDataplane(
	ctx context.Context,
	config npmconfig.Config,
) (*NetworkPolicyDataplane, error) {

	// FIXME (nitishm): Where do these come from? Should we p
	pod := os.Getenv(POD_NAME_ENV)
	node := os.Getenv(NODE_NAME_ENV)

	addr := config.Transport.Address + ":" + strconv.Itoa(config.Transport.Port)

	client, err := transport.NewDataplaneEventsClient(ctx, pod, node, addr)
	if err != nil {
		return nil, err
	}

	return &NetworkPolicyDataplane{
		ctx:    ctx,
		config: config,
		client: client,
	}, nil
}

func (n *NetworkPolicyDataplane) Start(config npmconfig.Config, stopCh <-chan struct{}) error {
	_ = n.client.EventsChannel()
	// TODO (nitishm): Start the goalstate processor which reads events from the
	// channel and sends them to the goalstate manager.

	return n.client.Start(n.ctx, stopCh)
}
