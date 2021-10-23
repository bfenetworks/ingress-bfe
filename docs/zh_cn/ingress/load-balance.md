# 多Service之间的负载均衡
## 说明

BFE Ingress Controller支持在提供相同服务的多个Service（为便于理解，在BFE Ingress文档中称其为子服务，Sub-Service）之间按权重进行负载均衡。

配置方式

BFE Ingress Controller通过`注解`（`Annotation`）的方式支持多个Sub-Service之间的负载均衡。配置方式为：

- 在`annotations`中

  - 为多个Sub-Service分别指定流量分配权重

  - 为它们提供的服务设置一个Service名称

  - 格式如下：

    ``` yaml
    bfe.ingress.kubernetes.io/balance.weight: '{"service": {"sub-service1":80, "sub-service2":20}}'
    ```

- 在`rules`中

  - 将backend的serviceName设置为注解中设置的Service名称，并指定servicePort

## 示例

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: tls-example-ingress
  annotations:
    kubernetes.io/ingress.class: bfe    
    bfe.ingress.kubernetes.io/balance.weight: '{"service": {"service1":80, "service2":20}}'
spec:
  tls:
  - hosts:
      - https-example.foo.com
    secretName: testsecret-tls
  rules:
  - host: https-example.foo.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          serviceName: service
          servicePort: 80
```