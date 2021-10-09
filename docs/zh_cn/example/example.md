# 部署配置示例

## deployment
| 程序 | 文件  | 说明 |
| ---- | ---- | ---- |
| bfe ingress controller  | [controller.yaml](../../../examples/controller.yaml)| 用于 bfe ingress controller 的部署|
| 示例后端服务 whoami  | [whoami.yaml](../../../examples/whoami.yaml) | 用于示例服务(whoami)的部署 |

## ingress
| 文件  | 说明 |
| ---- | ---- |
| [ingress.yaml](../../../examples/ingress.yaml) | 用于配置示例服务(whoami)的流量调度 |

## rbac
| 文件  | 说明 |
| ---- | ---- |
| [rbac.yaml](../../../examples/rbac.yaml) | 用于授予 bfe ingress controller 的权限 |

