# 多Service之间负载均衡
BFE-Ingress通过`Ingress Annotation`的方式支持多个Service之间按权重进行负载均衡，配置如下：
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: tls-example-ingress
  annotations:
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