/*
Copyright 2023.

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
	"context"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/builder" // Required for Watching
	"sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/handler" // Required for Watching
    "sigs.k8s.io/controller-runtime/pkg/reconcile" // Required for Watching
    "sigs.k8s.io/controller-runtime/pkg/source" // Required for Watching
	"sigs.k8s.io/controller-runtime/pkg/log"

	//porchv1alpha1 "github.com/GoogleContainerTools/kpt/porch/api/porch/v1alpha1"
	//nfdeployv1alpha1 "github.com/s3wong/api/nf_deployments/v1alpha1"
	//nfreqv1alpha1 "github.com/s3wong/api/nf_requirements/v1alpha1"
	nftopov1alpha1 "github.com/s3wong/api/nf_topology/v1alpha1"
	//nftopov1alpha1 "github.com/nephio-project/api/nf_topology/v1alpha1"
	porchv1alpha1 "github.com/s3wong/porchapi/api/v1alpha1"
)

type PackageRevisionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
    NFTopologyR *NFTopologyReconciler
}

/*
 * BuildAttachmentMap builds a map from attachment name to a list of NFInstances (template)
 */
func BuildAttachmentMap(nfTopo *nftopov1alpha1.NFTopology) map[string][]nftopov1alpha1.NFInstance {
	ret := make(map[string][]nftopov1alpha1.NFInstance)
	nfTopoSpec := nfTopo.Spec
	for _, nfInst := range nfTopoSpec.NFInstances {
		for _, nfAttachment := range nfInst.NFTemplate.NFInterfaces {
			ret[nfAttachment.NetworkInstanceName] = append(ret[nfAttachment.NetworkInstanceName], nfInst)
		}
	}
	return ret
}

/*
 * BuildDeployedInstanceMap builds a map to map instance name (from NFTopology) to a NF instance
 * on NFDeployed
 */
func BuildDeployedInstanceMap(nfDeployed []nftopov1alpha1.NFDeployedInstance) map[string]int {
	nfDeployedMap := map[string]int{}

	for idx, nfInst := range nfDeployed {
		nfDeployedMap[nfInst.NFInstaceName] = idx
	}
	return nfDeployedMap
}

/*
 * BuildNeighbormap builds a map of NFInstance(s) that are connected to each other
 */
func BuildNeighbormap(attachmentMap map[string][]nftopov1alpha1.NFInstance) map[string][]nftopov1alpha1.NFInstance {
	ret := make(map[string][]nftopov1alpha1.NFInstance)
	for _, nfInstList := range attachmentMap {
		for _, nfInst := range nfInstList {
			for _, neighbor := range nfInstList {
				if nfInst.Name != neighbor.Name {
					ret[nfInst.Name] = append(ret[nfInst.Name], neighbor)
				}
			}
		}
	}
	return ret
}

func BuildNFDeployed(nfDeployedName string, nfTopo *nftopov1alpha1.NFTopology, nfInstance *nftopov1alpha1.NFInstance, nfdeployed []nftopov1alpha1.NFDeployedInstance, neighborMap map[string][]nftopov1alpha1.NFInstance, instMap map[string]int) error {
	nfdeployedIdx := -1

	for idx, nfInst := range nfdeployed.NFInstances {
		if nfInst.ID == nfDeployedName {
			nfdeployedIdx = idx
			break
		}
	}

	if nfdeployedIdx == -1 {
		nfdeployedInst := nftopov1alpha1.NFDeployedInstance{}
		nfTemplate := nfInstance.NFTemplate
		nfdeployedInst.Id = nfDeployedName
		//nfdeployedInst.ClusterName = clusterName
		nfdeployedInst.NFType = string(nfTemplate.NFType)
        /*
		nfdeployedInst.NFVendor = vendor
		nfdeployedInst.NFVersion = version
        */
        nfdeployedInst.NFInstaceName = nfInstance.ObjectMeta.Name
		nfdeployed = append(nfdeployed, nfdeployedInst)
		nfdeployedIdx = len(nfdeployed) - 1
	}

	// this is the continuous update case; so first build the neighbor list for this instance, then
	// for all the connected instance(s), update by appending the neighbor list
	/*
	 * TODO(s3wong): this now assumes each NFInstance from NFTopology.NFInstance will generate exactly
	 * one NF instance, an assumption that should hold true for R1
	 */
	neighborSlice, _ := neighborMap[nfInstance.ObjectMeta.Name]
	for _, neighbor := range neighborSlice {
		if neighborIdx, ok := instMap[neighbor.Name]; !ok {
			// neighbor packagerevision object not created yet
			continue
		} else {
			neighborInst := &nfdeployed[neighborIdx]
			nfdeployedInstance := &nfdeployed[nfdeployedIdx]
			con := nftopov1alpha1.NFConnectivity{}
			con.NeighborName = neighborInst.Id
			nfdeployedInstance.Connectivities = append(nfdeployedInstance.Connectivities, con)
			neighborCon := nftopov1alpha1.NFConnectivity{}
			neighborCon.NeighborName = nfdeployedInstance.Id
			neighborInst.Connectivities = append(neighborInst.Connectivities, neighborCon)
		}
	}
	return nil
}

