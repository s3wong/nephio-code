/*
Copyright 2024.

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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	workloadv1alpha1 "nephio-sdk/api/v1alpha1"
)

// NFDeploymentReconciler reconciles a NFDeployment object
type NFDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=workload.nephio.org,resources=nfdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workload.nephio.org,resources=nfdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=workload.nephio.org,resources=nfdeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NFDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *NFDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("NFDeployment", req.NamespacedName)

    nfDeployment := new(nephiov1alpha1.NFDeployment)
    err := r.Client.Get(ctx, req.NamespacedName, nfDeployment)
    if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("NFDeployment resource not found, ignoring because object must be deleted")
			return reconcile.Result{}, nil
		}
		log.Error(err, "Failed to get NFDeployment")
		return reconcile.Result{}, err
	}

    // name of custom finalizer
    finalizerName := "nfdeployment.nephio.org/finalizer"

    if nfDeployment.ObjectMeta.DeletionTimestamp.IsZero() {
        if !controllerutil.ContainsFinalizer(nfDeployment, finalizerName) {
            controllerutil.AddFinalizer(nfDeployment, finalizerName)
            if err := r.Update(ctx, nfDeployment); err != nil {
                return ctrl.Result{}, err
            }
        }
    } else {
        // The object is being deleted
        if controllerutil.ContainsFinalizer(nfDeployment, finalizerName) {
            // our finalizer is present, so lets handle any external dependency
            if err := r.OnDeleteResource(nfDeployment); err != nil {
                // retry
                return ctl.Result{}, err
            }
        }

        // delete successful
        return ctrl.Result{}, nil
    }

    if err := r.OnCreateUpdateResource(nfDeployment); err != nil {
        // retry
        return ctl.Result{}, err
    }

	return ctrl.Result{}, nil
}

func (r *NFDeploymentReconciler) OnCreateUpdateResource(nfDeployment *nephiov1alpha1.NFDeployment) error {
    switch nfDeployment.Spec.Provider {
    case "sdk.nephio.org/helm/flux":
        // TODO(s3wong): temp function, should use plugin
        return HandleHelmFlux(r.Client, nfDeployment)
    default:
        return fmt.Errorf("Unknown provider for NFDeployment: %s", nfDeployment.Spec.Provider)
    }
}

func (r *NFDeploymentReconciler) OnDeleteResource(nfDeployment *nephiov1alpha1.NFDeployment) error {
    return nil
}

func HandleHelmFlux(c client.Client, nfDeployment *nephiov1alpha1.NFDeployment) {
    // temp func

}

// SetupWithManager sets up the controller with the Manager.
func (r *NFDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workloadv1alpha1.NFDeployment{}).
		Complete(r)
}
