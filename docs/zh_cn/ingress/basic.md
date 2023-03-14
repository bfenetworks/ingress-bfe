# 配置指南

## 概述
通过配置K8S Ingress资源，可以定义K8S集群中的服务对外暴露时的流量路由规则。更多Ingress相关信息，可参考[Ingress介绍]。

我们提供了配置文件示例 [ingress.yaml](../../deploy/ingress.yaml)，可供配置时参考。

## Ingress示例
### 示例1
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: simple-ingress
  annotations:
    kubernetes.io/ingress.class: bfe  
spec:
  rules:
  - host: whoami.com
    http:
      paths:
      - path: /testpath
        pathType: Prefix
        backend:
          service:
            name: whoami
            port:
              number: 80
```
上述示例中定义了一个 Ingress 资源

- 设置`kubernetes.io/ingress.class`为`bfe`，标识该Ingress由BFE Ingress Controller处理。
- 定义了一条简单的路由规则：若请求流量的域名为 `whoami.com`，路径前缀为 `/testpath`，则将流量转发给`whoami` Service 的80端口处理

### 示例2
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: complex-ingress
  namespace: my-namespace
  annotations:
    kubernetes.io/ingress.class: bfe
    bfe.ingress.kubernetes.io/loadbalance: '{"foo": {"sub-foo1":80, "sub-foo2":20}}'
    bfe.ingress.kubernetes.io/router.cookie: "Session: 123"
    bfe.ingress.kubernetes.io/router.header: "Content-Language: zh-cn"
spec:
  tls:
  - hosts:
      - foo.com
    secretName: secret-foo-com
  rules:
  - host: foo.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: foo
            port:
              number: 80
      - path: /bar
        pathType: Exact
        backend:
          service:
            name: bar
            port:
              number: 80
```
上述 Ingress定义了 2 条复杂的路由规则，并且为 `foo.com` 配置了证书。其中，annotations定义了BFE支持的规则选项。

- 路由规则 1：若请求流量满足以下所有条件，则将流量转发给服务`foo`的80端口处理。 `foo`实际由 `sub-foo1` 和 `sub-foo2`两个Service组成，分别接受80%和20%的流量。详见[多Service之间的负载均衡](load-balance.md)。

    - 域名为 `foo.com`
    - 路径前缀为 `/foo`
    - Cookie 中，Session 值为 `123`
    - Header 中，Content-Language 值为 `zh-cn`
  
- 路由规则 2：若请求流量满足以下所有条件，则将流量转发给服务`bar`的80端口处理
    - 域名为 `foo.com`
    - 路径前缀为 `/bar`
    - Cookie 中，Session 值为 `123`
    - Header 中，Content-Language 值为 `zh-cn`

## Ingress中路由的匹配条件

### 主机名条件(host)

由规则(rules)中的`host`字段指定

BFE Ingress Controller支持[Kubernetes原生定义的host匹配][hostname-wildcards]
                

### 路径条件(path)
由规则(rules)中的`path`和`pathType`字段指定

BFE Ingress Controller支持如下三种pathType：

- Prefix: 前缀匹配
- Exact: 精确匹配
- ImplementationSpecific: __默认__，BFE Ingress Controller实现为前缀匹配

### 高级匹配条件

BFE Ingress Controller支持以annotation的方式设置高级匹配条件。目前支持cookie和header两种高级匹配条件。

高级匹配条件在Ingress资源内共享，即同一个Ingress资源内的所有规则，都会受高级匹配条件的约束。

#### cookie

格式：
``` yaml
bfe.ingress.kubernetes.io/router.cookie: "key: value"
```

含义：

对于包含了名为key，值为value的cookie的请求，视为符合该cookie条件。

#### header      

格式：

``` yaml
bfe.ingress.kubernetes.io/router.header: "key: value"
```

含义：

对于包含了名为key，值为value的请求，视为符合该header条件

#### 限制

- 在同一个Ingress资源中，一个高级匹配条件类型仅支持设置一个值。
  
- 若设置了多个同一条件类型的`Annotation`，位置靠后的`Annotation`生效。
  
    ```yaml
    # 例
    annotation:
      bfe.ingress.kubernetes.io/router.header: "key1: value1" # 不生效
      bfe.ingress.kubernetes.io/router.header: "key2: value2" # 生效
    ```

## ingress class
BFE Ingress Controller支持两种方式设置ingress class。

### annotation方式

在Ingress的annotation中设置`kubernetes.io/ingress.class`，缺值为`bfe`
```yaml
  annotations:
    kubernetes.io/ingress.class: bfe  
```

### IngressClass方式

k8s 1.19+

```yaml
kind: IngressClass
apiVersion: networking.k8s.io/v1
metadata:
  name: external-lb
spec:
  controller: bfe-networks.com/ingress-controller
```

如果k8s版本为1.14~1.18，可以在K8S集群中配置IngressClass，指定controller为`bfe-networks.com/ingress-controller`:

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: IngressClass
metadata:
  name: external-lb
  controller: bfe-networks.com/ingress-controller
```

在Ingress的配置中使用名为`external-lb`的上述IngressClass:

```yaml
apiVersion: "networking.k8s.io/v1beta1"
kind: "Ingress"
metadata:
  name: "example-ingress"
spec:
  ingressClassName: "external-lb"
...
```

更多IngressClass相关信息，参见[IngressClass]。


[Ingress介绍]: https://kubernetes.io/docs/concepts/services-networking/ingress/#what-is-ingress
[pathType]: https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types
[hostname-wildcards]: https://kubernetes.io/docs/concepts/services-networking/ingress/#hostname-wildcards
[IngressClass]: https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/#extended-configuration-with-ingress-classes
