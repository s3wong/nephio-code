/*
Copyright 2023 The Nephio Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nfreqv1alpha1 "github.com/nephio-project/api/nf_requirements/v1alpha1"
	nftopov1alpha1 "github.com/nephio-project/api/nf_topology/v1alpha1"
	//nftopov1alpha1 "github.com/s3wong/api/nf_topology/v1alpha1"
)

const (
	LABEL_REGION             string = "us-west-1"
	LABEL_AGGREGATION        string = "aggregation"
	LABEL_EDGE               string = "edge"
	LABEL_NEW_EDGE           string = "edge-new"
	NETWORK_US_WEST          string = "us-west-network-1"
	NETWORK_US_WEST_RAN_ONE  string = "us-west-ran-core-network-1"
	NETWORK_US_WEST_DATA_ONE string = "us-west-data-network-1"
	NETWORK_US_WEST_RAN_TWO  string = "us-west-ran-core-network-2"
	NETWORK_US_WEST_DATA_TWO string = "us-west-data-network-2"
)

func getSMF1() nftopov1alpha1.NFInstance {
	return nftopov1alpha1.NFInstance{
		Name: "SMF-1",
		ClusterSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"region":  LABEL_REGION,
				"cluster": LABEL_AGGREGATION,
			},
		},
		NFTemplate: nftopov1alpha1.NFTemplate{
			NFType: "SMF",
			NFPackageRef: nftopov1alpha1.PackageRevisionReference{
				RepositoryName: "github.com/sample-repo",
				PackageName:    "mysmfpackage",
				Revision:       "version-1",
			},
			Capacity: nfreqv1alpha1.CapacitySpec{
				MaxSessions:      1000000,
				MaxNFConnections: 1000,
			},
			NFInterfaces: []nftopov1alpha1.NFInterface{
				{
					Name:                "n4",
					NetworkInstanceName: NETWORK_US_WEST,
				},
			},
		},
	}
}

func getUPF1() nftopov1alpha1.NFInstance {
	return nftopov1alpha1.NFInstance{
		Name: "UPF-1",
		ClusterSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"region":  LABEL_REGION,
				"cluster": LABEL_EDGE,
			},
		},
		NFTemplate: nftopov1alpha1.NFTemplate{
			NFType: "UPF",
			NFPackageRef: nftopov1alpha1.PackageRevisionReference{
				RepositoryName: "github.com/sample-repo",
				PackageName:    "myupfpackage",
				Revision:       "version-1",
			},
			Capacity: nfreqv1alpha1.CapacitySpec{
				MaxUplinkThroughput:   resource.MustParse("1G"),
				MaxDownlinkThroughput: resource.MustParse("10G"),
			},
			NFInterfaces: []nftopov1alpha1.NFInterface{
				{
					Name:                "n3",
					NetworkInstanceName: NETWORK_US_WEST_RAN_ONE,
				},
				{
					Name:                "n4",
					NetworkInstanceName: NETWORK_US_WEST,
				},
				{
					Name:                "n6",
					NetworkInstanceName: NETWORK_US_WEST_DATA_ONE,
				},
			},
		},
	}
}

func getUPF2() nftopov1alpha1.NFInstance {
	return nftopov1alpha1.NFInstance{
		Name: "UPF-2",
		ClusterSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"region":  LABEL_REGION,
				"cluster": LABEL_NEW_EDGE,
			},
		},
		NFTemplate: nftopov1alpha1.NFTemplate{
			NFType: "UPF",
			NFPackageRef: nftopov1alpha1.PackageRevisionReference{
				RepositoryName: "github.com/sample-repo",
				PackageName:    "myupfpackage",
				Revision:       "version-1",
			},
			Capacity: nfreqv1alpha1.CapacitySpec{
				MaxUplinkThroughput:   resource.MustParse("1G"),
				MaxDownlinkThroughput: resource.MustParse("5G"),
			},
			NFInterfaces: []nftopov1alpha1.NFInterface{
				{
					Name:                "n3",
					NetworkInstanceName: NETWORK_US_WEST_RAN_TWO,
				},
				{
					Name:                "n4",
					NetworkInstanceName: NETWORK_US_WEST,
				},
				{
					Name:                "n6",
					NetworkInstanceName: NETWORK_US_WEST_DATA_TWO,
				},
			},
		},
	}
}

func getSMF1NFDeployedInstance() nftopov1alpha1.NFDeployedInstance {
	return nftopov1alpha1.NFDeployedInstance{
		ID:             "SMF-1",
		NFType:         "smf",
		NFInstanceName: "SMF-1",
		Connectivities: []nftopov1alpha1.NFConnectivity{
			{
				NeighborName: "UPF-1",
			},
			{
				NeighborName: "UPF-2",
			},
		},
	}
}

func getUPF1NFDeployedInstance() nftopov1alpha1.NFDeployedInstance {
	return nftopov1alpha1.NFDeployedInstance{
		ID:             "UPF-1",
		NFType:         "upf",
		NFInstanceName: "UPF-1",
		Connectivities: []nftopov1alpha1.NFConnectivity{
			{
				NeighborName: "SMF-1",
			},
		},
	}
}

func getUPF2NFDeployedInstance() nftopov1alpha1.NFDeployedInstance {
	return nftopov1alpha1.NFDeployedInstance{
		ID:             "UPF-2",
		NFType:         "upf",
		NFInstanceName: "UPF-2",
		Connectivities: []nftopov1alpha1.NFConnectivity{
			{
				NeighborName: "SMF-1",
			},
		},
	}
}

func newNFTopology(name string, namespace string) *nftopov1alpha1.NFTopology {
	nfTopo := &nftopov1alpha1.NFTopology{}
	nfTopo.ObjectMeta.Name = name
	nfTopo.ObjectMeta.Namespace = namespace

	nfInstanceSmf := getSMF1()
	nfInstanceUpf1 := getUPF1()
	nfInstanceUpf2 := getUPF2()

	nfTopo.Spec.NFInstances = append(nfTopo.Spec.NFInstances, nfInstanceSmf)
	nfTopo.Spec.NFInstances = append(nfTopo.Spec.NFInstances, nfInstanceUpf1)
	nfTopo.Spec.NFInstances = append(nfTopo.Spec.NFInstances, nfInstanceUpf2)

	return nfTopo
}

func TestBuildDeployedInstanceMap(t *testing.T) {
	nftopo := newNFTopology("test-nf-topology", "test")

	// the packagerevision helps build NFTopology's Status NFInstances array
	// here we simulate it
	nftopo.Status.NFInstances = append(nftopo.Status.NFInstances, getSMF1NFDeployedInstance())
	nftopo.Status.NFInstances = append(nftopo.Status.NFInstances, getUPF1NFDeployedInstance())
	nftopo.Status.NFInstances = append(nftopo.Status.NFInstances, getUPF2NFDeployedInstance())

	nfDeployMap := BuildDeployedInstanceMap(nftopo.Status.NFInstances)

	expected := map[string]int{
		"SMF-1": 0,
		"UPF-1": 1,
		"UPF-2": 2,
	}

	for nf, idx := range expected {
		if val, found := nfDeployMap[nf]; !found {
			t.Errorf("BuildDeployedInstanceMap: expected key %s not in resulting map. Expected %+v, map %+v", nf, expected, nfDeployMap)
		} else {
			if val != idx {
				t.Errorf("BuildDeployedInstanceMap: value for key %s not equal %d != %d. Expected %+v, map %+v", nf, idx, val, expected, nfDeployMap)
			}
		}
	}
}

func TestBuildNeighbormap(t *testing.T) {
	nftopo := newNFTopology("test-nf-topology", "test")

	nfAttachmentMap := BuildAttachmentMap(nftopo)

	nfNeighborMap := BuildNeighbormap(nfAttachmentMap)

	smfInst := getSMF1()
	upf1Inst := getUPF1()
	upf2Inst := getUPF2()

	expected := map[string][]nftopov1alpha1.NFInstance{
		"SMF-1": {upf1Inst, upf2Inst},
		"UPF-1": {smfInst},
		"UPF-2": {smfInst},
	}

	for nf, nfInstSlice := range expected {
		if valSlice, ok := nfNeighborMap[nf]; !ok {
			t.Errorf("BuildNeighbormap has missing elements %s; expected: %+v, neighborMap: %+v", nf, expected, nfNeighborMap)
		} else {
			for _, nfInst := range nfInstSlice {
				found := false
				for _, val := range valSlice {
					if nfInst.Name == val.Name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("BuildNeighbormap has missing elements; \nexpected: %+v, \nneighborMap: %+v", expected, nfNeighborMap)
				}
			}
		}
	}
}

func TestBuildAttachmentMap(t *testing.T) {
	nftopo := newNFTopology("test-nf-topology", "test")

	attachmentMap := BuildAttachmentMap(nftopo)

	smfInst := getSMF1()
	upf1Inst := getUPF1()
	upf2Inst := getUPF2()

	expected := map[string][]nftopov1alpha1.NFInstance{
		NETWORK_US_WEST:          {smfInst, upf1Inst, upf2Inst},
		NETWORK_US_WEST_RAN_ONE:  {upf1Inst},
		NETWORK_US_WEST_DATA_ONE: {upf1Inst},
		NETWORK_US_WEST_RAN_TWO:  {upf2Inst},
		NETWORK_US_WEST_DATA_TWO: {upf2Inst},
	}

	for key, val := range attachmentMap {
		if cmpSlice, ok := expected[key]; !ok {
			t.Errorf("BuildAttachmentMap(%+v) returned %+v, expected: %+v, key %s not found", nftopo.Spec, attachmentMap, expected, key)
		} else {
			if !reflect.DeepEqual(cmpSlice, val) {
				t.Errorf("BuildAttachmentMap(%+v)\n Returned %+v\n Expected: %+v", nftopo.Spec, attachmentMap, expected)
			}
		}
	}
}

func TestBuildNFDeployed(t *testing.T) {
	nfTopo := newNFTopology("test-nf-topology", "test")
	nfInstMap := map[string]*nftopov1alpha1.NFInstance{}
	for _, ele := range nfTopo.Spec.NFInstances {
		nfInstMap[ele.Name] = &ele
	}

	neighborMap := BuildNeighbormap(BuildAttachmentMap(nfTopo))
	deployedInstanceMap := BuildDeployedInstanceMap(nfTopo.Status.NFInstances)

	expected_connectivities := map[string][]nftopov1alpha1.NFConnectivity{
		"UPF-1": {
			{
				NeighborName: "SMF-1",
			},
			{
				NeighborName: "UPF-2",
			},
		},
		"SMF-1": {
			{
				NeighborName: "UPF-1",
			},
			{
				NeighborName: "UPF-2",
			},
		},
		"UPF-2": {
			{
				NeighborName: "SMF-1",
			},
			{
				NeighborName: "UPF-1",
			},
		},
	}

	expected_nfdeployedinstance := map[string]nftopov1alpha1.NFDeployedInstance{
		"UPF-1": {
			ID:             "UPF-1",
			NFType:         "upf",
			NFInstanceName: "UPF-1",
			Connectivities: expected_connectivities["UPF-1"],
		},
		"SMF-1": {
			ID:             "SMF-1",
			NFType:         "smf",
			NFInstanceName: "SMF-1",
			Connectivities: expected_connectivities["SMF-1"],
		},
		"UPF-2": {
			ID:             "UPF-2",
			NFType:         "upf",
			NFInstanceName: "UPF-2",
			Connectivities: expected_connectivities["UPF-2"],
		},
	}

	prList := []string{
		"SMF-1",
		"UPF-1",
		"UPF-2",
	}

	for _, pr := range prList {
		/*
		 * the NFInstance mapping to PackageRevision name is taking advantage of the two
		 * sharing the same name and only 1:1 within the context of this test
		 */
		if err := BuildNFDeployed(pr, nfTopo, nfInstMap[pr], nfTopo.Status.NFInstances, neighborMap, deployedInstanceMap); err != nil {
			t.Errorf("BuildNFDeployed failed with topology %+v\n, for nfInst %s: %+v", nfTopo, pr, err.Error())
		}
	}

	for _, pr := range prList {
		found := false
		for _, nfDeployedInst := range nfTopo.Status.NFInstances {
			if nfDeployedInst.ID == pr {
				found = true
				if !reflect.DeepEqual(nfDeployedInst, expected_nfdeployedinstance[pr]) {
					t.Errorf("BuildNFDeployed returned %+v, expected: %+v", nfDeployedInst, expected_nfdeployedinstance[pr])
				}
			}
		}
		if !found {
			t.Errorf("BuildNFDeployed: %s not found (%+v)", pr, nfTopo.Status.NFInstances)
		}
	}
}
