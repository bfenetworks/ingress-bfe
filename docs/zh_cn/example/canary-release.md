# 支持灰度发布

## 说明
BFE-Ingress-controller支持通过配置`Annotation`，实现基于`Header/Cookie`的灰度发布功能。

## 配置示例
* 初始的ingress配置如下，请求转发到服务`service`：
```yaml
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
              serviceName: service
              servicePort: 80
```

* 做灰度发布，对特定请求，转发到新的服务`service2`。
* 为实现上述目的，创建一个新的ingress，在annotations中包含特定请求的header或cookie的信息。
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
              serviceName: service2
              servicePort: 80

```
* 基于上面的配置，BFE对
1. 满足 `host == example.net && path == /bar && cookie[key] == value && Header[Key] == Value`，
   则转发到`service-new`集群
1. 仅满足 `host == example.net && path == /bar`，
   仍转发到`service`集群
