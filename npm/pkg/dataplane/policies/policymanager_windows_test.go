package policies

import "testing"

func TestCompareAndRemovePolicies(t *testing.T) {
	epbuilder := newEndpointPolicyBuilder()

	testPol := &NPMACLPolSettings{
		Id:        "test1",
		Protocols: string(TCP),
	}
	testPol2 := &NPMACLPolSettings{
		Id:        "test1",
		Protocols: string(UDP),
	}

	epbuilder.aclPolicies = append(epbuilder.aclPolicies, []*NPMACLPolSettings{testPol, testPol2}...)

	epbuilder.compareAndRemovePolicies("test1", 2)

	if len(epbuilder.aclPolicies) != 0 {
		t.Errorf("Expected 0 policies, got %d", len(epbuilder.aclPolicies))
	}
}
