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
	"fmt"
	"strings"
	"time"

	nephiov1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"
	"github.com/s3wong/nephio-code/nfdeploylib"
	pb "github.com/s3wong/nephio-code/nfdeploymentrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	apiv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	//grpcClient  *pb.NFDeploymentRPCClient
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

	conn, err := grpc.Dial("free5gc-upf:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error(err, "grpc connection to free5gc-upf failed")
		return reconcile.Result{}, err
	}
	defer conn.Close()

	c := pb.NewNFDeploymentRPCClient(conn)
	grpcCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

	/*
		if err := r.OnCreateUpdateResource(nfDeployment); err != nil {
			// retry
			return ctrl.Result{}, err
		}
	*/
	nfdeploymentPB := ConvertNFDeployment2gRPCMsg(nfDeployment)
	rsp, err := c.CreateUpdate(grpcCtx, nfdeploymentPB)
	if err != nil {
		log.Error(err, "CreateUpdate RPC failed")
		return ctrl.Result{}, err
	}
	fmt.Printf("Response: condition => %v   errormsg => %v\n", rsp.Condition, rsp.Errormsg)

	return ctrl.Result{}, nil
}

func BuildInterfaceConfig(nfDeployment *nephiov1alpha1.NFDeployment) []*pb.InterfaceConfig {
	var ret []*pb.InterfaceConfig
	for _, intf := range nfDeployment.Spec.Interfaces {
		intfConfigPB := pb.InterfaceConfig{}
		if intf.IPv4 != nil {
			ipv4PB := pb.IPv4{}
			if intf.IPv4.Gateway != nil {
				ipv4PB.Gateway = *intf.IPv4.Gateway
			} else {
				ipv4PB.Gateway = ""
			}
			ipv4PB.Address = intf.IPv4.Address
			intfConfigPB.Ipv4 = &ipv4PB
		} else {
			intfConfigPB.Ipv4 = nil
		}
		if intf.IPv6 != nil {
			ipv6PB := pb.IPv6{}
			if intf.IPv6.Gateway != nil {
				ipv6PB.Gateway = *intf.IPv6.Gateway
			} else {
				ipv6PB.Gateway = ""
			}
			ipv6PB.Address = intf.IPv6.Address
			intfConfigPB.Ipv6 = &ipv6PB
		} else {
			intfConfigPB.Ipv6 = nil
		}
		if intf.VLANID != nil {
			intfConfigPB.Vlanid = uint32(*intf.VLANID)
		} else {
			intfConfigPB.Vlanid = 0
		}
		ret = append(ret, &intfConfigPB)
	}
	return ret
}

func ConvertNFDeployment2gRPCMsg(nfDeployment *nephiov1alpha1.NFDeployment) *pb.NFDeployment {
	intfSlice := BuildInterfaceConfig(nfDeployment)
	ret := &pb.NFDeployment{
		Name:      nfDeployment.ObjectMeta.Name,
		Namespace: nfDeployment.ObjectMeta.Namespace,
		Capacity: &pb.Capacity{
			Maxuplinkthroughput:   nfDeployment.Spec.Capacity.MaxUplinkThroughput.String(),
			Maxdownlinkthroughput: nfDeployment.Spec.Capacity.MaxDownlinkThroughput.String(),
			Maxsessions:           int32(nfDeployment.Spec.Capacity.MaxSessions),
			Maxsubscribers:        int32(nfDeployment.Spec.Capacity.MaxSubscribers),
			Maxnfconnections:      uint32(nfDeployment.Spec.Capacity.MaxNFConnections),
		},
		Ifconfig: intfSlice,
	}
	return ret
}

func (r *NFDeploymentReconciler) OnCreateUpdateResource(nfDeployment *nephiov1alpha1.NFDeployment) error {
	switch nfDeployment.Spec.Provider {
	case "sdk.nephio.org/helm/flux":
		return HandleHelmFlux(r.Client, nfDeployment)
	default:
		return fmt.Errorf("Unknown provider for NFDeployment: %s", nfDeployment.Spec.Provider)
	}
}

func (r *NFDeploymentReconciler) OnDeleteResource(nfDeployment *nephiov1alpha1.NFDeployment) error {
	return nil
}

func HandleHelmFlux(c client.Client, nfDeployment *nephiov1alpha1.NFDeployment) error {
	// temp func
	namespace := nfDeployment.Namespace
	instanceName := nfDeployment.Name

	templateValues := new(configurationTemplateValues)
	n4ip, n4Gateway, n4net, err := nfdeploylib.GetFirstInterfaceConfigIPv4(nfDeployment.Spec.Interfaces, "n4")
	if err == nil {
		templateValues.N4ENABLED = true
		templateValues.N4SUBNET = n4ip
		templateValues.N4CIDR = strings.Split(n4net, "/")[1]
		templateValues.N4GATEWAY = n4Gateway
		templateValues.N4EXCLUDEIP = n4Gateway
		// TODO: hardcoded values
		templateValues.N4NETWORKNAME = "n4network"
		templateValues.N4CNINAME = "macvlan"
		templateValues.N4CNIMASTERINTF = "eth0"
	} else {
		fmt.Printf("SKW GetFirstInterfaceConfigIPv4 for n4 returns error %v\n", err)
		templateValues.N4ENABLED = false
	}

	n3ip, n3Gateway, n3net, err := nfdeploylib.GetFirstInterfaceConfigIPv4(nfDeployment.Spec.Interfaces, "n3")
	if err == nil {
		templateValues.N3ENABLED = true
		templateValues.N3SUBNET = n3ip
		templateValues.N3CIDR = strings.Split(n3net, "/")[1]
		templateValues.N3GATEWAY = n3Gateway
		templateValues.N3EXCLUDEIP = n3Gateway
		// TODO: hardcoded values
		templateValues.N3NETWORKNAME = "n3network"
		templateValues.N3CNINAME = "macvlan"
		templateValues.N3CNIMASTERINTF = "eth0"
	} else {
		fmt.Printf("SKW GetFirstInterfaceConfigIPv4 for n3 returns error %v\n", err)
		templateValues.N3ENABLED = false
	}

	n6ip, n6Gateway, n6net, err := nfdeploylib.GetFirstInterfaceConfigIPv4(nfDeployment.Spec.Interfaces, "n6")
	if err == nil {
		templateValues.N6ENABLED = true
		templateValues.N6SUBNET = n6ip
		templateValues.N6CIDR = strings.Split(n6net, "/")[1]
		templateValues.N6GATEWAY = n6Gateway
		templateValues.N6EXCLUDEIP = n6Gateway
		// TODO: hardcoded values
		templateValues.N6NETWORKNAME = "n6network"
		templateValues.N6CNINAME = "macvlan"
		templateValues.N6CNIMASTERINTF = "eth0"
	} else {
		fmt.Printf("SKW GetFirstInterfaceConfigIPv4 for n6 returns error %v\n", err)
		templateValues.N6ENABLED = false
	}

	fmt.Printf("SKW: nfDeployment is %v\n", nfDeployment.Spec)
	fmt.Printf("SKW: templateValues is %v\n", templateValues)

	if configuration, err := renderConfigurationTemplate(*templateValues); err != nil {
		return err
	} else {
		configMap := &apiv1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      instanceName,
			},
			Data: map[string]string{
				"values.yaml": configuration,
			},
		}
		fmt.Printf("ConfigMap generated is %v\n", configMap)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NFDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(new(nephiov1alpha1.NFDeployment)).
		Complete(r)
}
