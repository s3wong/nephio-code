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
	"fmt"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//"sigs.k8s.io/controller-runtime/pkg/handler" // Required for Watching
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile" // Required for Watching
	//"sigs.k8s.io/controller-runtime/pkg/source"    // Required for Watching

	"github.com/go-logr/logr"
	nftopov1alpha1 "github.com/nephio-project/api/nf_topology/v1alpha1"
	porchv1alpha1 "github.com/nephio-project/porch/api/porch/v1alpha1"
	pvv1alpha1 "github.com/nephio-project/porch/controllers/packagevariants/api/v1alpha1"
	pvsv1alpha2 "github.com/nephio-project/porch/controllers/packagevariantsets/api/v1alpha2"
	//nftopov1alpha1 "github.com/s3wong/api/nf_topology/v1alpha1"
)

const (
	WORKLOADAPIVERSION string = "infra.nephio.org/v1alpha1"
	WORKLOADKIND       string = "WorkloadCluster"
)

// NFTopologyReconciler reconciles a NFTopology object
type NFTopologyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	l      logr.Logger
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
		nfDeployedMap[nfInst.NFInstanceName] = idx
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

/*
 * BuildNFDeployed --- builds connectivity info as part of the NFTopology resource's Status subresource
 */
func BuildNFDeployed(nfDeployedName string, nfTopo *nftopov1alpha1.NFTopology, nfInstance *nftopov1alpha1.NFInstance, nfdeployed []nftopov1alpha1.NFDeployedInstance, neighborMap map[string][]nftopov1alpha1.NFInstance, instMap map[string]int) error {
	nfdeployedIdx := -1

	for idx, nfInst := range nfdeployed {
		if nfInst.ID == nfDeployedName {
			nfdeployedIdx = idx
			break
		}
	}

	if nfdeployedIdx == -1 {
		nfdeployedInst := nftopov1alpha1.NFDeployedInstance{}
		nfTemplate := nfInstance.NFTemplate
		nfdeployedInst.ID = nfDeployedName
		//nfdeployedInst.ClusterName = clusterName
		nfdeployedInst.NFType = string(nfTemplate.NFType)
		/*
		   nfdeployedInst.NFVendor = vendor
		   nfdeployedInst.NFVersion = version
		*/
		nfdeployedInst.NFInstanceName = nfInstance.Name
		nfdeployed = append(nfdeployed, nfdeployedInst)
		nfdeployedIdx = len(nfdeployed) - 1
	}

	fmt.Printf("SKW: nfdeployedIndex for %s is %d\n", nfDeployedName, nfdeployedIdx)
	// this is the continuous update case; so first build the neighbor list for this instance, then
	// for all the connected instance(s), update by appending the neighbor list
	/*
	 * TODO(s3wong): this now assumes each NFInstance from NFTopology.NFInstance will generate exactly
	 * one NF instance, an assumption that should hold true for R1
	 */
	neighborSlice, _ := neighborMap[nfInstance.Name]
	for _, neighbor := range neighborSlice {
		if neighborIdx, ok := instMap[neighbor.Name]; !ok {
			// neighbor packagerevision object not created yet
			fmt.Printf("SKW: no package for %s yet, continue...\n", neighbor.Name)
			continue
		} else {
			neighborInst := &nfdeployed[neighborIdx]
			nfdeployedInstance := &nfdeployed[nfdeployedIdx]
			con := nftopov1alpha1.NFConnectivity{}
			con.NeighborName = neighborInst.ID
			nfdeployedInstance.Connectivities = append(nfdeployedInstance.Connectivities, con)
			neighborCon := nftopov1alpha1.NFConnectivity{}
			neighborCon.NeighborName = nfdeployedInstance.ID
			neighborInst.Connectivities = append(neighborInst.Connectivities, neighborCon)
		}
	}
	return nil
}

// remove an nf from the NFInstance list
// TODO(): highly inefficient algorithm, it runs the list once to look for the index for the NF, and
// the remove utilizes an O(N) algorithm to shift element to maintain order
func RemoveNFfromList(nfName string, nfInstList []nftopov1alpha1.NFDeployedInstance) {
	var nfIdx int = -1
	for idx, nf := range nfInstList {
		if nfName == nf.ID {
			nfIdx = idx
			break
		}
	}
	if nfIdx == -1 {
		return
	}
	copy(nfInstList[nfIdx:], nfInstList[nfIdx+1:])
	nfInstList[len(nfInstList)-1] = nftopov1alpha1.NFDeployedInstance{}
	nfInstList = nfInstList[:len(nfInstList)-1]
}

/*
 * BuildPVS: builds a PVS object from NFTopology's NFInstance
 * NFClass needed for package reference
 */
