// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"os"

	controllersv1 "github.com/Azure/azure-container-networking/npm/pkg/controlplane/controllers/v1"
	controllersv2 "github.com/Azure/azure-container-networking/npm/pkg/controlplane/controllers/v2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	networkinginformers "k8s.io/client-go/informers/networking/v1"
)

var (
	aiMetadata         string
	errMarshalNPMCache = errors.New("failed to marshal NPM Cache")
)

const (
	heartbeatIntervalInMinutes = 30 //nolint:unused,deadcode,varcheck // ignore this error
	// TODO: consider increasing thread number later when logics are correct
	// threadness = 1
)

type CacheKey string

// NPMCache Key Contract for Json marshal and unmarshal
const (
	NodeName    CacheKey = "NodeName"
	NsMap       CacheKey = "NsMap"
	PodMap      CacheKey = "PodMap"
	ListMap     CacheKey = "ListMap"
	SetMap      CacheKey = "SetMap"
	EnvNodeName          = "HOSTNAME"
)

// K8SControllerV1 are the legacy k8s controllers
type K8SControllersV1 struct {
	podControllerV1       *controllersv1.PodController           //nolint:structcheck //ignore this error
	namespaceControllerV1 *controllersv1.NamespaceController     //nolint:structcheck // false lint error
	npmNamespaceCacheV1   *controllersv1.NpmNamespaceCache       //nolint:structcheck // false lint error
	netPolControllerV1    *controllersv1.NetworkPolicyController //nolint:structcheck // false lint error
}

// K8SControllerV2 are the optimized k8s controllers that replace the legacy controllers
type K8SControllersV2 struct {
	podControllerV2       *controllersv2.PodController           //nolint:structcheck //ignore this error
	namespaceControllerV2 *controllersv2.NamespaceController     //nolint:structcheck // false lint error
	npmNamespaceCacheV2   *controllersv2.NpmNamespaceCache       //nolint:structcheck // false lint error
	netPolControllerV2    *controllersv2.NetworkPolicyController //nolint:structcheck // false lint error
}

// Informers are the informers for the k8s controllers
type Informers struct {
	informerFactory informers.SharedInformerFactory           //nolint:structcheck //ignore this error
	podInformer     coreinformers.PodInformer                 //nolint:structcheck // false lint error
	nsInformer      coreinformers.NamespaceInformer           //nolint:structcheck // false lint error
	npInformer      networkinginformers.NetworkPolicyInformer //nolint:structcheck // false lint error
}

// AzureConfig captures the Azure specific configurations and fields
type AzureConfig struct {
	k8sServerVersion *version.Info
	NodeName         string
	version          string
	TelemetryEnabled bool
}

// GetAIMetadata returns ai metadata number
func GetAIMetadata() string {
	return aiMetadata
}

func GetNodeName() string {
	nodeName := os.Getenv(EnvNodeName)
	return nodeName
}
