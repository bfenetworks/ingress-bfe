# Validate State

## Validate state response
Validating the Ingress config is an async process and the result can only be returned after resources applied.

In order to response the result of whether the Ingress takes effect, BFE Ingress Controller will write the validate state of the Ingress back to its annotations. 

**BFE Ingress Controller defines the annotation for validate state as follow:**

```yaml
#bfe.ingress.kubernetes.io/bfe-ingress-status is the reserved Annotation key of BFE Ingress Controller，
#used for validate state response
# status; indicate if this ingress is valid, value can be: success -> ingress is valid and takes effect， error -> ingress is not valid
# message; if ingress is not valid, error messages will be recoreded
bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "", "message": ""}
```
## Example

Below example shows the validate state response of two ingress with route rules conflict
`Ingress1` and `Ingress2` have one identical route rule (`Host:example.net, Path:/bar`)

```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "ingress1"
  namespace: production
  annotations:
    kubernetes.io/ingress.class: bfe 
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service1
              servicePort: 80
---
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "ingress2"
  namespace: production
  annotations:
    kubernetes.io/ingress.class: bfe 
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service2
              servicePort: 80
```
According to conflict handling principle for [route rule conflict](conflict.md), `Ingress1` will take effect and `Ingress2` will be ignored. After validate state responsed, `status` of `Ingress1` will be "success" and for `Ingress2` it will be "fail".
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "ingress1"
  namespace: production
  annotations:
    kubernetes.io/ingress.class: bfe   
    bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "success"}
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service1
              servicePort: 80
---
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "ingress2"
  namespace: production
  annotations:
    kubernetes.io/ingress.class: bfe   
    bfe.ingress.kubernetes.io/bfe-ingress-status: |
    	{"status": "fail", "message": "conflict with production/ingress1"}
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service2
              servicePort: 80
```