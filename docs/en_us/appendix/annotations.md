# Annotation Indexes

## Specifying Ingress class

| Annotation Name | Function | Value |
|:---|:---|:---|
| [kubernetes.io/ingress.class][] | Specify Ingress class | fixed `bfe` |

## Routing

| Annotation Name | Function | Value |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/router.cookie][] | Cookie condition (exact match) for all routers in current ingress resource | key-value pair separated by `:`. i.e. `key:value` |
| [bfe.ingress.kubernetes.io/router.header][] | Header condition (exact match) for all routers in current ingress resource | key-value pair separated by `:`. i.e. `Key:value` |

## Load Balancing

| Annotation Name | Function | Value |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/balance.weight][] | Configure load balancing between multiple services | JSON string, i.e. `{"svc": {"sub-svc1":80, "sub-svc2":20}}` |

## BFE-Reserved

| Annotation Name | Function | Value |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/bfe-ingress-status][] | Feedback ingress status | `Read-only` JSON string, which contains ingress status, error message |

[kubernetes.io/ingress.class]: https://kubernetes.io/docs/concepts/services-networking/ingress/#deprecated-annotation

[bfe.ingress.kubernetes.io/bfe-ingress-status]: ../ingress/validate-state.md

[bfe.ingress.kubernetes.io/router.cookie]: ../ingress/basic.md#cookie

[bfe.ingress.kubernetes.io/router.header]: ../ingress/basic.md#header

[bfe.ingress.kubernetes.io/balance.weight]: ../ingress/load-balance.md