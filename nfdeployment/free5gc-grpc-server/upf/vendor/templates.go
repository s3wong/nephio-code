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

package vendor

import (
	"bytes"
	"text/template"
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
      ({{ if .N4ENABLED }}
      n4network:
        enabled: true
        name: {{ .N4NETWORKNAME }}
        type: {{ .N4CNINAME }}
        masterIf: {{ .N4CNIMASTERINTF }}
        subnetIP: {{ .N4SUBNET }}
        cidr: {{ .N4CIDR }}
        gatewayIP: {{ .N4GATEWAY }}
        excludeIP: {{ .N4EXCLUDEIP }}
      {{ end }})
      ({{ if .N6ENABLED }}
      n6network:
        enabled: true
        name: {{ .N6NETWORKNAME }}
        type: {{ .N6CNINAME }}
        masterIf: {{ .N6CNIMASTERINTF }}
        subnetIP: {{ .N6SUBNET }}
        cidr: {{ .N6CIDR }}
        gatewayIP: {{ .N6GATEWAY }}
        excludeIP: {{ .N6EXCLUDEIP }}
      {{ end }})
      ({{ if .N9ENABLED }}
      n9network:
        enabled: true
        name: {{ .N9NETWORKNAME }}
        type: {{ .N9CNINAME }}
        masterIf: {{ .N9CNIMASTERINTF }}
        subnetIP: {{ .N9SUBNET }}
        cidr: {{ .N9CIDR }}
        gatewayIP: {{ .N9GATEWAY }}
        excludeIP: {{ .N9EXCLUDEIP }}
      {{ end }})
    deployUpf: true
`

var (
	configurationTemplate = template.Must(template.New("Free5gcHelmFluxConfig").Parse(free5gcHelmConfigMapData))
)

type configurationTemplateValues struct {
	NAME            string
	N3ENABLED       bool
	N3NETWORKNAME   string
	N3CNINAME       string
	N3CNIMASTERINTF string
	N3SUBNET        string
	N3CIDR          string
	N3GATEWAY       string
	N3EXCLUDEIP     string
	N4ENABLED       bool
	N4NETWORKNAME   string
	N4CNINAME       string
	N4CNIMASTERINTF string
	N4SUBNET        string
	N4CIDR          string
	N4GATEWAY       string
	N4EXCLUDEIP     string
	N6ENABLED       bool
	N6NETWORKNAME   string
	N6CNINAME       string
	N6CNIMASTERINTF string
	N6SUBNET        string
	N6CIDR          string
	N6GATEWAY       string
	N6EXCLUDEIP     string
	N9ENABLED       bool
	N9NETWORKNAME   string
	N9CNINAME       string
	N9CNIMASTERINTF string
	N9SUBNET        string
	N9CIDR          string
	N9GATEWAY       string
	N9EXCLUDEIP     string
}

func renderConfigurationTemplate(values configurationTemplateValues) (string, error) {
	var buffer bytes.Buffer
	if err := configurationTemplate.Execute(&buffer, values); err == nil {
		return buffer.String(), nil
	} else {
		return "", err
	}
}
