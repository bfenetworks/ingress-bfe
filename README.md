# BFE Ingress Controller

English | [中文](README-CN.md)

## Overview

BFE Ingress Controller is an implementation of Kubernetes [Ingress Controller][] based on [BFE][], to fulfill [Ingress][] in Kubernetes.

## Features and Advantages

- Traffic routing based on Host, Path, Cookie and Header
- Support for load balancing among multiple Services of the same application
- Flexible plugin framework, based on which developers can add new features efficiently
- Configuration hot reload, avoiding impact on existing long connections

## Quick start

See [Deployment](docs/en_us/deployment.md) for quick start of using BFE Ingress Controller.

## Documentation
See [Document Summary](docs/en_us/SUMMARY.md).

## Contributing
- Create and issue in [Issue List](https://github.com/bfenetworks/ingress-bfe/issues)
- If necessary, contact and discuss with maintainer
- Follow the [Golang style guide](https://github.com/golang/go/wiki/Style)

## Communication

- [Forum](https://github.com/bfenetworks/ingress-bfe/discussions)
- BFE community on Slack: [Sign up](https://slack.cncf.io/) CNCF Slack and join bfe channel.
- BFE developer group on WeChat: [Send a request mail](mailto:iyangsj@gmail.com) with your WeChat ID and a contribution you've made to BFE(such as a PR/Issue). We will invite you right away.

## License

BFE is under the Apache 2.0 license. See the [LICENSE](https://github.com/bfenetworks/ingress-bfe/blob/master/LICENSE) file for details

[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/ "Kubernetes"
[Ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/ "Kubernetes"
[BFE]: https://github.com/bfenetworks/bfe "Github"
