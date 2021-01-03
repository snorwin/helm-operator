> :warning: DISCLAIMER: This project is a work in progress and many things can still change.

# helm-operator
[![GitHub Action](https://img.shields.io/badge/GitHub-Action-blue)](https://github.com/features/actions)
[![Build](https://img.shields.io/github/workflow/status/snorwin/helm-operator/CI?label=build&logo=github)](https://github.com/snorwin/helm-operator/actions)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The **helm-operator** is an operator to depoly Helm Charts and Values stored as Custom Resources which allows you to dynamically change your Helm releases based on aribtariy Kubernetes resources.



```
apiVersion: helm.snorwin.io/v1
kind: Release
metadata:
  name: example
spec:
  chart:
    apiVersion: helm.snorwin.io/v1
    kind: Chart
    name: example
    namesapce: default
  values:
    - apiVersion: helm.snorwin.io/v1
      kind: Values
      name: example
      namesapce: default
```


```
apiVersion: helm.snorwin.io/v1
kind: Chart
metadata:
  name: example
spec:
  files:
    - data: |
        apiVersion: v1
        description: A Helm chart for a Kubernetes app
        name: hello
        version: 1.0
      name: Chart.yaml
    - data: |
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: hello-configmap
          namespace: {{ .Release.Namespace }}
        data:
          myvalue: "Hello {{ .Values.name }} :)"
      name: templates/configmap.yaml
```

```
apiVersion: helm.snorwin.io/v1
kind: Values
metadata:
  name: example
spec:
  file:
    data: |
      name: John Doe
    name: values-test.yaml
 ```
