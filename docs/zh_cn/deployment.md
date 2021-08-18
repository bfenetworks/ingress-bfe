# 快速开始

## 安装指南
* 部署 bfe-ingress-controller，以及相关权限配置。
    ``` shell script
    kubectl apply -f deployment.yaml
    kubectl apply -f rbac.yaml
    ```
    - bfe-ingress-controller部署可参考 [deployment.yaml](../../deploy/deployment.yaml) 文件：修改文件中的`${image_repo}`和`${version}`，使用正确的bfe-ingress-controller镜像信息。

    - 权限配置可参考 [rbac.yaml](../../deploy/rbac.yaml)
* 创建测试服务(例：whoami)

* 创建ingress资源
   ``` shell script
   kubectl apply -f ingress.yaml  
   ```
   简单的ingess配置可参考 [ingress.yaml](../../deploy/ingress.yaml)。
   
   更多的bfe-ingress-controller所支持的Ingress配置，可参考[配置文档](ingress/configuration.md)。

## 权限配置文件说明

* [RBAC 文件编写指南](rbac.md)
