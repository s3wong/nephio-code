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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pvv1alpha1 "github.com/GoogleContainerTools/kpt/porch/controllers/packagevariants/api/v1alpha1"
	pvsv1alpha2 "github.com/GoogleContainerTools/kpt/porch/controllers/packagevariantsets/api/v1alpha2"
	"github.com/go-logr/logr"
	deployv1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"
	reqv1alpha1 "github.com/nephio-project/api/nf_requirements/v1alpha1"
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
 * BuildPVS: builds a PVS object from NFTopology's NFInstance
 * NFClass needed for package reference
 */
func BuildPVS(nfInst *reqv1alpha1.NFInstance, packageRef *reqv1alpha1.PackageRevisionReference) *pvsv1alpha2.PackageVariantSet {
	retPVS := &pvsv1alpha2.PackageVariantSet{}

	retPVS.ObjectMeta.Name = nfInst.Name
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

    template.Annotations = map[string]string{"approval.nephio.org/policy" : "initial"}

    pipeline := &pvsv1alpha2.PipelineTemplate{}

    mutator := &pvsv1alpha2.FunctionTemplate{}
    mutator.Image = "gcr.io/kpt-fn/search-replace:v0.2.0"
    mutator.ConfigMap = map[string]string{
        "by-path": "spec.maxUplinkThroughput",
        "by-file-path": "**/capacity.yaml",
        "put-value": nfInst.NFTemplate.Capacity.Spec.MaxUplinkThroughput
    }
    pipeline.Mutators = append(pipeline.Mutators, mutator)

    mutator := &pvsv1alpha2.FunctionTemplate{}
    mutator.Image = "gcr.io/kpt-fn/search-replace:v0.2.0"
    mutator.ConfigMap = map[string]string{
        "by-path": "spec.maxDownlinkThroughput",
        "by-file-path": "**/capacity.yaml",
        "put-value": nfInst.NFTemplate.Capacity.Spec.MaxDownlinkThroughput
    }
    pipeline.Mutators = append(pipeline.Mutators, mutator)

    mutator := &pvsv1alpha2.FunctionTemplate{}
    mutator.Image = "gcr.io/kpt-fn/search-replace:v0.2.0"
    mutator.ConfigMap = map[string]string{
        "by-path": "spec.maxSessions",
        "by-file-path": "**/capacity.yaml",
        "put-value": nfInst.NFTemplate.Capacity.Spec.MaxSessions
    }
    pipeline.Mutators = append(pipeline.Mutators, mutator)

	retPVS.Spec.Targets = append(retPVS.Spec.Targets, *target)

	return retPVS
}

//+kubebuilder:rbac:groups=req.nephio.org,resources=nftopologies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=req.nephio.org,resources=nftopologies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=req.nephio.org,resources=nftopologies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NFTopologyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.l = log.FromContext(ctx).WithValues("req", req)
	r.l.Info("Reconcile NFTopology")

	nfTopo := &reqv1alpha1.NFTopology{}
	if err := r.Client.Get(ctx, req.NamespacedName, nfTopo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nfDeployed := &deployv1alpha1.NFDeployed{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: nfTopo.ObjectMeta.Namespace, Name: nfTopo.ObjectMeta.Name}, nfDeployed); err != nil {
		r.l.Info(fmt.Sprintf("NF Deployed object for %s not created, continue...\n", nfTopo.ObjectMeta.Name))
	}

	neighborMap := BuildNeighbormap(BuildAttachmentMap(nfTopo))
	deployedInstanceMap := BuildDeployedInstanceMap(nfDeployed)
	continueReconciling := false
	for _, nfInst := range nfTopo.Spec.NFInstances {

		packageRef := &nfInst.NFTemplate.NFPackageRef

		pvs := BuildPVS(&nfInst, packageRef)
		r.l.Info(fmt.Sprintf("PVS for NF inst %s is %+v\n\n\n", nfInst.Name, pvs))
		if err := r.Client.Create(ctx, pvs); err != nil {
			r.l.Error(err, fmt.Sprintf("Failed to create PVS %s: %s\n", nfInst.Name, err.Error()))
		} else {
			if err := r.Client.Update(ctx, pvs); err != nil {
				r.l.Error(err, fmt.Sprintf("Failed to create PVS %s: %s\n", nfInst.Name, err.Error()))
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NFTopologyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&reqv1alpha1.NFTopology{}).
		Complete(r)
}
