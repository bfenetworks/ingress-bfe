# 部署指南

## 安装
可通过下述两种方式中任意一种进行安装：

* 通过配置文件安装
* 通过helm安装

### 配置文件安装

``` shell script
kubectl apply -f https://raw.githubusercontent.com/bfenetworks/ingress-bfe/develop/examples/controller-all.yaml
```

- 配置文件中使用了Docker Hub 上的[BFE Ingress Controller镜像](https://hub.docker.com/r/bfenetworks/bfe-ingress-controller)的最新版本。如需使用指定版本的镜像，修改配置文件，指定镜像版本。
- 权限配置具体说明可参见[RBAC 文件编写指南](rbac.md)。    

### Helm安装

```
helm upgrade --install bfe-ingress-controller bfe-ingress-controller --repo https://bfenetworks.github.io/ingress-bfe  --namespace ingress-bfe --create-namespace
```
- 要求helm3

## 测试
* 创建测试服务whoami
``` shell script
kubectl apply -f https://raw.githubusercontent.com/bfenetworks/ingress-bfe/develop/examples/whoami.yaml
```

* 创建k8s Ingress资源，验证消息路由
``` shell script
kubectl apply -f https://raw.githubusercontent.com/bfenetworks/ingress-bfe/develop/examples/ingress.yaml  
```


