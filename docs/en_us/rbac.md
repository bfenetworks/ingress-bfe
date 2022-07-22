# Role-Based Access Control (RBAC)

## Introduction

This document describes how to deploy BFE Ingress Controller in an environment with RBAC enabled.

Kubernetes use [Role-based access control](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), and define below objects：

- Define 'role', to set permissions for the role：
  - `ClusterRole` - to define permissions of a role which is cluster-wide
  - `Role` - to define permissions of a role which belongs to specific namespace

- Define 'role binding', to grant permissions defined in a role to a user or set of users:
  - `ClusterRoleBinding` , to grant permissions defined in `ClusterRole` to user
  - `RoleBinding` , to grant permissions defined in `Role` to user

To deploy a BFE Ingress Controller instance in an environment with RBAC enabled, use the `ServiceAccount` that bound to a `ClusterRole`, which has been granted with all permissions BFE Ingress Controller required.

## Minimum permission set

BFE Ingress Controller required at least below permissions：

- permissions defined for a ClusterRole：

  ```yaml
  services, endpoints, secrets, namespaces: get, list, watch
  ingresses, ingressclasses: get, list, watch, update
  ```

## Example

### Example config files

[controller.yaml](../../examples/controller.yaml)

[rbac.yaml](../../examples/rbac.yaml)

### Define and refer ServiceAccount

In [controller.yaml](../../examples/controller.yaml) ：

- define a `ServiceAccount` ,
  - name it as `bfe-ingress-controller`
- define a BFE Ingress Controller instance deployment
  - Instance deployed should be linked to ServiceAccount `bfe-ingress-controller`

### Define ClusterRole

In [rbac.yaml](../../examples/rbac.yaml) ：
- define a `ClusterRole`,
  - name it as `bfe-ingress-controller`
  - grant cluster-wide permissions below to it：

    ```yaml
    services, endpoints, secrets, namespaces: get, list, watch
    ingresses, ingressclasses: get, list, watch, update
    ```

### Bind ClusterRole

In [rbac.yaml](../../examples/rbac.yaml) ：

- define a `ClusterRoleBinding`,
  - bind ServiceAccount `bfe-ingress-controller` to ClusterRole `bfe-ingress-controller`

