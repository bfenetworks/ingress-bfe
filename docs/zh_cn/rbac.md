# 基于角色的控制访问（RBAC）

## 说明

本文档说明如何在启用了RBAC的环境中部署BFE Ingress controller。

Kubernetes 中，采用[基于角色的访问控制](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)，使用了如下对象：

- 通过定义角色，配置角色相关的权限：
  - `ClusterRole` - 定义适用于整个集群的角色及其所具有的权限
  - `Role` - 定义适用于特定名称空间的角色及其所具有的权限

- 通过配置角色绑定，可以将角色中定义的权限赋予一个或者一组用户
  - `ClusterRoleBinding` ，赋予用户`ClusterRole`角色中定义的权限
  - `RoleBinding` ，赋予用户`ClusterRole`角色中定义的权限

在启用了RBAC的环境中部署BFE Ingress Controller，应该将其关联到一个`ServiceAccount`，且该`ServiceAccount`需要绑定到具有BFE Ingress Controller所需权限的`ClusterRole`。

## 最小权限集

BFE Ingress Controller所需要的权限至少应该包括：

- 具有ClusterRole中定义的如下权限：

  ```yaml
  services, endpoints, secrets, namespaces: get, list, watch
  ingresses, ingressclasses: get, list, watch, update
  ```

## 示例

### 示例配置文件controller

[controller.yaml](../../examples/controller.yaml)

[rbac.yaml](../../examples/rbac.yaml)

### 创建并引用ServiceAccount

在 [controller.yaml](../../examples/controller.yaml) 中：

- 定义一个 `ServiceAccount` ,
  - 命名为`bfe-ingress-controller`
- 定义了BFE Ingress Controller的部署
  - 部署的实例关联 ServiceAccount `bfe-ingress-controller`

### 创建ClusterRole

在 [rbac.yaml](../../examples/rbac.yaml) 中：
- 定义了一个`ClusterRole`,
  - 命名为`bfe-ingress-controller`
  - 定义了它具有如下的集群权限(适用于整个集群)：

    ```yaml
    services, endpoints, secrets, namespaces: get, list, watch
    ingresses, ingressclasses: get, list, watch, update
    ```

### 绑定ClusterRole

在 [rbac.yaml](../../examples/rbac.yaml) 中：

- 定义了一个`ClusterRoleBinding`,
  - 将 ServiceAccount `bfe-ingress-controller` 绑定到 ClusterRole `bfe-ingress-controller`

