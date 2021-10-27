# Canary Release

## Introduction
BFE Ingress Controller supports `Header/Cookie` based "canary release" by configuring`Annotation`.

## Config Example
* Original ingress configuration is shown as follows. Ingress will forward matched requests to `service`：
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

* Canary release is applied and interested requests should be forwarded to a new service `service2`.
* To achieve this, create a new ingress, with header or cookie information of interested requests included in annotations.
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
* Based on above configuration, BFE Ingress Controller will
1. forward requests with `host == example.net && path == /bar && cookie[key] == value && Header[Key] == Value`
   to service `service-new`
1. forward other requests with `host == example.net && path == /bar`
   to service `service`
