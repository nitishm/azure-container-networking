// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"context"
	"fmt"

	npmconfig "github.com/Azure/azure-container-networking/npm/config"
	"github.com/Azure/azure-container-networking/npm/pkg/controlplane/goalstateprocessor"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane"
	"github.com/Azure/azure-container-networking/npm/pkg/transport"
)

type NetworkPolicyDaemon struct {
	ctx     context.Context
	config  npmconfig.Config
	client  *transport.DataplaneEventsClient
	version string
	gsp     *goalstateprocessor.GoalStateProcessor
}

func NewNetworkPolicyDaemon(
	ctx context.Context,
	config npmconfig.Config,
	dp dataplane.GenericDataplane,
	gsp *goalstateprocessor.GoalStateProcessor,
	client *transport.DataplaneEventsClient,
	npmVersion string,
) (*NetworkPolicyDaemon, error) {

	if dp == nil {
		return nil, fmt.Errorf("dataplane is nil")
	}

	return &NetworkPolicyDaemon{
		ctx:     ctx,
		config:  config,
		gsp:     gsp,
		client:  client,
		version: npmVersion,
	}, nil
}

func (n *NetworkPolicyDaemon) Start(config npmconfig.Config, stopCh <-chan struct{}) {
	go n.gsp.Start(stopCh)
	go n.client.Start(stopCh)
	<-stopCh
}

func (n *NetworkPolicyDaemon) GetAppVersion() string {
	return n.version
}
