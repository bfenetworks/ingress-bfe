# 基于角色的控制访问（RBAC）

## 总览

此示例适用于在启用了RBAC的环境中部署的 bfe-ingres-controller

基于角色的访问控制由四层组成:

1. `ClusterRole` - 分配给适用于整个集群的角色的权限
2. `ClusterRoleBinding` - 将ClusterRole绑定到特定帐户
3. `Role` - 分配给适用于特定名称空间的角色的权限
4. `RoleBinding` - 将角色绑定到特定帐户

为了将RBAC应用于`bfe-ingres-controller`，应将该控制器分配给`ServiceAccount`。
该`ServiceAccount`应该绑定到为`bfe-ingres-controller`定义的`Roles`和`ClusterRoles`

## 示例说明

### 创建 ServiceAccount

在此 [示例](../../deploy/deployment.yaml) 中，创建了一个 ServiceAccount ，即`bfe-ingres-controller`。

### 创建权限集

在此 [示例](../../deploy/rbac.yaml) 中定义了 1 组权限:
- 由名为`bfe-ingres-controller`的`ClusterRole`定义的集群范围权限，

#### 集群权限

授予这些权限是为了使 bfe-ingres-controller 能够充当跨集群的入口。
这些权限被授予名为`bfe-ingres-controller`的 ClusterRole

- `services`, `endpoints`, `secrets`, `namespaces`: get, list, watch
- `ingresses`, `ingressclasses`: get, list, watch, update

如果在启动bfe-ingres-controller时覆盖了两个参数，请进行相应调整

### 权限绑定

在此 [示例](../../deploy/rbac.yaml) 中，ServiceAccount `bfe-ingres-controller` 绑定到 ClusterRole `bfe-ingres-controller`。

!!! 注意：[deployment](../../deploy/deployment.yaml) 中
- 容器关联的 serviceAccountName 必须与 serviceAccount 匹配。
- metadata，容器参数 和 POD_NAMESPACE 中的 namespace 应位于对应 ingress namespace 中