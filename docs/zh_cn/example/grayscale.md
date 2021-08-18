# BEF-Ingress支持 Header/Cookie 灰度发布
### 配置说明
BFE-Ingress通过`Ingress Annotation`的方式支持`Header/Cookie`灰度发布功能，配置如下：

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
              serviceName: service-new
              servicePort: 80
---
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
              serviceName: service-old
              servicePort: 80
```
基于上面的配置，BFE将会
1. 若满足 `host == example.net && path == /bar && cookie[key] == value && Header[Key] == Value`，
   则分流到`service-new`集群
1. 否则，若仅满足 `host == example.net && path == /bar`，
   则分流到`service-old`集群
