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
	"sort"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//porchv1alpha1 "github.com/GoogleContainerTools/kpt/porch/api/porch/v1alpha1"
	//nfdeployv1alpha1 "github.com/s3wong/api/nf_deployments/v1alpha1"
	//nfreqv1alpha1 "github.com/s3wong/api/nf_requirements/v1alpha1"
	nfdeployv1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"
	nfreqv1alpha1 "github.com/nephio-project/api/nf_requirements/v1alpha1"
	//porchv1alpha1 "github.com/s3wong/porchapi/api/v1alpha1"
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

func newNFTopology(name string, namespace string) *nfreqv1alpha1.NFTopology {
	nfTopo := &nfreqv1alpha1.NFTopology{}
	nfTopo.ObjectMeta.Name = name
	nfTopo.ObjectMeta.Namespace = namespace

	nfInstanceSmf := nfreqv1alpha1.NFInstance{
		Name: "SMF-1",
		ClusterSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"region":  LABEL_REGION,
				"cluster": LABEL_AGGREGATION,
			},
		},
		NFTemplate: nfreqv1alpha1.NFTemplate{
			NFType:    nfreqv1alpha1.NFTypeSMF,
			ClassName: "smf-class-1",
			Capacity: nfreqv1alpha1.CapacitySpec{
				MaxSessions:      1000000,
				MaxNFConnections: 1000,
			},
			NFAttachments: []nfreqv1alpha1.NFAttachment{
				{
					Name:                "n4",
					NetworkInstanceName: NETWORK_US_WEST,
				},
			},
		},
	}

	nfInstanceUpf1 := nfreqv1alpha1.NFInstance{
		Name: "UPF-1",
		ClusterSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"region":  LABEL_REGION,
				"cluster": LABEL_EDGE,
			},
		},
		NFTemplate: nfreqv1alpha1.NFTemplate{
			NFType:    nfreqv1alpha1.NFTypeUPF,
			ClassName: "upf-class-1",
			Capacity: nfreqv1alpha1.CapacitySpec{
				MaxUplinkThroughput:   resource.MustParse("1G"),
				MaxDownlinkThroughput: resource.MustParse("10G"),
			},
			NFAttachments: []nfreqv1alpha1.NFAttachment{
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

	nfInstanceUpf2 := nfreqv1alpha1.NFInstance{
		Name: "UPF-2",
		ClusterSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"region":  LABEL_REGION,
				"cluster": LABEL_NEW_EDGE,
			},
		},
		NFTemplate: nfreqv1alpha1.NFTemplate{
			NFType:    nfreqv1alpha1.NFTypeUPF,
			ClassName: "upf-class-1",
			Capacity: nfreqv1alpha1.CapacitySpec{
				MaxUplinkThroughput:   resource.MustParse("1G"),
				MaxDownlinkThroughput: resource.MustParse("5G"),
			},
			NFAttachments: []nfreqv1alpha1.NFAttachment{
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

	nfTopo.Spec.NFInstances = append(nfTopo.Spec.NFInstances, nfInstanceSmf)
	nfTopo.Spec.NFInstances = append(nfTopo.Spec.NFInstances, nfInstanceUpf1)
	nfTopo.Spec.NFInstances = append(nfTopo.Spec.NFInstances, nfInstanceUpf2)

	return nfTopo
}

func TestBuildAttachmentMap(t *testing.T) {
	nftopo := newNFTopology("test-nf-topology", "test")

	attachmentMap := BuildAttachmentMap(nftopo)

	expected := map[string][]string{
		NETWORK_US_WEST:          {"SMF-1", "UPF-1", "UPF-2"},
		NETWORK_US_WEST_RAN_ONE:  {"UPF-1"},
		NETWORK_US_WEST_DATA_ONE: {"UPF-1"},
		NETWORK_US_WEST_RAN_TWO:  {"UPF-2"},
		NETWORK_US_WEST_DATA_TWO: {"UPF-2"},
	}

	for key, val := range attachmentMap {
		if cmpSlice, ok := expected[key]; !ok {
			t.Errorf("BuildAttachmentMap(%+v) returned %+v, expected: %+v, key %s not found", nftopo.Spec, attachmentMap, expected, key)
		} else {
			sort.Strings(cmpSlice)
			//sort.Strings(val)
			if !reflect.DeepEqual(cmpSlice, val) {
				t.Errorf("BuildAttachmentMap(%+v) returned %+v, expected: %+v", nftopo.Spec, attachmentMap, expected)
			}
		}
	}
}

func TestBuildNFDeployed(t *testing.T) {
	var nfInst *nfreqv1alpha1.NFInstance = nil
	nfDeployed := &nfdeployv1alpha1.NFDeployed{}
	clusterName := "test-cluster"
	nfTopo := newNFTopology("test-nf-topology", "test")

	nfDeployed.ObjectMeta.Name = "test-nf-topology"
	nfDeployed.ObjectMeta.Namespace = "test"
	for _, ele := range nfTopo.Spec.NFInstances {
		if ele.Name == "UPF-1" {
			nfInst = &ele
			break
		}
	}

	if nfInst == nil {
		t.Errorf("nfTopo(%+v) does not contain NFInstance UPF-1", nfTopo)
	}

	expected_connectivities := []nfdeployv1alpha1.NFDeployedConnectivity{
		{
			NeighborName: "SMF-1",
		},
		{
			NeighborName: "UPF-2",
		},
	}

	nfDeploymentName := ConstructNFDeployedName("UPF-1", LABEL_REGION, clusterName)
	expected := &nfdeployv1alpha1.NFDeployed{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-nf-topology",
			Namespace: "test",
		},
		Spec: nfdeployv1alpha1.NFDeployedSpec{
			NFInstances: []nfdeployv1alpha1.NFDeployedInstance{
				{
					Id:             nfDeploymentName,
					ClusterName:    clusterName,
					NFType:         "upf",
					NFVendor:       "vendor",
					NFVersion:      "v1.0",
					Connectivities: expected_connectivities,
				},
			},
		},
	}

	if err := BuildNFDeployed(nfDeploymentName, "test", nfTopo, nfInst, clusterName, "vendor", "v1.0", nfDeployed); err != nil {
		t.Errorf("BuildNFDeployed failed with topology %+v, for nfInst UPF-1: %+v", nfTopo.Spec, err.Error())
	} else {
		if !reflect.DeepEqual(nfDeployed, expected) {
			t.Errorf("BuildNFDeployed returned %+v, expected: %+v", nfDeployed, expected)
		}
	}
}
