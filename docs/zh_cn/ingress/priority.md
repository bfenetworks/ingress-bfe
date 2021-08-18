# 优先级说明
当请求满足多条导流规则的情况下，BFE-Ingress会按照如下的优先级进行排序，使用优先级最高的导流规则：
- 优先满足域名匹配；
- 域名相同的场景下，优先满足更精确的路径匹配；
- 域名、路径相同的场景下，优先满足匹配条件更多的场景；
- 域名、路径、匹配条件个数相同情况下，按照匹配条件的固定顺序确定优先级；
   - Cookie的优先级高于Header；
   
## 优先级示例
### 域名优先
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "host_priority1"
  namespace: production

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
  name: "host_priority2"
  namespace: production

spec:
  rules:
    - host: example2.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service2
              servicePort: 80
```
针对`curl "http://example.net/bar"`优先匹配规则`host_priority1`

### 域名相同，优先路径
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "path_priority1"
  namespace: production

spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /bar/foo
            backend:
              serviceName: service1
              servicePort: 80
---
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "path_priority2"
  namespace: production
  bfe.ingress.kubernetes.io/router.header: "key: value"
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
针对`curl "http://example.net/bar/foo" -H "Key: value"`优先匹配规则`path_priority1`

### 路径优先规则个数
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "cond_priority1"
  namespace: production
  bfe.ingress.kubernetes.io/router.header: "key: value"
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
  name: "cond_priority1"
  namespace: production
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
针对`curl "http://example.net/bar/foo" -H "Key: value"`优先匹配规则`cond_priority1`

### 规则个数相同，按固定顺序排序
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "multi_cond_priority1"
  namespace: production
  bfe.ingress.kubernetes.io/router.header: "header-key: value"
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
  name: "multi_cond_priority2"
  namespace: production
  bfe.ingress.kubernetes.io/router.cookie: "cookie-key: value"
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
例如当前BFE-Ingress中`Cookie`的优先级高于`Header`的优先级。
针对`curl "http://example.net/bar/foo" -H "Header-key: value" --cookie "cookie-key: value"`优先匹配规则`multi_cond_priority2`