// remove an nf from the NFInstance list
// TODO(): highly inefficient algorithm, it runs the list once to look for the index for the NF, and
// the remove utilizes an O(N) algorithm to shift element to maintain order
func RemoveNFfromList(nfName string, nfInstList []nftopov1alpha1.NFInstance) {
    var nfIdx := -1
    for idx, nf := range nfInstList {
        if nfName == nf.Name {
            nfIdx = idx
            break
        }
    }
    if nfIdx == -1 {
        return
    }
    copy(nfInstList[nfIdx:], nfInstList[nfIdx+1:])
    nfInstList[len(nfInstList) - 1] = nftopov1alpha1.NFInstance{}
    nfInstList = nfInstList[:len(nfInstList)-1]
}

// 
func (r *PackageRevisionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("NFTopology-PackageRevision", req.NamespacedName)

    /*
	if pkgRev.Spec.Lifecycle != porchv1alpha1.PackageRevisionLifecyclePublished {
		l.Info("Package lifecycle is not Published, ignore", "PackageRevision.lifecycle",
			pkgRev.Spec.Lifecycle)
		return ctrl.Result{}, nil
	}
    */

	// repo name is the cluster name
	clusterName = pkgRev.Spec.RepositoryName

	// TODO(s3wong): potentially, there would be multiple NFDeployedInstance CRs for an instance of
	// NFInstance --- which means deploymentName should be unique
	deploymentName := ConstructNFDeployedName(nfInstanceName, regionName, clusterName)

	// Search for corresponding NFInstance
	var nfInstance *nftopov1alpha1.NFInstance = nil
	for _, nfInst := range nfTopology.Spec.NFInstances {
		if nfInst.Name == nfInstanceName {
			nfInstance = &nfInst
			break
		}
	}
	if nfInstance == nil {
		err := errors.New(fmt.Sprintf("%s not found in NF Topology %s\n", nfInstanceName, topologyName))
		l.Error(err, fmt.Sprintf("NF Instance %s not found in NF Topology %s\n", nfInstanceName, topologyName))
		return ctrl.Result{}, err
	}

    nfTopoStatus := &nfTopology.Status
    nfDeployed := nfTopoStatus.NFInstances

	if err := BuildNFDeployed(deploymentName, topoNamespace, nfTopology, nfInstance, clusterName, nfClass.Spec.Vendor, nfClass.Spec.Version, nfDeployed); err != nil {
		l.Error(err, fmt.Sprintf("Failed to build NFDeployed %s: %s\n", deploymentName, err.Error()))
		return ctrl.Result{}, err
	}

    if err := r.syncStatus(ctx, nfTopology, nfTopoStatus); err != nil {
        log.Error(err, "Failed to update status")
        return reconcile.Result{}, err
    }

	return ctrl.Result{}, nil
}

// syncStatus creates or updates NFTopology status subresource
func (r *PackageRevisionReconciler) syncStatus(ctx context.Context, nftopology *nftopov1alpha1.NFTopology, nftopoStatus *nftopov1alpha1.NFTopologyStatus) error {
    nftopology = nftopology.DeepCopy()
    nftopology.Status = nftopoStatus
    if err := r.Client.Update(ctx, nftopology); err != nil {
        log.Error(err, "Failed to update NFTopology status", "NFTopology.namespace", nftopology.Namespace, "NFTopology.Name", nftopology.Name)
        return err
    }
    return nil
}
