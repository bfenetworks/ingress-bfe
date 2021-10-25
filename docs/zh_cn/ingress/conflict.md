# 路由冲突处理

## 路由冲突的定义

当Ingress配置最终生成包含相同的Ingress规则（Host、Path、高级匹配条件均完全相同）的Ingress资源的情况下，会产生路由冲突。

## 处理原则：最先创建的生效

对存在路由冲突的多个Ingress资源，BFE Ingress Controller将按照"最先创建的生效"的原则，仅使最早被创建的Ingress资源生效。

在同一个namespace之间，或在多个namespace之间的路由冲突，均按照此原则处理。

对于因路由冲突导致的没有在BFE Ingress Controller中生效的Ingress资源，可在[生效状态](validate-state.md)反馈的Annotation中查找相应的错误消息。

## 示例

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
在以上配置中，ingress-A和ingress-B冲突，ingress-A先于ingress-B创建，所以最终仅ingress-A生效。

## 生效状态反馈
若一个Ingress资源因路由冲突而被忽略（未生效），生效状态反馈回写后，对于生效状态的注解的status会被设为“fail”，message中会包含和哪个Ingress资源发生了冲突。

在前面的示例中，ingress-B的生效状态的注解将会如下面所示：


```yaml
metadata:
  annotations:
    bfe.ingress.kubernetes.io/bfe-ingress-status: |
    	{"status": "fail", "message": "conflict with production/ingress-A"}
```

更多生效状态的说明见[生效状态](validate-state.md)。

