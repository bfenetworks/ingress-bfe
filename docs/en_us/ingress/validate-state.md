# Ingress status

## Feedback for ingress status
The validation of the Ingress configuration is an asynchronous process. The status can only be returned after the configuation has taken effect.

In order to provide feedback for ingress status, BFE Ingress Controller will write status back to its annotations. 

**BFE Ingress Controller defines the annotation for status as follow:**

```yaml
#bfe.ingress.kubernetes.io/bfe-ingress-status is the reserved Annotation key of BFE Ingress Controller
#used for status feedback. 
# status: success -> ingress is valid, error -> ingress is invalid.
# message: if ingress is invalid, error messages will be recorded
bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "", "message": ""}
```
## Example

The following example shows the status of two ingresses with route rules conflict.
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
According to [principles of handling route rule conflict](conflict.md), `Ingress1` will take effect and `Ingress2` will be ignored. After the status is returned, `status` of `Ingress1` will be "success" and status of `Ingress2` it will be "fail".
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
