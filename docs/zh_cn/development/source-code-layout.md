# BFE Ingress 源代码框架

## 核心代码
- /[internal][]
  - /[option][]: BFE Ingress Controller 配置选项定义
  - /[controllers][]: k8s 集群交互相关代码，主要用于获取 Ingress 资源
  - /[bfeConfig][]: BFE 配置相关代码，主要用于生成 BFE 配置并使之生效

## 持续集成

### 工作流
- /[.github][]: Github 工作流
- /[Makefile][]: 程序编译与镜像制作的指令文件

### 程序编译
- /cmd/[ingress-controller][]: 主程序
- /[VERSION][]: 程序版本
- /[build][]: 编译脚本

### 镜像制作
- /[Dockerfile][]: 构建镜像的指令文件
- /[scripts][]: 镜像依赖的脚本（如镜像启动脚本）

## Helm 支持
- /[charts][]: BFE Ingress Controller 的 Helm Charts

## 项目文档
- /[docs][]: 中英文用户文档及其素材
- /[examples][]: k8s 资源描述文件示例
- /[CHANGELOG.md][]: 版本修改日志
- /[LICENSE][]: License 协议
- /[SECURITY.md][]: 安全策略

[.github]: ../../../.github
[build]: ../../../build
[charts]: ../../../charts
[ingress-controller]: ../../../cmd/ingress-controller
[docs]: ../../../docs
[examples]: ../../../examples
[internal]: ../../../internal
[bfeConfig]: ../../../internal/bfeConfig
[controllers]: ../../../internal/controllers
[option]: ../../../internal/option
[scripts]: ../../../scripts
[CHANGELOG.md]: ../../../CHANGELOG.md
[Dockerfile]: ../../../Dockerfile
[LICENSE]: ../../../LICENSE
[Makefile]: ../../../Makefile
[SECURITY.md]: ../../../SECURITY.md
[VERSION]: ../../../VERSION