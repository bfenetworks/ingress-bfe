# 配置优先级

## 路由配置冲突
当用户的ingress配置最终生成相同的路由规则的情况下（Host、Path、Header/Cookie完全相同），
BFE-Ingress将按照`创建时间优先`的原则使用先配置的路由规则。

对于因路由冲突导致的配置生成失败，可查找相应的 ingress-controller 错误日志。

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-A
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
#其中ingress-A先于配置ingress-B创建，则最终仅生效ingress-A。
```
## 跨 namespace 冲突

当 BFE-Ingress 在监听多个 namespace 时，如何判断是否存在冲突？如何处理？

以下场景可以为您解答这个问题：
* 场景 1: namespace 之间存在相同路由规则

    依照`创建时间优先`的原则处理，详见[路由配置冲突](#路由配置冲突)
   
* 场景 2: namespace 之间存在相同命名的资源（如 ingress/service ）
    
    不同 namespace 的相同资源名***不存在***冲突。
    * ingress 资源：controller 中通过 `${namespace}/${ingress}` 定位 ingress 资源；
    故不同 namespace 下的 ingress 资源无歧义，不存在冲突
    * service 资源：每个 ingress 资源 与 其中引用的 service 资源 拥有相同的 namespace 属性（默认为 default），
    controller 中通过 `${namespace}/${service}` 定位 service 资源；
    故不同 namespace 下的 service 资源无歧义，不存在冲突
   
## 状态回写
当前Ingress的合法性是在配置生效的过程才能感知，是一个异步过程。为了能给用户反馈当前Ingress是否生效，BFE-Ingress会将Ingress的实际生效状态回写到Ingress的一个Annotation当中。
**BFE-Ingress状态Annotation定义如下：**
```yaml
#bfe.ingress.kubernetes.io/bfe-ingress-status为BFE-Ingress预留的Annotation key，
#用于BFE-Ingress回写状态
# status; 表示当前ingress是否合法， 取值为：success -> ingress合法， error -> ingress不合法
# message; 当ingress不合法的情况下，message记录错误详细原因。
bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "", "message": ""}
```
**下面是BFE-Ingress状态回写的示例：**
`Ingress1`和`Ingress2`的路由规则完全一样(`Host:example.net, Path:/bar`)。
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "ingress1"
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
  name: "ingress2"
  namespace: production
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /foo
            backend:
              serviceName: service2
              servicePort: 80
```
根据路由冲突配置规则，`Ingress1`将生效，而`Ingress2`将被忽略。状态回写后，`Ingress1`的状态为success，而`Ingress2`的状态为fail。
```yaml
kind: Ingress
apiVersion: networking.k8s.io/v1beta1
metadata:
  name: "ingress1"
  namespace: production
  annotations:
    bfe.ingress.kubernetes.io/bfe-ingress-status: {"status": "success", "message": ""}
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
    bfe.ingress.kubernetes.io/bfe-ingress-status: |
    	{"status": "fail", "message": "conflict with production/ingress1"}
spec:
  rules:
    - host: example.net
      http:
        paths:
          - path: /foo
            backend:
              serviceName: service2
              servicePort: 80
```