func BuildPVS(nfInst *nftopov1alpha1.NFInstance, packageRef *nftopov1alpha1.PackageRevisionReference, nfTopo *nftopov1alpha1.NFTopology) *pvsv1alpha2.PackageVariantSet {
	retPVS := &pvsv1alpha2.PackageVariantSet{}

	retPVS.ObjectMeta.Name = nfInst.Name
	retPVS.ObjectMeta.Namespace = nfTopo.ObjectMeta.Namespace
	retPVS.ObjectMeta.Labels = map[string]string{
		"nftopology": nfTopo.ObjectMeta.Name,
		"nfinstance": nfInst.Name,
	}
	upstream := &pvv1alpha1.Upstream{}
	upstream.Repo = packageRef.RepositoryName
	upstream.Package = packageRef.PackageName
	upstream.Revision = packageRef.Revision

	retPVS.Spec.Upstream = upstream

	target := &pvsv1alpha2.Target{}
	objectSelector := &pvsv1alpha2.ObjectSelector{}
	template := &pvsv1alpha2.PackageVariantTemplate{}

	labelMap := map[string]string{}
	for k, v := range nfInst.ClusterSelector.MatchLabels {
		labelMap[k] = v
	}

	objectSelector.LabelSelector = metav1.LabelSelector{MatchLabels: labelMap}
	objectSelector.APIVersion = WORKLOADAPIVERSION
	objectSelector.Kind = WORKLOADKIND
	target.ObjectSelector = objectSelector

	template.Annotations = map[string]string{"approval.nephio.org/policy": "initial"}

	pipeline := &pvsv1alpha2.PipelineTemplate{}

	// TODO(): the Image path should NOT be hardcoded
	mutator := pvsv1alpha2.FunctionTemplate{}
	mutator.Image = "gcr.io/kpt-fn/search-replace:v0.2.0"
	mutator.ConfigMap = map[string]string{
		"by-path":      "spec.maxUplinkThroughput",
		"by-file-path": "**/capacity.yaml",
		"put-value":    nfInst.NFTemplate.Capacity.MaxUplinkThroughput.String(),
	}
	pipeline.Mutators = append(pipeline.Mutators, mutator)

	mutator = pvsv1alpha2.FunctionTemplate{}
	mutator.Image = "gcr.io/kpt-fn/search-replace:v0.2.0"
	mutator.ConfigMap = map[string]string{
		"by-path":      "spec.maxDownlinkThroughput",
		"by-file-path": "**/capacity.yaml",
		"put-value":    nfInst.NFTemplate.Capacity.MaxDownlinkThroughput.String(),
	}
	pipeline.Mutators = append(pipeline.Mutators, mutator)

	mutator = pvsv1alpha2.FunctionTemplate{}
	mutator.Image = "gcr.io/kpt-fn/search-replace:v0.2.0"
	mutator.ConfigMap = map[string]string{
		"by-path":      "spec.maxSessions",
		"by-file-path": "**/capacity.yaml",
		"put-value":    strconv.Itoa(nfInst.NFTemplate.Capacity.MaxSessions),
	}
	pipeline.Mutators = append(pipeline.Mutators, mutator)

	template.Pipeline = pipeline

	target.Template = template

	retPVS.Spec.Targets = append(retPVS.Spec.Targets, *target)

	return retPVS
}

func CompareNFPR(prList []porchv1alpha1.PackageRevision, nfList []nftopov1alpha1.NFDeployedInstance, nfInstanceName string) ([]string, []string) {
	prMap := make(map[string]bool)
	nfMap := make(map[string]bool)
	prSlice := make([]string, 0)
	nfSlice := make([]string, 0)

	for _, pr := range prList {
		prMap[pr.ObjectMeta.Name] = true
	}
	for _, nf := range nfList {
		if nf.NFInstanceName == nfInstanceName {
			nfMap[nf.ID] = true
		}
	}
	for key := range prMap {
		if _, ok := nfMap[key]; !ok {
			prSlice = append(prSlice, key)
		}
	}
	for key := range nfMap {
		if _, ok := prMap[key]; !ok {
			nfSlice = append(nfSlice, key)
		}
	}
	return prSlice, nfSlice
}

func (r *NFTopologyReconciler) FetchPR4NFInst(ctx context.Context, namespace string, nfInstName string) ([]porchv1alpha1.PackageRevision, error) {
	prList := &porchv1alpha1.PackageRevisionList{}
	/*
	   labelSelector := metav1.LabelSelector{
	       MatchLabels: map[string]string{"nfinstance:": nfInstName},
	   }
	*/
	labelMap := map[string]string{
		"nfinstance": nfInstName,
	}
	err := r.List(ctx, prList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(labels.Set(labelMap)),
	})
	if err == nil {
		return prList.Items, nil
	}
	ret := make([]porchv1alpha1.PackageRevision, 0)
	return ret, err
}

