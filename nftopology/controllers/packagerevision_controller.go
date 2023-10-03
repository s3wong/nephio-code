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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	//porchv1alpha1 "github.com/GoogleContainerTools/kpt/porch/api/porch/v1alpha1"
	//nfdeployv1alpha1 "github.com/s3wong/api/nf_deployments/v1alpha1"
	//nfreqv1alpha1 "github.com/s3wong/api/nf_requirements/v1alpha1"
	nfdeployv1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"
	nfreqv1alpha1 "github.com/nephio-project/api/nf_requirements/v1alpha1"
	porchv1alpha1 "github.com/s3wong/porchapi/api/v1alpha1"
)

type PackageRevisionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	delimiter = "--"
)

func ConstructNFDeployedName(nfInstanceName string, regionName string, clusterName string) string {
	return nfInstanceName + delimiter + regionName + delimiter + clusterName
}

func GetNFInstanceNameFromNFDeployedName(nfDeployedName string) string {
	return strings.Split(nfDeployedName, delimiter)[0]
}

/*
 * BuildAttachmentMap builds a map from attachment name to a list of NFInstances (template)
 */
func BuildAttachmentMap(nfTopo *nfreqv1alpha1.NFTopology) map[string][]nfreqv1alpha1.NFInstance {
	ret := make(map[string][]nfreqv1alpha1.NFInstance)
	nfTopoSpec := nfTopo.Spec
	for _, nfInst := range nfTopoSpec.NFInstances {
		for _, nfAttachment := range nfInst.NFTemplate.NFAttachments {
			ret[nfAttachment.NetworkInstanceName] = append(ret[nfAttachment.NetworkInstanceName], nfInst)
		}
	}
	return ret
}

/*
 * BuildDeployedInstanceMap builds a map to map instance name (from NFTopology) to a NF instance
 * on NFDeployed
 */
func BuildDeployedInstanceMap(nfDeployed *nfdeployv1alpha1.NFDeployed) map[string]int {
	nfDeployedMap := map[string]int{}

	for idx, nfInst := range nfDeployed.Spec.NFInstances {
		instName := GetNFInstanceNameFromNFDeployedName(nfInst.Id)
		nfDeployedMap[instName] = idx
	}
	return nfDeployedMap
}

/*
 * BuildNeighbormap builds a map of NFInstance(s) that are connected to each other
 */
