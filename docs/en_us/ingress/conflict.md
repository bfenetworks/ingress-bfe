# Route Rule Conflict

## Definition
If Ingress configuration will create Ingress resources containing a same Ingress rule (host, path and advanced conditions are all the same), a route rule conflict happens. 

## Conflict handling: first-created-resource-win principle

For those Ingress resources with route rule conflict, BFE Ingress Controller will follow first-created-resource-win principle and only takes the first created Ingress resource as valid.  

Route rule conflicts within a namespace or among different namespaces will both follow this principle.

For those Ingress resources that not taken as valid by BFE Ingress Controller due to route rule conflict, related error messages will be writen to its annotation, see [Validate State](validate-state.md) `annotations`.

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
In above config, ingress-A and ingress-B have conflict, and ingress-A is created before ingress-B. So only ingress-A will been created and take effect.

## Validate state writeback
If a Ingress resource is ignored (not take effect) due to route rule conflict, after validate state writeback, the `status` of validate state `annotation` will be set as “fail”, and `message` will tell which Ingres resource it conflict with.

In previous example, validate state `annotation` will be like:


```yaml
metadata:
  annotations:
    bfe.ingress.kubernetes.io/bfe-ingress-status: |
    	{"status": "fail", "message": "conflict with production/ingress-A"}
```

For more information about validate state, refer to [Validate state](validate-state.md)。

