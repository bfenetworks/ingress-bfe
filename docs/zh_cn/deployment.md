# 部署指南

## 安装
* 部署BFE Ingress Controller
    ``` shell script
    kubectl apply -f controller.yaml
    ```
    - 配置文件示例[controller.yaml](../../examples/controller.yaml)
        - 配置文件中使用了Docker Hub 上的[BFE Ingress Controller]:latest镜像。如需使用指定版本的镜像，修改配置文件，指定镜像版本。
        - 也可在项目根目录下执行`make docker`，创建自己的本地镜像。

* 配置所需权限
    ``` shell script
    kubectl apply -f rbac.yaml
    ```
    - 权限配置文件参考 [rbac.yaml](../../examples/rbac.yaml)。
    - 权限配置具体说明可参见[RBAC 文件编写指南](rbac.md)。    

## 测试
* 创建测试服务whoami
   ``` shell script
   kubectl apply -f whoami.yaml
   ```

    - whoami服务配置参考 [whoami.yaml](../../examples/whoami.yaml)。

* 创建k8s Ingress资源，验证消息路由
   ``` shell script
   kubectl apply -f ingress.yaml  
   ```

   - 基本的Ingess配置可参考 [ingress.yaml](../../examples/ingress.yaml)。
   - 更多的BFE Ingress Controller所支持的Ingress配置，可参考配置[相关文档](SUMMARY.md)。

[BFE Ingress Controller]: https://hub.docker.com/r/bfenetworks/bfe-ingress-controller
