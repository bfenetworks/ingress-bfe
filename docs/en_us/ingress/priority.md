# Priority of route rules
If a request matches multiple ingress rules, BFE Ingress Controller will decide 

当请求能匹配到多条Ingress规则时，BFE Ingress Controller会按照以下优先级策略来选择规则：

-  根据主机名，优先选择主机名匹配更精确的规则；
-  主机名相同时，优先选择路径匹配更精确的规则；
-  主机名、路径均相同时，优先选择高级匹配条件更多的规则；
-  主机名、路径、高级匹配条件个数均相同时，优先选择高级匹配条件的优先级更高的规则；
   - 对于高级匹配条件，Cookie的优先级高于Header；

## 优先级示例
### 主机名精确优先
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
    - host: *.net
      http:
        paths:
          - path: /bar
            backend:
              serviceName: service2
              servicePort: 80
```
在以上示例中，针对`curl "http://example.net/bar"`产生的请求，优先匹配规则`host_priority1`

### 主机名相同，路径匹配精确优先
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
在以上示例中，针对`curl "http://example.net/bar/foo" -H "Key: value"`产生的请求，优先匹配规则`path_priority1`

### 主机名、路径均相同，高级匹配条件个数优先
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
在以上示例中，针对`curl "http://example.net/bar/foo" -H "Key: value"`产生的请求，优先匹配规则`cond_priority1`

### 主机名、路径、高级匹配条件个数均相同，按高级匹配条件的优先级排序
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
在以上示例中，针对`curl "http://example.net/bar/foo" -H "Header-key: value" --cookie "cookie-key: value"`产生的请求，优先匹配规则`multi_cond_priority2`，因为`Cookie`的优先级高于`Header`的优先级。

