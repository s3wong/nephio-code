/*
Copyright 2022.

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

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	resourcesv1alpha1 "nephio.io/code/api/v1alpha1"
	networkfunctionv1alpha1 "nephio.io/code/apis/networkfunction/v1alpha1"

	porchv1alpha1 "github.com/GoogleContainerTools/kpt/porch/api/porch/v1alpha1"
	"github.com/nephio-project/nephio-controller-poc/pkg/porch"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// NfNetworkResourceReconciler reconciles a NfNetworkResource object
type NfNetworkResourceReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	PorchClient client.Client
}

// TODO(user) hardcoded, not good
var DEPLOY_REPO = "nephio-test-deployment"
var ROOT_PACKAGE = "nephio-test-deployment-396ba414eb70a0d0de2b29f70f2ce9368dea2a25"

var UpfNadCfg string = `apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $NAD_NAME
  namespace: $NAD_NAMESPACE
spec:
  config: '{
      "cniVersion": "$CNI_VERSION",
      "plugins": [
        {
          "type": "$CNI_TYPE",
          "capabilities": { "ips": true },
          "master": "$NAD_MASTER",
          "mode": "bridge",
          "ipam": {
            "type": "static",
            "routes": [
              {
                "dst": "0.0.0.0/0",
                "gw": "$NAD_GW"
              }
            ]
          }
        }, {
          "capabilities": { "mac": true },
          "type": "tuning"
        }
      ]
    }'`

// TODO(user): hardcoded image path, should retrieve from package
var UpfDeployCfg string = `apiVersion: nfdeploy.nephio.io/v1alpha1
kind: UpfDeploy
metadata:
  name: $UPF_INS_NAME
  namespace: $UPF_INS_NS
spec:$UPF_INTERFACES
  capacity:
    downlinkThroughput: $UPF_DL_TP
    uplinkThroughput: $UPF_UL_TP
  imagePaths:
    upf: towards5gs/free5gc-upf:v3.1.1`

func (r *NfNetworkResourceReconciler) clonePackage(ctx context.Context, packageName string) (*porchv1alpha1.PackageRevision, error) {
	ns := "default"

	newPR := &porchv1alpha1.PackageRevision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PackageRevision",
			APIVersion: porchv1alpha1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
		},
		Spec: porchv1alpha1.PackageRevisionSpec{
			PackageName:    packageName,
			Revision:       "v1",
			RepositoryName: DEPLOY_REPO,
			Tasks: []porchv1alpha1.Task{
				{
					Type: porchv1alpha1.TaskTypeClone,
					Clone: &porchv1alpha1.PackageCloneTaskSpec{
						Upstream: porchv1alpha1.UpstreamPackage{
							UpstreamRef: &porchv1alpha1.PackageRevisionRef{
								Name: ROOT_PACKAGE,
							},
						},
					},
				},
			},
		},
	}

	err := r.PorchClient.Create(ctx, newPR)
	if err != nil {
		return nil, err
	}

	return newPR, nil
}

func (r *NfNetworkResourceReconciler) loadResourceList(ctx context.Context, pr *porchv1alpha1.PackageRevision) (*porchv1alpha1.PackageRevisionResources, error) {
	var resources porchv1alpha1.PackageRevisionResources
	if err := r.PorchClient.Get(ctx, client.ObjectKey{
		Namespace: pr.Namespace,
		Name:      pr.Name,
	}, &resources); err != nil {
		return nil, err
	}

	return &resources, nil
}

func (r *NfNetworkResourceReconciler) createUpfNadPackage(ctx context.Context, name string, namespace string,
	s *resourcesv1alpha1.UpfNadResourceSpec, newPR *porchv1alpha1.PackageRevision) error {
	resources, err := r.loadResourceList(ctx, newPR)
	if err != nil {
		return err
	}

	var pkgBuf *kio.PackageBuffer
	if pkgBuf, err = porch.ResourcesToPackageBuffer(resources.Spec.Resources); err != nil {
		return err
	}

	intfTypes := [3]string{"n3", "n4", "n6"}
	var cni, master, gw string
	for _, intfType := range intfTypes {
		switch intfType {
		case "n3":
			cni = s.N3Cni
			master = s.N3Master
			gw = s.N3Gw
		case "n4":
			cni = s.N4Cni
			master = s.N4Master
			gw = s.N4Gw
		case "n6":
			cni = s.N6Cni
			master = s.N6Master
			gw = s.N6Gw
		default:
			return errors.New("Unsupported UPF interface Type: " + intfType)
		}
		nadcfg := strings.Clone(UpfNadCfg)
		nadcfg = strings.Replace(nadcfg, "$NAD_NAME", name+"-"+intfType, 1)
		nadcfg = strings.Replace(nadcfg, "$NAD_NAMESPACE", namespace, 1)
		nadcfg = strings.Replace(nadcfg, "$CNI_VERSION", "0.3.1", 1)
		nadcfg = strings.Replace(nadcfg, "$CNI_TYPE", cni, 1)
		nadcfg = strings.Replace(nadcfg, "$NAD_MASTER", master, 1)
		nadcfg = strings.Replace(nadcfg, "$NAD_GW", gw, 1)

		obj, err := yaml.Parse(nadcfg)
		if err != nil {
			return err
		}

		pkgBuf.Nodes = append(pkgBuf.Nodes, obj)
	}

	if newResources, err := porch.CreateUpdatedResources(resources.Spec.Resources, pkgBuf); err != nil {
		return nil
	} else {
		resources.Spec.Resources = newResources
		if err = r.PorchClient.Update(context.TODO(), resources); err != nil {
			return err
		}
	}

	return nil
}

func (r *NfNetworkResourceReconciler) generateUpfDeployment(ctx context.Context, resourceNamespace string, name string, namespace string, newPR *porchv1alpha1.PackageRevision) error {
	upf := &networkfunctionv1alpha1.Upf{}

	if err := r.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: resourceNamespace}, upf); err != nil {
		fmt.Printf("Error: failed to get upf %s: %v\n", name, err.Error())
		return err
	}

	upfSpec := upf.Spec
	upfClassName := upfSpec.UpfClassName

	upfClass := &networkfunctionv1alpha1.UpfClass{}

	if err := r.Client.Get(ctx, client.ObjectKey{Name: upfClassName}, upfClass); err != nil {
		fmt.Printf("Error: failed to get upfClass %s: %v\n", upfClassName, err.Error())
		return err
	}

	resources, err := r.loadResourceList(ctx, newPR)
	if err != nil {
		return err
	}

	var pkgBuf *kio.PackageBuffer
	if pkgBuf, err = porch.ResourcesToPackageBuffer(resources.Spec.Resources); err != nil {
		return err
	}

	upfDeploy := strings.Clone(UpfDeployCfg)
	upfDeploy = strings.Replace(upfDeploy, "$UPF_INS_NAME", name, 1)
	upfDeploy = strings.Replace(upfDeploy, "$UPF_INS_NS", namespace, 1)
	// TODO(user): assuming just one network interface for each of N3/4/6
	// Note: the interface name needs to be short --- Linux does not support long intf name
	// build N3
	var interfaces = [3]string{"n3", "n4", "n6"}
	var interfaceBlock, interfaceName, intfKey string
	var nx *[]networkfunctionv1alpha1.NfEndpoint
	for _, intf := range interfaces {
		intfKey = ""
		switch intf {
		case "n3":
			nx = &upfSpec.N3.Endpoints
			interfaceName = "n3-"
			intfKey = `
  n3Interfaces:`
		case "n4":
			nx = &upfSpec.N4.Endpoints
			interfaceName = "n4-"
			intfKey = `
  n4Interfaces:`
		case "n6":
			interfaceName = "n6-"
			intfKey = `
  n6Interfaces:`
		default:
			// do nothing
		}
		if intf == "n6" {
			idx := 1
			for key, ep := range upfSpec.N6.Endpoints {
				item := fmt.Sprintf(`
  - dnn: %s
    interface:
      name: %s
      ipAddr:
        - %s
      gwAddr:
        - %s
    ipAddrPool: %s`, key, interfaceName+fmt.Sprintf("%d", idx), ep.IpEndpoints.Ipv4Addr[0], ep.IpEndpoints.Gwv4Addr, ep.IpAddrPool)
				idx += 1
				intfKey = intfKey + item
			}
			interfaceBlock += intfKey
		} else { // n3 or n4
			for idx, ep := range *nx {
				item := fmt.Sprintf(`
  - name: %s%d
    gwAddr:
    - %s
    ipAddr:
    - %s`, interfaceName, idx, ep.Gwv4Addr, ep.Ipv4Addr[0])
				intfKey = intfKey + item
			}
			interfaceBlock += intfKey
		}
	}
	upfDeploy = strings.Replace(upfDeploy, "$UPF_INTERFACES", interfaceBlock, 1)
	upfDeploy = strings.Replace(upfDeploy, "$UPF_DL_TP", upfClass.Spec.DownlinkThroughput, 1)
	upfDeploy = strings.Replace(upfDeploy, "$UPF_UL_TP", upfClass.Spec.UplinkThroughput, 1)

	fmt.Printf("upfDeploy is \n%v\n", upfDeploy)

	obj, err := yaml.Parse(upfDeploy)
	if err != nil {
		fmt.Printf("Error parsing UPF deploy string to yaml: %v\n", err.Error())
		return err
	}

	pkgBuf.Nodes = append(pkgBuf.Nodes, obj)

	if newResources, err := porch.CreateUpdatedResources(resources.Spec.Resources, pkgBuf); err != nil {
		return nil
	} else {
		resources.Spec.Resources = newResources
		if err = r.PorchClient.Update(context.TODO(), resources); err != nil {
			return err
		}
	}

	return nil
}

//+kubebuilder:rbac:groups=resources.nephio.io,resources=nfnetworkresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=resources.nephio.io,resources=nfnetworkresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=resources.nephio.io,resources=nfnetworkresources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NfNetworkResource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *NfNetworkResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	nfResource := &resourcesv1alpha1.NfNetworkResource{}
	err := r.Client.Get(ctx, req.NamespacedName, nfResource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// TODO(user): deleted after reconcile request --- need to handle
			return reconcile.Result{}, nil
		}
		fmt.Printf("Error: failed to get UpfDeploy %s\n", err.Error())
		return reconcile.Result{}, err
	}

	// look up Upf resource for IP addrerss block info
	name := nfResource.Name
	spec := nfResource.Spec
	namespace := spec.Namespace
	upfNad := spec.UpfNad
	resourceNamespace := nfResource.ObjectMeta.Namespace

	newPR, err := r.clonePackage(ctx, name+"-nad")
	if err != nil {
		return reconcile.Result{}, err
	}

	if err = r.createUpfNadPackage(ctx, name, namespace, &upfNad, newPR); err != nil {
		return reconcile.Result{}, err
	}

	secondPR, err := r.clonePackage(ctx, name+"-upf")
	if err != nil {
		return reconcile.Result{}, err
	}

	if err = r.generateUpfDeployment(ctx, resourceNamespace, name, namespace, secondPR); err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NfNetworkResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resourcesv1alpha1.NfNetworkResource{}).
		Complete(r)
}
