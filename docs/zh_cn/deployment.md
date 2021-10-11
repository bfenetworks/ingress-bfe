# 快速开始

## 部署
* 安装bfe-ingress-controller
    ``` shell script
    kubectl apply -f controller.yaml
    ```
    > controller配置文件参考 [controller.yaml](../../examples/controller.yaml) 文件。可修改该文件中的`image`字段，使用期望的bfe-ingress-controller镜像版本。
    
* 配置bfe-ingress-controller所需权限
    ``` shell script
    kubectl apply -f rbac.yaml
    ```

    > 权限配置文件参考 [rbac.yaml](../../examples/rbac.yaml)。

## 测试
* 创建测试服务(例：whoami)
   ``` shell script
   kubectl apply -f whoami.yaml  
   ```

    > whoami服务配置参考 [whoami.yaml](../../examples/whoami.yaml)。

* 创建k8s ingress资源
   ``` shell script
   kubectl apply -f ingress.yaml  
   ```
   > 简单的ingess配置可参考 [ingress.yaml](../../examples/ingress.yaml)。更多的bfe-ingress-controller所支持的Ingress配置，可参考[配置文档](ingress/configuration.md)。

## 权限配置文件说明

* [RBAC 文件编写指南](rbac.md)
