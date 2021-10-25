# Load banlancing between Services
## Introduction

For `Service`s that providing the same service (called Sub-Services),  BFE Ingress Controller supports load balancing between them, based on weight configured for each `Service`.

## Configuration

BFE Ingress Controller use `Annotation` to support load-balancing between multiple Sub-Services：

- in `annotations`

  - configurate weight for each Sub-Service.

  - define a `Service` name for the service they provided together：

    ``` yaml
    bfe.ingress.kubernetes.io/balance.weight: '{"service": {"sub-service1":80, "sub-service2":20}}'
    ```

- in `rules`

  - set the `serviceName` of `backend` as the `Service` name in `Annotation`, and set the `servicePort`.

## Example

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: tls-example-ingress
  annotations:
    kubernetes.io/ingress.class: bfe    
    bfe.ingress.kubernetes.io/balance.weight: '{"service": {"service1":80, "service2":20}}'
spec:
  tls:
  - hosts:
      - https-example.foo.com
    secretName: testsecret-tls
  rules:
  - host: https-example.foo.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          serviceName: service
          servicePort: 80
```
