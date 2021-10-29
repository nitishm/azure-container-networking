package ipsets

import (
	"testing"

	"github.com/Azure/azure-container-networking/common"
	testutils "github.com/Azure/azure-container-networking/test/utils"
	"github.com/stretchr/testify/require"
)

var (
	hns = GetHNSFake()
)

func TestAddToSetWindows(t *testing.T) {
	iMgr := NewIPSetManager(iMgrApplyAlwaysCfg, common.NewMockIOShimWithFakeHNS([]testutils.TestCmd{}, hns))

	setMetadata := NewIPSetMetadata(testSetName, Namespace)
	iMgr.CreateIPSets([]*IPSetMetadata{setMetadata})

	err := iMgr.AddToSets([]*IPSetMetadata{setMetadata}, testPodIP, testPodKey)
	require.NoError(t, err)

	err = iMgr.AddToSets([]*IPSetMetadata{setMetadata}, "2001:db8:0:0:0:0:2:1", "newpod")
	require.NoError(t, err)

	// same IP changed podkey
	err = iMgr.AddToSets([]*IPSetMetadata{setMetadata}, testPodIP, "newpod")
	require.NoError(t, err)

	listMetadata := NewIPSetMetadata("testipsetlist", KeyLabelOfNamespace)
	iMgr.CreateIPSets([]*IPSetMetadata{listMetadata})
	err = iMgr.AddToSets([]*IPSetMetadata{listMetadata}, testPodIP, testPodKey)
	require.Error(t, err)

	err = iMgr.applyIPSets()
	require.NoError(t, err)
}
