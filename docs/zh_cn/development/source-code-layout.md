# BFE Ingress 源代码框架

```shell
.
├── .github       : Github 工作流目录
├── build         : 编译脚本目录
├── charts        : BFE Ingress Controller 的 Helm Charts目录
├── cmd
│   └── ingress-controller  : 主程序目录
├── docs          : 中英文用户文档及其素材目录
├── examples      : k8s 资源描述文件示例目录
├── internal      : 核心源码目录
│   ├── bfeConfig     : BFE Ingress Controller 配置选项定义代码
│   ├── controllers   : k8s 集群交互相关代码，主要包含各资源的controller的实现以及reconcile逻辑
│   └── option        : BFE 配置相关代码，主要包含各 BFE 配置的生成和热加载逻辑
├── scripts       : 镜像依赖的脚本目录
├── CHANGELOG.md  : 版本修改日志
├── Dockerfile    : 构建镜像的指令文件
├── LICENSE       : License 协议
├── Makefile      : 程序编译与镜像制作的指令文件
├── SECURITY.md   : 安全策略
├── VERSION       : 程序版本
├── go.mod        : Go 语言依赖管理文件
└── go.sum        : Go 语言依赖管理文件
```

## 核心代码
- /[internal][]: 核心源码目录
  - /[option][]: BFE Ingress Controller 配置选项定义代码
  - /[controllers][]: k8s 集群交互相关代码，主要包含各资源的controller的实现以及reconcile逻辑
  - /[bfeConfig][]: BFE 配置相关代码，主要包含各 BFE 配置的生成和热加载逻辑

## 持续集成

### 工作流
- /[.github][]: Github 工作流目录
- /[Makefile][]: 程序编译与镜像制作的指令文件

### 程序编译
- /cmd/[ingress-controller][]: 主程序目录
- /[VERSION][]: 程序版本
- /[build][]: 编译脚本目录
- /[go.mod][]: Go 语言依赖管理文件
- /[go.sum][]: Go 语言依赖管理文件

### 镜像制作
- /[Dockerfile][]: 构建镜像的指令文件
- /[scripts][]: 镜像依赖的脚本目录（如镜像启动脚本）

## Helm 支持
- /[charts][]: BFE Ingress Controller 的 Helm Charts

## 项目文档
- /[docs][]: 中英文用户文档及其素材目录
- /[examples][]: k8s 资源描述文件示例目录
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
[go.mod]: ../../../go.mod
[go.sum]: ../../../go.sum
[option]: ../../../internal/option
[scripts]: ../../../scripts
[CHANGELOG.md]: ../../../CHANGELOG.md
[Dockerfile]: ../../../Dockerfile
[LICENSE]: ../../../LICENSE
[Makefile]: ../../../Makefile
[SECURITY.md]: ../../../SECURITY.md
[VERSION]: ../../../VERSION