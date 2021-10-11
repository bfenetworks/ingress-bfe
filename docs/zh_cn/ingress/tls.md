# TLS 配置
BFE Ingress Controller按照Kubernetes原生定义的方式来管理TLS的证书和密钥。

TLS的证书和密钥通过Secrets进行保存，示例如下：

**Secret配置**

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
**Ingress配置**
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: tls-example-ingress
  annotations:
    kubernetes.io/ingress.class: bfe  
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
