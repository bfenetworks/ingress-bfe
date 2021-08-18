# Ingress 资源

## 什么是 Ingress 资源
Ingress 资源定义了 Kubernetes 集群内服务对外提供服务时的流量路由规则。
详见 [Ingress]

## 示例
### 简单示例
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: simple-ingress
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
上述 Ingress 资源定义了 1 条简单的路由规则：
若请求流量的域名为 `whoami.com`，路径前缀为 `/testpath`，
则将流量转发给`whoami` Service 的 80 端口处理

### 复杂示例
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: complex-ingress
  namespace: my-namespace
  annotations:
    bfe.ingress.kubernetes.io/loadbalance: '{"foo": {"foo1":80, "foo2":20}}'
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
上述 Ingress 资源定义了 2 条复杂的路由规则，并且为 `foo.com` 配置了证书
- 路由规则 1：若请求流量满足以下所有条件，则
将 80%流量转发给 `foo1` Service 的 80 端口处理，
将 20%流量转发给 `foo2` Service 的 80 端口处理
    - 域名为 `foo.com`
    - 路径前缀为 `/foo`
    - Cookie 中，Session 值为 `123`
    - Header 中，Content-Language 值为 `zh-cn`
- 路由规则 2：若请求流量满足以下所有条件，则
将流量转发给 `bar` Service 的 80 端口处理
    - 域名为 `foo.com`
    - 路径前缀为 `/bar`
    - Cookie 中，Session 值为 `123`
    - Header 中，Content-Language 值为 `zh-cn`
    
## 路由规则组成
- metadata
    - name: Ingress 资源名
    - namespace: 应用的命名空间
    - [annotations](annotation.md)
        - bfe.ingress.kubernetes.io/loadbalance: [多 Service 负载均衡](load-balance.md) 配置
        - [bfe.ingress.kubernetes.io/router.cookie](annotation.md#cookie): Cookie 匹配条件（同 Ingress 资源内共享）
        - [bfe.ingress.kubernetes.io/router.header](annotation.md#header): Header 匹配条件（同 Ingress 资源内共享）
- spec
    - [tls](tls.md)
        - host: 证书匹配域名
        - secretName: TLS 证书
    - rules
        - [host](#host): 域名匹配条件
        - http.paths
            - path & [pathType](#pathtype): 路径匹配条件
            - backend.service
                - name: 转发目标服务名
                - port: 转发目标服务端口
### host
BFE Ingress 支持[前缀匹配][hostname-wildcards]
                
### pathType
BFE Ingress 支持的 pathType 与 [Kubernetes 原生定义][pathType] 相近，具体为：
- Prefix: __默认__，前缀匹配
- Exact: 精确匹配
- ImplementationSpecific: 前缀匹配
        
 [Ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/#what-is-ingress
 [pathType]: https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types
 [hostname-wildcards]: https://kubernetes.io/docs/concepts/services-networking/ingress/#hostname-wildcards