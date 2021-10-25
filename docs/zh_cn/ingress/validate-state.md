# 生效状态

## 生效状态反馈
Ingress配置的合法性检查是一个异步过程，检查结果在配置生效的过程中才能返回。

为了能给用户反馈当前Ingress是否生效，BFE Ingress Controller会将Ingress的实际生效状态回写到Ingress的一个Annotation当中。
**BFE Ingress Controller的状态Annotation定义如下：**

```yaml
#bfe.ingress.kubernetes.io/bfe-ingress-status为BFE-Ingress预留的Annotation key，
#用于BFE-Ingress反馈生效状态
# status: 表示当前ingress是否合法， 取值为：success -> ingress合法， error -> ingress不合法
# message: 当ingress不合法的情况下，message记录错误详细原因。
bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "", "message": ""}
```
## 示例

下面是BFE-Ingress生效状态反馈的一个示例，展示发生路由冲突的两个Ingress资源的生效状态反馈。
`Ingress1`和`Ingress2`的路由规则完全一样(`Host:example.net, Path:/bar`)。

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
根据[路由冲突处理原则](conflict.md)，`Ingress1`将生效，而`Ingress2`将被忽略。状态回写反馈后，`Ingress1`的状态为success，而`Ingress2`的状态为fail。
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
