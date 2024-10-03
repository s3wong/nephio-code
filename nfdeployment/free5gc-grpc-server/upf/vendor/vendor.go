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

package vendor

import (
    "context"
    "fmt"
    "strings"

    "github.com/nephio-experimental/nephio-sdk-poc/nfdeploylib"
    pb "github.com/nephio-experimental/nephio-sdk-poc/nfdeploymentrpc"
    nephiov1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"

    apiv1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "sigs.k8s.io/controller-runtime/pkg/client"
    //"sigs.k8s.io/controller-runtime/pkg/log"
)

func CreateUpdate(ctx context.Context, c client.Client, nfDeployment *nephiov1alpha1.NFDeployment) (*pb.NFDeploymentResponse, error) {
    namespace := nfDeployment.Namespace
    instanceName := nfDeployment.Name

    templateValues := new(configurationTemplateValues)
    templateValues.NAME = instanceName
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
        return nil, err
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

    return nil, nil
}

func Delete(ctx context.Context, c client.Client, nfDeployment *nephiov1alpha1.NFDeployment) (*pb.                     NFDeploymentResponse, error) {
    return nil, nil
}
