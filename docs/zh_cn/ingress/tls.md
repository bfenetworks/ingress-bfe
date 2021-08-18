# TLS 配置
原生Ingress的TLS中，证书和密钥是通过Secrets进行保存，例子如下：
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: testsecret-tls
  namespace: default
data:
  tls.crt: base64 encoded cert
  tls.key: base64 encoded key
type: kubernetes.io/tls
```
Ingress配置
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: tls-example-ingress
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
          serviceName: service1
          servicePort: 80
```
BFE-Ingress按照同样的方式来管理TLS的证书和密钥，其余更高级的TLS功能，包括一些TLS的配置，密钥加密等功能，需要参考BFE-Ingress CRD来实现；
