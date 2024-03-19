/*
Copyright 2024 The Nephio Authors.

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
	"bytes"
	"text/template"

	nephiov1alpha1 "github.com/nephio-project/api/nf_deployments/v1alpha1"
)

const free5gcHelmConfigMapData = `
    global:
      name: {{ .NAME }}
      userPlaneArchitecture: single  # possible values are "single" and "ulcl"
      #Global network parameters
      ({{ if .N3ENABLED }}
      n3network:
        enabled: true
        name: {{ .N3NETWORKNAME }}
        type: {{ .N3CNINAME }}
        masterIf: {{ .N3CNIMASTERINTF }}
        subnetIP: {{ .N3SUBNET }}
        cidr: {{ .N3CIDR }}
        gatewayIP: {{ .N3GATEWAY }}
        excludeIP: {{ .N3EXCLUDEIP }}
      {{ end }})
      n4network:
        enabled: true
        name: n4network
        type: ipvlan
        masterIf: eth0
        subnetIP: 10.100.50.240
        cidr: 29
        gatewayIP: 10.100.50.246
        excludeIP: 10.100.50.246
      ({{ if .N6ENABLED }}
      n6network:
        enabled: true
        name: n6network
        type: ipvlan
        masterIf: eth1
        subnetIP: 10.100.100.0
        cidr: 24
        gatewayIP: 10.100.100.1
        excludeIP: 10.100.100.254
      {{ end }})
      ({{ if .N9ENABLED }}
      n9network:
        enabled: true
        name: n9network
        type: ipvlan
        masterIf: eth0
        subnetIP: 10.100.50.224
        cidr: 29
        gatewayIP: 10.100.50.230
        excludeIP: 10.100.50.230
      {{ end }})
    deployUpf: true
`

var (
	helmConfigTemplate = template.Must(template.New("Free5gcHelmFluxConfig").Parse(free5gcHelmConfigMapData))
)

type configurationTemplateValues struct {
	PFCP_IP string
	GTPU_IP string
	N6cfg   []nephiov1alpha1.NetworkInstance
	N6gw    string
}

func renderConfigurationTemplate(values configurationTemplateValues) (string, error) {
	var buffer bytes.Buffer
	if err := configurationTemplate.Execute(&buffer, values); err == nil {
		return buffer.String(), nil
	} else {
		return "", err
	}
}

func renderWrapperScriptTemplate(values configurationTemplateValues) (string, error) {
	var buffer bytes.Buffer
	if err := wrapperScriptTemplate.Execute(&buffer, values); err == nil {
		return buffer.String(), nil
	} else {
		return "", err
	}
}
