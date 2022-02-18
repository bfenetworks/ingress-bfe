# Annotation 开发指南

## 概述

在为 BFE Ingress Controller 开发 Annotation 时，需要考虑以下方面：
- Annotation 的定义
- Ingress 资源中 Annotation 含义的解析
- BFE 配置的生成
- BFE 配置的热加载

下面的讲述中，将结合 `bfe.ingress.kubernetes.io/balance.weight` 的实现作为例子。
> [balance.go][]

## 1. Annotation 的定义

### 格式
- Key:
  - BFE Ingress Controller 的 Annotation
    - `bfe.ingress.kubernetes.io/{module}.{key}`
    - `bfe.ingress.kubernetes.io/{key}`
  - k8s Ingress 约定的 Annotation
    - `kubernetes.io/{key}`
    - `ingressclass.kubernetes.io/{key}`
- Value: 根据需求设计定义

> 本例中:
> - Key: `bfe.ingress.kubernetes.io/balance.weight`
> - Value: 详见[负载均衡][]
>   - Demo: `{"service": {"service1":80, "service2":20}}`

### 已有的 Annotation
| Annotation 名  | 作用 |
| :--- | :---: |
| bfe.ingress.kubernetes.io/balance.weight | [负载均衡][] |
| bfe.ingress.kubernetes.io/router.cookie | 路由匹配条件：[匹配 Cookie](../ingress/basic.md#cookie) |
| bfe.ingress.kubernetes.io/router.header | 路由匹配条件：[匹配 Header](../ingress/basic.md#header) |
| bfe.ingress.kubernetes.io/bfe-ingress-status | [生效状态](../ingress/validate-state.md) |
| kubernetes.io/ingress.class | [申明 Ingress 类](https://kubernetes.io/zh/docs/concepts/services-networking/ingress/#deprecated-annotation) |
| ingressclass.kubernetes.io/is-default-class | [申明默认 Ingress 类](https://kubernetes.io/docs/concepts/services-networking/ingress/#default-ingress-class) |

### 注意事项
新 BFE Ingress Controller Annotation 的定义需兼容已有的 Annotation，在实现指定功能的基础上，尽量做到：
- 设计简洁
- 避免与已有 Annotation 功能重复

更多细节建议在 [Issue][] 中讨论


## 2. Annotation 的解析

### 触发时机
1. [Kubernetes controller-runtime][] 监听事件后，触发 Reconcile
2. Reconcile 在[回调函数][]中，触发 configBuilder 的更新
3. configBuilder 更新过程中，根据输入的 Ingress 资源 [解析][balance.go] 指定 Annotation，用于后续生成 BFE 配置

> 本例中:
> - 指定 Annotation： `bfe.ingress.kubernetes.io/balance.weight`

### 注意事项
- configBuilder 更新时，输入是当前 k8s 集群最新的 Ingress 资源

## 3. BFE 配置的生成

### 触发时机
1. [Kubernetes controller-runtime][] 监听事件后，触发 Reconcile
2. Reconcile 在[回调函数][]中，触发 configBuilder 的更新
3. configBuilder 更新过程中，根据输入的 Ingress 资源 [生成][clusterConfig.go] 多种 BFE 配置对象

> 本例中:
> - 更新配置对象：`configBuilder.clusterConf`

### 注意事项
- configBuilder 更新时，输入是当前 k8s 集群最新的 Ingress 资源
- Ingress 资源的新增、更新、删除需分别适当处理
- BFE 配置对象`configBuilder.*`可能存在缓存，注意对缓存内容的更新

## 4. BFE 配置的热加载

### 触发时机
根据配置时间间隔，定时触发

### 实现逻辑
对于多种 BFE 配置对象`configBuilder.*`，分别执行以下逻辑：
1. 将配置对象以指定格式 [持久化为文件][clusterConfig.go]，存放在 BFE 指定路径
2. 调用 BFE 进程 [reload 指令][clusterConfig.go]，完成相关配置的热加载

> 本例详见：[子集群负载均衡配置](https://www.bfe-networks.net/zh_cn/configuration/cluster_conf/gslb.data/)

### FAQ
- [如何查询特定 BFE 配置的格式和文件路径？](core-logic.md#BFE配置如何定义)

[Issue]: https://github.com/bfenetworks/ingress-bfe/labels/enhancement
[Kubernetes controller-runtime]: https://github.com/kubernetes-sigs/controller-runtime
[负载均衡]: ../ingress/load-balance.md
[回调函数]: ../../../internal/controllers/netv1/ingress_controller.go
[balance.go]: ../../../internal/bfeConfig/annotations/balance.go
[clusterConfig.go]: ../../../internal/bfeConfig/configs/clusterConfig.go