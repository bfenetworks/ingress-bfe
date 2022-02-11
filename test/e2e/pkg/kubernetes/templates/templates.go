/*
Copyright 2020 The Kubernetes Authors.

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

package templates

import (
	"bytes"
	"fmt"
	text_template "text/template"
)

var k8sTemplates = map[string]string{
	"deployment": `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Name }}
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
  selector:
    matchLabels:
      app: {{ .MatchLabels }}
  template:
    metadata:
      labels:
        app: {{ .Labels }}
    spec:
      containers:
      - name: ingress-conformance-echo
        image: {{ .Image }}
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: INGRESS_NAME
          value: {{ .Ingress }}
        - name: SERVICE_NAME
          value: {{ .Service }}
        ports:
        - name: {{ .PortName }}
          containerPort: 3000
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
            scheme: HTTP
          initialDelaySeconds: 1
          periodSeconds: 1
          timeoutSeconds: 1
          successThreshold: 1
          failureThreshold: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 3000
            scheme: HTTP
          initialDelaySeconds: 1
          periodSeconds: 1
          timeoutSeconds: 1
          successThreshold: 1
          failureThreshold: 10
`,
	"service": `
apiVersion: v1
kind: Service
metadata:
  name: {{ .Name }}
spec:
  type: NodePort
  selector:
    app: {{ .Selector }}
  ports:
    - port: {{ .Port }}
      targetPort: 3000
`,
}

var templates = map[string]*text_template.Template{}

// Load parses templates required to deploy Kubernetes objects
func Load() error {
	for name, template := range k8sTemplates {
		tmpl, err := text_template.New(name).Parse(template)
		if err != nil {
			return err
		}

		templates[name] = tmpl
	}

	return nil
}

// Render executes a parsed template to the specified data object
func Render(name string, data interface{}) (string, error) {
	tmpl, ok := templates[name]
	if !ok {
		return "", fmt.Errorf("there is no template with name %v", name)
	}

	var tpl bytes.Buffer
	err := tmpl.Execute(&tpl, data)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
