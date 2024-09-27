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
	"errors"
	"fmt"
	"time"

	nephiov1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"
	pb "github.com/s3wong/nephio-code/nfdeploymentrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	//apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	//workloadv1alpha1 "nephio-sdk/api/v1alpha1"
)

// NFDeploymentReconciler reconciles a NFDeployment object
type NFDeploymentReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	GrpcClients map[string]pb.NFDeploymentRPCClient
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
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(nfDeployment, finalizerName)
			if err := r.Update(ctx, nfDeployment); err != nil {
				return ctrl.Result{}, err
			}
		}

		// delete successful
		return ctrl.Result{}, nil
	}

	if err := r.OnCreateUpdateResource(nfDeployment); err != nil {
		// retry
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func ConvertNFDeployment2gRPCMsg(nfDeployment *nephiov1alpha1.NFDeployment) *pb.NFDeployment {
	ret := &pb.NFDeployment{
		ApiVersion: nfDeployment.TypeMeta.APIVersion,
		Name:       nfDeployment.ObjectMeta.Name,
		Namespace:  nfDeployment.ObjectMeta.Namespace,
	}
	return ret
}

func (r *NFDeploymentReconciler) GetVendorSvcNamePort(nfDeployment *nephiov1alpha1.NFDeployment) (string, string) {
	// TODO(s3wong): Get vendor specific service name and gRPC port number
	return "free5gc-upf", "50051"
}

func (r *NFDeploymentReconciler) EnsureGrpc2Backend(nfDeployment *nephiov1alpha1.NFDeployment) (pb.NFDeploymentRPCClient, error) {
	c, ok := r.GrpcClients[nfDeployment.Spec.Provider]
	if ok {
		svcName, grpcPort := r.GetVendorSvcNamePort(nfDeployment)
		if len(svcName) == 0 || len(grpcPort) == 0 {
			return nil, errors.New(fmt.Sprintf("Service name and gRPC port for %s not found\n", nfDeployment.Spec.Provider))
		}
		grpcUrl := svcName + ":" + grpcPort
		if conn, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
			return nil, err
		} else {
			c = pb.NewNFDeploymentRPCClient(conn)
			r.GrpcClients[nfDeployment.Spec.Provider] = c
		}
	}
	return c, nil
}

func (r *NFDeploymentReconciler) OnCreateUpdateResource(nfDeployment *nephiov1alpha1.NFDeployment) error {
	if c, err := r.EnsureGrpc2Backend(nfDeployment); err != nil {
		return nil
	} else {
		grpcCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		nfdeploymentPB := ConvertNFDeployment2gRPCMsg(nfDeployment)
		rsp, err := c.CreateUpdate(grpcCtx, nfdeploymentPB)
		if err == nil {
			// TODO(s3wong): update NFDeployment status
			fmt.Printf("SKW: CreateUpdate gRPC to %s returns %v\n", nfDeployment.Spec.Provider, rsp)
		}
		return err
	}
}

func (r *NFDeploymentReconciler) OnDeleteResource(nfDeployment *nephiov1alpha1.NFDeployment) error {
	if c, err := r.EnsureGrpc2Backend(nfDeployment); err != nil {
		return nil
	} else {
		grpcCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		nfdeploymentPB := ConvertNFDeployment2gRPCMsg(nfDeployment)
		rsp, err := c.Delete(grpcCtx, nfdeploymentPB)
		if err == nil {
			// TODO(s3wong): update NFDeployment status
			fmt.Printf("SKW: CreateUpdate gRPC to %s returns %v\n", nfDeployment.Spec.Provider, rsp)
		}
		return err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NFDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(nephiov1alpha1.NFDeployment)).
		Complete(r)
}
