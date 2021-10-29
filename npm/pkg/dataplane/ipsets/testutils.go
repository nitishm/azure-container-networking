package ipsets

import (
	"github.com/Azure/azure-container-networking/network/hnswrapper"
	"github.com/Azure/azure-container-networking/npm/util"
	"github.com/Microsoft/hcsshim/hcn"
)

type TestSet struct {
	Metadata   *IPSetMetadata
	PrefixName string
	HashedName string
}

func CreateTestSet(name string, setType SetType) *TestSet {
	set := &TestSet{
		Metadata: &IPSetMetadata{
			Name: name,
			Type: setType,
		},
	}
	set.PrefixName = set.Metadata.GetPrefixName()
	set.HashedName = util.GetHashedName(set.PrefixName)
	return set
}

func GetHNSFake() *hnswrapper.Hnsv2wrapperFake {
	hns := hnswrapper.NewHnsv2wrapperFake()
	network := &hcn.HostComputeNetwork{
		Id:   "1234",
		Name: "azure",
	}

	hns.CreateNetwork(network)

	return hns
}

var (
	TestNSSet           = CreateTestSet("test-ns-set", Namespace)
	TestKeyPodSet       = CreateTestSet("test-keyPod-set", KeyLabelOfPod)
	TestKVPodSet        = CreateTestSet("test-kvPod-set", KeyValueLabelOfPod)
	TestNamedportSet    = CreateTestSet("test-namedport-set", NamedPorts)
	TestCIDRSet         = CreateTestSet("test-cidr-set", CIDRBlocks)
	TestKeyNSList       = CreateTestSet("test-keyNS-list", KeyLabelOfNamespace)
	TestKVNSList        = CreateTestSet("test-kvNS-list", KeyValueLabelOfNamespace)
	TestNestedLabelList = CreateTestSet("test-nestedlabel-list", NestedLabelOfPod)
)