func BuildNeighbormap(attachmentMap map[string][]nfreqv1alpha1.NFInstance) map[string][]nfreqv1alpha1.NFInstance {
	ret := make(map[string][]nfreqv1alpha1.NFInstance)
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

func BuildNFDeployed(nfDeployedName string, topoNamespace string, nfTopo *nfreqv1alpha1.NFTopology, nfInstance *nfreqv1alpha1.NFInstance, clusterName string, vendor string, version string, nfdeployed *nfdeployv1alpha1.NFDeployed) error {
	//var nfdeployedInstance *nfdeployv1alpha1.NFDeployedInstance = nil
	nfdeployedIdx := -1

	for idx, nfInst := range nfdeployed.Spec.NFInstances {
		if nfInst.Id == nfDeployedName {
			nfdeployedIdx = idx
			break
		}
	}

	if nfdeployedIdx == -1 {
		nfdeployedInst := nfdeployv1alpha1.NFDeployedInstance{}
		nfTemplate := nfInstance.NFTemplate
		nfdeployedInst.Id = nfDeployedName
		nfdeployedInst.ClusterName = clusterName
		nfdeployedInst.NFType = string(nfTemplate.NFType)
		nfdeployedInst.NFVendor = vendor
		nfdeployedInst.NFVersion = version
		nfdeployed.Spec.NFInstances = append(nfdeployed.Spec.NFInstances, nfdeployedInst)
		nfdeployedIdx = len(nfdeployed.Spec.NFInstances) - 1
	}

	neighborMap := BuildNeighbormap(BuildAttachmentMap(nfTopo))
	instMap := BuildDeployedInstanceMap(nfdeployed)

	// this is the continuous update case; so first build the neighbor list for this instance, then
	// for all the connected instance(s), update by appending the neighbor list
	/*
	 * TODO(s3wong): this now assumes each NFInstance from NFTopology.NFInstance will generate exactly
	 * one NF instance, an assumption that should hold true for R1
	 */
	instName := GetNFInstanceNameFromNFDeployedName(nfDeployedName)
	neighborSlice, _ := neighborMap[instName]
	for _, neighbor := range neighborSlice {
		if neighborIdx, ok := instMap[neighbor.Name]; !ok {
			// neighbor packagerevision object not created yet
			continue
		} else {
			neighborInst := &(nfdeployed.Spec.NFInstances[neighborIdx])
			nfdeployedInstance := &(nfdeployed.Spec.NFInstances[nfdeployedIdx])
			con := nfdeployv1alpha1.NFDeployedConnectivity{}
			con.NeighborName = neighborInst.Id
			nfdeployedInstance.Connectivities = append(nfdeployedInstance.Connectivities, con)
			neighborCon := nfdeployv1alpha1.NFDeployedConnectivity{}
			neighborCon.NeighborName = nfdeployedInstance.Id
			neighborInst.Connectivities = append(neighborInst.Connectivities, neighborCon)
		}
	}
	return nil
}

// +kubebuilder:rbac:groups=porch.kpt.dev,resources=packagerevisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=porch.kpt.dev,resources=packagerevisions/status,verbs=get;update;patch
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PackageRevisionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("NFTopology-PackageRevision", req.NamespacedName)

	pkgRev := &porchv1alpha1.PackageRevision{}
	if err := r.Get(ctx, req.NamespacedName, pkgRev); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Step 1: check package revision state
	// TODO(s3wong): if package deleted, inform ESA
	if pkgRev.Spec.Lifecycle != porchv1alpha1.PackageRevisionLifecyclePublished {
		l.Info("Package lifecycle is not Published, ignore", "PackageRevision.lifecycle",
			pkgRev.Spec.Lifecycle)
		return ctrl.Result{}, nil
	}

	var regionName, clusterName, nfInstanceName, topologyName, topoNamespace string
	var createDeployed, found bool
	pkgRevLabels := pkgRev.ObjectMeta.Labels
	if regionName, found = pkgRevLabels["region"]; !found {
		l.Info("region name not found from package")
		return ctrl.Result{}, nil
	}
	if nfInstanceName, found = pkgRevLabels["nfinstance"]; !found {
		l.Info("NF Instance name not found from package")
		return ctrl.Result{}, nil
	}
	if topologyName, found = pkgRevLabels["nftopology"]; !found {
		l.Info("NF Topology name not found from package")
		return ctrl.Result{}, nil
	}
	if topoNamespace, found = pkgRevLabels["nftopology-namespace"]; !found {
		l.Info("NF Topology namespace not found from package")
		return ctrl.Result{}, nil
	}
	// repo name is the cluster name
	clusterName = pkgRev.Spec.RepositoryName

	// Step 2: search for NFTopology and NFDeployed
	nfTopology := &nfreqv1alpha1.NFTopology{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: topoNamespace, Name: topologyName}, nfTopology); err != nil {
		l.Error(err, fmt.Sprintf("NFTopology object not found: %s\n", topologyName))
		return ctrl.Result{}, nil
	}

	// TODO(s3wong): potentially, there would be multiple NFDeployedInstance CRs for an instance of
	// NFInstance --- which means deploymentName should be unique
	deploymentName := ConstructNFDeployedName(nfInstanceName, regionName, clusterName)

	nfDeployed := &nfdeployv1alpha1.NFDeployed{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: topoNamespace, Name: nfTopology.ObjectMeta.Name}, nfDeployed); err != nil {
		createDeployed = true
		nfDeployed.ObjectMeta.Name = nfTopology.ObjectMeta.Name
		nfDeployed.ObjectMeta.Namespace = nfTopology.ObjectMeta.Namespace
	}

	// Search for corresponding NFInstance
	var nfInstance *nfreqv1alpha1.NFInstance = nil
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
	// Get NFClass, cluster scope
	className := nfInstance.NFTemplate.ClassName
	var nfClass = &nfreqv1alpha1.NFClass{}
	if err := r.Client.Get(ctx, client.ObjectKey{Name: className}, nfClass); err != nil {
		l.Error(err, fmt.Sprintf("NFClass object not found: %s: %s\n", className, err.Error()))
		return ctrl.Result{}, err
	}
	if err := BuildNFDeployed(deploymentName, topoNamespace, nfTopology, nfInstance, clusterName, nfClass.Spec.Vendor, nfClass.Spec.Version, nfDeployed); err != nil {
		l.Error(err, fmt.Sprintf("Failed to build NFDeployed %s: %s\n", deploymentName, err.Error()))
		return ctrl.Result{}, err
	}

	if createDeployed {
		if err := r.Client.Create(ctx, nfDeployed); err != nil {
			l.Error(err, fmt.Sprintf("Failed to create NFDeployed %s: %s\n", deploymentName, err.Error()))
			return ctrl.Result{}, err
		}
	} else {
		if err := r.Client.Update(ctx, nfDeployed); err != nil {
			l.Error(err, fmt.Sprintf("Failed to update NFDeployed %s: %s\n", deploymentName, err.Error()))
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PackageRevisionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&porchv1alpha1.PackageRevision{}).
		Owns(&nfdeployv1alpha1.NFDeployed{}).
		Complete(r)
}
