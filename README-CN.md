# BFE Ingress Controller

中文 | [English](README.md)

## 简介

BFE Ingress Controller 为基于 [BFE][] 实现的[Kubernetes Ingress Controller][]，用于支持在 Kubernetes 中使用 [Ingress][] 进行流量接入，并从BFE的众多优秀特点和强大能力中受益。

## 特性和优势

- 基于Ingress的路由：支持基于Host、Path、Cookie、Header的Ingress路由规则
- 高级负载均衡：支持在提供相同服务的多个Service之间进行负载均衡
- 灵活的模块框架：采用灵活的模块框架设计，支持高效率定制开发扩展功能
- 配置热加载：支持配置热加载，配置更新和生效时无需重启BFE进程

## 开始使用

详见[部署指南](docs/zh_cn/deployment.md)

## 说明文档
详见[文档列表](docs/zh_cn/SUMMARY.md)

## 参与贡献
- 请首先在 [issue 列表](https://github.com/bfenetworks/ingress-bfe/issues) 中创建一个 issue
- 如有必要，请联系项目维护者/负责人进行进一步讨论
- 请遵循 [Golang 编程规范](https://github.com/golang/go/wiki/Style)

## 社区交流

- [用户论坛](https://github.com/bfenetworks/ingress-bfe/discussions)

- **开源BFE微信公众号**：扫码关注公众号“BFE开源项目”，及时获取项目最新信息和技术分享

  <table>
  <tr>
  <td><img src="./docs/images/qrcode_for_gh.jpg" width="100"></td>
  </tr>
  </table>

- **开源BFE用户微信群**：扫码加入，探讨和分享对BFE的建议、使用心得、疑问等

  <table>
  <tr>
  <td><img src="https://raw.githubusercontent.com/clarinette9/bfe-external-resource/main/wechatQRCode.png" width="100"></td>
  </tr>
  </table>

- **开源BFE开发者微信群**: [发送邮件](mailto:iyangsj@gmail.com)说明您的微信号及贡献(例如PR/Issue)，我们将及时邀请您加入

## 许可
基于 Apache 2.0 许可证，详见 [LICENSE](https://github.com/bfenetworks/ingress-bfe/blob/master/LICENSE) 文件说明

[Kubernetes Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/ "Kubernetes"
[Ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/ "Kubernetes"
[BFE]: https://github.com/bfenetworks/bfe "Github"
