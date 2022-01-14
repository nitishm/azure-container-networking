// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"encoding/json"
	"fmt"

	npmconfig "github.com/Azure/azure-container-networking/npm/config"
	controllersv2 "github.com/Azure/azure-container-networking/npm/pkg/controlplane/controllers/v2"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane"
	"github.com/Azure/azure-container-networking/npm/pkg/transport"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type NetworkPolicyServer struct {
	config npmconfig.Config

	// tm is the transport layer (gRPC) manager/server
	tm *transport.Manager

	// Informers are the Kubernetes Informer
	// https://pkg.go.dev/k8s.io/client-go/informers
	Informers

	// Controllers for handling Kubernetes resource watcher events
	K8SControllersV2

	// Azure-specific variables
	AzureConfig
}

func NewNetworkPolicyServer(
	config npmconfig.Config,
	informerFactory informers.SharedInformerFactory,
	mgr *transport.Manager,
	dp dataplane.GenericDataplane,
	npmVersion string,
	k8sServerVersion *version.Info,
) (*NetworkPolicyServer, error) {
	klog.Infof("API server version: %+v AI metadata %+v", k8sServerVersion, aiMetadata)

	if informerFactory == nil {
		return nil, fmt.Errorf("informer factory is nil")
	}

	if mgr == nil {
		return nil, fmt.Errorf("transport manager is nil")
	}

	if dp == nil {
		return nil, fmt.Errorf("dataplane is nil")
	}

	if k8sServerVersion == nil {
		return nil, fmt.Errorf("k8s server version is nil")
	}

	n := &NetworkPolicyServer{
		config: config,
		tm:     mgr,
		Informers: Informers{
			informerFactory: informerFactory,
			podInformer:     informerFactory.Core().V1().Pods(),
			nsInformer:      informerFactory.Core().V1().Namespaces(),
			npInformer:      informerFactory.Networking().V1().NetworkPolicies(),
		},
		AzureConfig: AzureConfig{
			k8sServerVersion: k8sServerVersion,
			NodeName:         GetNodeName(),
			version:          npmVersion,
			TelemetryEnabled: true,
		},
	}

	n.npmNamespaceCacheV2 = &controllersv2.NpmNamespaceCache{NsMap: make(map[string]*controllersv2.Namespace)}
	n.podControllerV2 = controllersv2.NewPodController(n.podInformer, dp, n.npmNamespaceCacheV2)
	n.namespaceControllerV2 = controllersv2.NewNamespaceController(n.nsInformer, dp, n.npmNamespaceCacheV2)
	n.netPolControllerV2 = controllersv2.NewNetworkPolicyController(n.npInformer, dp)

	return n, nil
}

func (n *NetworkPolicyServer) MarshalJSON() ([]byte, error) {
	m := map[CacheKey]json.RawMessage{}

	var npmNamespaceCacheRaw []byte
	var err error
	npmNamespaceCacheRaw, err = json.Marshal(n.npmNamespaceCacheV2)

	if err != nil {
		return nil, errors.Errorf("%s: %v", errMarshalNPMCache, err)
	}
	m[NsMap] = npmNamespaceCacheRaw

	var podControllerRaw []byte
	podControllerRaw, err = json.Marshal(n.podControllerV2)

	if err != nil {
		return nil, errors.Errorf("%s: %v", errMarshalNPMCache, err)
	}
	m[PodMap] = podControllerRaw

	nodeNameRaw, err := json.Marshal(n.NodeName)
	if err != nil {
		return nil, errors.Errorf("%s: %v", errMarshalNPMCache, err)
	}
	m[NodeName] = nodeNameRaw

	npmCacheRaw, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Errorf("%s: %v", errMarshalNPMCache, err)
	}

	return npmCacheRaw, nil
}

func (n *NetworkPolicyServer) GetAppVersion() string {
	return n.version
}

func (n *NetworkPolicyServer) Start(config npmconfig.Config, stopCh <-chan struct{}) error {
	// Starts all informers manufactured by n's informerFactory.
	n.informerFactory.Start(stopCh)

	// Wait for the initial sync of local cache.
	if !cache.WaitForCacheSync(stopCh, n.podInformer.Informer().HasSynced) {
		return fmt.Errorf("Pod informer failed to sync")
	}

	if !cache.WaitForCacheSync(stopCh, n.nsInformer.Informer().HasSynced) {
		return fmt.Errorf("Namespace informer failed to sync")
	}

	if !cache.WaitForCacheSync(stopCh, n.npInformer.Informer().HasSynced) {
		return fmt.Errorf("Network policy informer failed to sync")
	}

	// start v2 NPM controllers after synced
	go n.podControllerV2.Run(stopCh)
	go n.namespaceControllerV2.Run(stopCh)
	go n.netPolControllerV2.Run(stopCh)

	// start the transport layer (gRPC) server
	// We block the main thread here until the server is stopped.
	// This is unlike the other start methods in this package, which returns nil
	// and blocks in the main thread during command invocation through the select {}
	// statement.
	return n.tm.Start(stopCh)
}
