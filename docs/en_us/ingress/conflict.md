# Route Rule Conflict

## Definition
If Ingress configuration is created with Ingress resources containing the same Ingress rule (host, path and advanced conditions are all the same), a route rule conflict happens. 

## Conflict handling: first-created-resource-win principle

For those Ingress resources with route rule conflict, BFE Ingress Controller will follow first-created-resource-win principle and only takes the first created Ingress resource as valid.  

This principle will be followed when route rule conflict happens within a namespace or among different namespaces.

For those Ingress resources invalid due to route rule conflict, error messages will be written to its annotation, see [Ingress Status](validate-state.md).

## Example

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-A
  namespace: production
  annotations:
    kubernetes.io/ingress.class: bfe  
spec:
  rules:
  - host: example.foo.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          serviceName: service1
          servicePort: 80
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-B
  namespace: production
  annotations:
    kubernetes.io/ingress.class: bfe  
spec:
  rules:
  - host: example.foo.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          serviceName: service2
          servicePort: 80

```
In above configuration, there is conflict between ingress-A and ingress-B, and ingress-A is created before ingress-B. So only ingress-A will been created and take effect.

## Ingress status feedback
If a Ingress resource is ignored (not take effect) due to route rule conflict, after the ingress status is written back, the `status` in `annotation` will be set as “fail”, and `message` will tell which Ingress resource it has conflict with.

In previous example, `annotation` for ingress status will be like:


```yaml
metadata:
  annotations:
    bfe.ingress.kubernetes.io/bfe-ingress-status: |
    	{"status": "fail", "message": "conflict with production/ingress-A"}
```

For more information about ingress status, refer to [ingress status](validate-state.md)。