// +kubebuilder:rbac:groups=req.nephio.org,resources=nftopologies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=req.nephio.org,resources=nftopologies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=req.nephio.org,resources=nftopologies/finalizers,verbs=update
// +kubebuilder:rbac:groups=porch.kpt.dev,resources=packagerevisions,verbs=get;list;watch
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NFTopologyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.l = log.FromContext(ctx).WithValues("req", req)
	r.l.Info("Reconcile NFTopology")

	nfTopo := &nftopov1alpha1.NFTopology{}
	if err := r.Client.Get(ctx, req.NamespacedName, nfTopo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nfTopoNamespace := nfTopo.Namespace

	neighborMap := BuildNeighbormap(BuildAttachmentMap(nfTopo))
	deployedInstanceMap := BuildDeployedInstanceMap(nfTopo.Status.NFInstances)
	for _, nfInst := range nfTopo.Spec.NFInstances {

		packageRef := &nfInst.NFTemplate.NFPackageRef
		pvs := &pvsv1alpha2.PackageVariantSet{}
		if err := r.Client.Get(ctx, types.NamespacedName{Name: nfInst.Name, Namespace: nfTopoNamespace}, pvs); err == nil {
			/*
			 * If PVS corresponding to this NFInstance already exists, the reconciling may be:
			 * (a) the reconile has nothing to do with this NFInstance
			 * (b) this NFInstance is updated; for example, the network connectivity is changed
			 * (c) this reconcilation is triggered via packagerevision state change
			 * on all three cases, we examine the corresponding packagerevision to update connectivities, if needed
			 */
			if prList, err := r.FetchPR4NFInst(ctx, nfTopo.ObjectMeta.Namespace, nfInst.Name); err != nil {
				r.l.Error(err, fmt.Sprintf("Failed to get list of PackageRevision for nfInst %s: %s\n", nfInst.Name, err.Error()))
				return ctrl.Result{}, err
			} else {
				prAddList, nfRemoveList := CompareNFPR(prList, nfTopo.Status.NFInstances, nfInst.Name)
				if len(prAddList) > 0 {
					// new packages related to this NFInstance is created, add status info
					for _, pr := range prAddList {
						if err = BuildNFDeployed(pr, nfTopo, &nfInst, nfTopo.Status.NFInstances, neighborMap, deployedInstanceMap); err != nil {
							r.l.Error(err, fmt.Sprintf("Failed to connectivity list for package %s: %s\n", pr, err.Error()))
							return ctrl.Result{}, err
						}
					}
				}
				if len(nfRemoveList) > 0 {
					for _, nf := range nfRemoveList {
						RemoveNFfromList(nf, nfTopo.Status.NFInstances)
					}
				}
			}
		} else {
			/*
			 * NOTE: PVS update case is not necessarily; there are two fields in NFTopology that can
			 * affect PVS: (1) Name, and (2) cluster selector. The former should not happen, and the
			 * latter is an undefined behavior at this point
			 */
			pvs = BuildPVS(&nfInst, packageRef, nfTopo)
			r.l.Info(fmt.Sprintf("PVS for NF inst %s is %+v\n\n\n", nfInst.Name, pvs))
			if err := r.Client.Create(ctx, pvs); err != nil {
				r.l.Error(err, fmt.Sprintf("Failed to create PVS %s: %s\n", nfInst.Name, err.Error()))
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

/*
 * packageRevisionUpdate is the callback function when a change to packagerevision resource is detected
 * (the NFTopology Controller watches PackageRevision). The handler extracts the corresponding NFToplogy
 * resource and wrap it up as a request, which then triggers the reconciliation loop
 */
func (r *NFTopologyReconciler) packageRevisionUpdate(packageRevision client.Object) []reconcile.Request {
	pr := &porchv1alpha1.PackageRevision{}
	prName := packageRevision.GetName()
	prNS := packageRevision.GetNamespace()
	if err := r.Client.Get(context.TODO(), types.NamespacedName{Name: prName, Namespace: prNS}, pr); err != nil {
		return []reconcile.Request{}
	}
	prLabels := pr.ObjectMeta.Labels
	if topologyName, found := prLabels["nftopology"]; !found {
		return []reconcile.Request{}
	} else {
		requests := make([]reconcile.Request, 1)
		requests[0] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      topologyName,
				Namespace: prNS,
			},
		}
		return requests
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NFTopologyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nftopov1alpha1.NFTopology{}).
		/*
			Watches(&source.Kind{Type: &porchv1alpha1.PackageRevision{}},
				handler.EnqueueRequestsFromMapFunc(r.packageRevisionUpdate)).
		*/
		Complete(r)
}
