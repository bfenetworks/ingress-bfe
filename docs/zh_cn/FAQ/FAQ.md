# 常见问题
1. 问题：BFE Ingress Controller支持哪些启动参数，如何设置

支持的启动参数：

 选项 | 默认值 | 用途|
| --- | --- | --- |
| --namespace <br> -n | 空 | 设置需监听的ingress所在的namespace，多个namespace 之间用`,`分割。<br>默认监听所有的 namespace。  |
| --ingress-class| bfe | 指定需监听的Ingress的`kubernetes.io/ingress.class`值。<br>如不指定，BFE Ingress Controller将监听class设置为bfe的Ingress。 通常无需设置。 |
| --default-backend| 空 | 指定default-backend服务的名字，格式为`namespace/name`。<br>如指定default-backend，没有命中任何Ingress规则的请求，将被转发到default-backend。 |

设置方式：
在BFE Ingress Controller的部署文件controller.yaml中指定。例如：
```yaml
...
      containers:
        - name: bfe-ingress-controller
          image: bfenetworks/bfe-ingress-controller:latest
          args: ["-n", "ns1,ns2", "--default-backend", "test/whoami"]
...
```
