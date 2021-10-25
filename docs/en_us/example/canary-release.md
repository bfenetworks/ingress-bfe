# Canary Release

## Introduction
BFE Ingress Controller support `Header/Cookie` based "canary release" by configuring`Annotation`.

## Config Example
* Original ingress config as follows, which will forward matched requests to `service`：
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "original"
  namespace: production

spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            pathType: Exact
            backend:
              serviceName: service
              servicePort: 80
```

* Canary release is required and interested requests should be forwarded to a new service `service2`.
* To implement this, create a new ingress, with header or cookie information of  interested requests included in annotations.
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "greyscale"
  namespace: production
  annotations:
    bfe.ingress.kubernetes.io/router.cookie: "key: value"
    bfe.ingress.kubernetes.io/router.header: "Key: Value"

spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            pathType: Exact
            backend:
              serviceName: service2
              servicePort: 80

```
* Based on above config, BFE Ingress Controller will
1. requests with `host == example.net && path == /bar && cookie[key] == value && Header[Key] == Value`,
   forwarded to service `service-new`
1. other request with `host == example.net && path == /bar`，
   forwarded to service `service`
