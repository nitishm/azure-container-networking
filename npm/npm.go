// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"encoding/json"
	"fmt"
	"os"

	npmconfig "github.com/Azure/azure-container-networking/npm/config"
	"github.com/Azure/azure-container-networking/npm/ipsm"
	controllersv1 "github.com/Azure/azure-container-networking/npm/pkg/controlplane/controllers/v1"
	controllersv2 "github.com/Azure/azure-container-networking/npm/pkg/controlplane/controllers/v2"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	utilexec "k8s.io/utils/exec"
)

var ErrDataplaneNotInitialized = errors.New("dataplane is not initialized")

// NewNetworkPolicyManager creates a NetworkPolicyManager
func NewNetworkPolicyManager(config npmconfig.Config,
	informerFactory informers.SharedInformerFactory,
	dp dataplane.GenericDataplane,
	exec utilexec.Interface,
	npmVersion string,
	k8sServerVersion *version.Info) *NetworkPolicyManager {
	klog.Infof("API server version: %+v AI metadata %+v", k8sServerVersion, aiMetadata)

	npMgr := &NetworkPolicyManager{
		config: config,
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

	// create v2 NPM specific components.
	if npMgr.config.Toggles.EnableV2NPM {
		npMgr.npmNamespaceCacheV2 = &controllersv2.NpmNamespaceCache{NsMap: make(map[string]*controllersv2.Namespace)}
		npMgr.podControllerV2 = controllersv2.NewPodController(npMgr.podInformer, dp, npMgr.npmNamespaceCacheV2)
		npMgr.namespaceControllerV2 = controllersv2.NewNamespaceController(npMgr.nsInformer, dp, npMgr.npmNamespaceCacheV2)
		// Question(jungukcho): Is config.Toggles.PlaceAzureChainFirst needed for v2?
		npMgr.netPolControllerV2 = controllersv2.NewNetworkPolicyController(npMgr.npInformer, dp)
		return npMgr
	}

	// create v1 NPM specific components.
	npMgr.ipsMgr = ipsm.NewIpsetManager(exec)

	npMgr.npmNamespaceCacheV1 = &controllersv1.NpmNamespaceCache{NsMap: make(map[string]*controllersv1.Namespace)}
	npMgr.podControllerV1 = controllersv1.NewPodController(npMgr.podInformer, npMgr.ipsMgr, npMgr.npmNamespaceCacheV1)
	npMgr.namespaceControllerV1 = controllersv1.NewNameSpaceController(npMgr.nsInformer, npMgr.ipsMgr, npMgr.npmNamespaceCacheV1)
	npMgr.netPolControllerV1 = controllersv1.NewNetworkPolicyController(npMgr.npInformer, npMgr.ipsMgr, config.Toggles.PlaceAzureChainFirst)
	return npMgr
}

func (npMgr *NetworkPolicyManager) MarshalJSON() ([]byte, error) {
	m := map[CacheKey]json.RawMessage{}

	var npmNamespaceCacheRaw []byte
	var err error
	if npMgr.config.Toggles.EnableV2NPM {
		npmNamespaceCacheRaw, err = json.Marshal(npMgr.npmNamespaceCacheV2)
	} else {
		npmNamespaceCacheRaw, err = json.Marshal(npMgr.npmNamespaceCacheV1)
	}

	if err != nil {
		return nil, errors.Errorf("%s: %v", errMarshalNPMCache, err)
	}
	m[NsMap] = npmNamespaceCacheRaw

	var podControllerRaw []byte
	if npMgr.config.Toggles.EnableV2NPM {
		podControllerRaw, err = json.Marshal(npMgr.podControllerV2)
	} else {
		podControllerRaw, err = json.Marshal(npMgr.podControllerV1)
	}

	if err != nil {
		return nil, errors.Errorf("%s: %v", errMarshalNPMCache, err)
	}
	m[PodMap] = podControllerRaw

	// TODO(jungukcho): NPM debug may be broken.
	// Will fix it later after v2 controller and linux test if it is broken.
	if !npMgr.config.Toggles.EnableV2NPM && npMgr.ipsMgr != nil {
		listMapRaw, listMapMarshalErr := npMgr.ipsMgr.MarshalListMapJSON()
		if listMapMarshalErr != nil {
			return nil, errors.Errorf("%s: %v", errMarshalNPMCache, listMapMarshalErr)
		}
		m[ListMap] = listMapRaw

		setMapRaw, setMapMarshalErr := npMgr.ipsMgr.MarshalSetMapJSON()
		if setMapMarshalErr != nil {
			return nil, errors.Errorf("%s: %v", errMarshalNPMCache, setMapMarshalErr)
		}
		m[SetMap] = setMapRaw
	}

	nodeNameRaw, err := json.Marshal(npMgr.NodeName)
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

// GetAppVersion returns network policy manager app version
func (npMgr *NetworkPolicyManager) GetAppVersion() string {
	return npMgr.version
}

// Start starts shared informers and waits for the shared informer cache to sync.
func (npMgr *NetworkPolicyManager) Start(config npmconfig.Config, stopCh <-chan struct{}) error {
	if !config.Toggles.EnableV2NPM {
		// Do initialization of data plane before starting syncup of each controller to avoid heavy call to api-server
		if err := npMgr.netPolControllerV1.ResetDataPlane(); err != nil {
			return fmt.Errorf("Failed to initialized data plane with err %w", err)
		}
	}

	// Starts all informers manufactured by npMgr's informerFactory.
	npMgr.informerFactory.Start(stopCh)

	// Wait for the initial sync of local cache.
	if !cache.WaitForCacheSync(stopCh, npMgr.podInformer.Informer().HasSynced) {
		return fmt.Errorf("Pod informer error: %w", ErrInformerSyncFailure)
	}

	if !cache.WaitForCacheSync(stopCh, npMgr.nsInformer.Informer().HasSynced) {
		return fmt.Errorf("Namespace informer error: %w", ErrInformerSyncFailure)
	}

	if !cache.WaitForCacheSync(stopCh, npMgr.npInformer.Informer().HasSynced) {
		return fmt.Errorf("NetworkPolicy informer error: %w", ErrInformerSyncFailure)
	}

	// start v2 NPM controllers after synced
	if config.Toggles.EnableV2NPM {
		go npMgr.podControllerV2.Run(stopCh)
		go npMgr.namespaceControllerV2.Run(stopCh)
		go npMgr.netPolControllerV2.Run(stopCh)
		return nil
	}

	// start v1 NPM controllers after synced
	go npMgr.podControllerV1.Run(stopCh)
	go npMgr.namespaceControllerV1.Run(stopCh)
	go npMgr.netPolControllerV1.Run(stopCh)
	go npMgr.netPolControllerV1.RunPeriodicTasks(stopCh)

	return nil
}

// GetAIMetadata returns ai metadata number
func GetAIMetadata() string {
	return aiMetadata
}

func GetNodeName() string {
	nodeName := os.Getenv(EnvNodeName)
	return nodeName
}
