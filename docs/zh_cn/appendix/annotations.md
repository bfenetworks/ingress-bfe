# Annotations 索引

## 指定 Ingress 控制器

| Annotation名 | 作用 | 值 |
|:---|:---|:---|
| [kubernetes.io/ingress.class][] | 指定 Ingress 控制器 | 固定 `bfe` |

## 配置路由

| Annotation名 | 作用 | 值 |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/router.cookie][] | 当前 Ingress 的所有路由需精确匹配指定的 Cookie 条件 | `:`分隔的键值对。示例：`key:value` |
| [bfe.ingress.kubernetes.io/router.header][] | 当前 Ingress 的所有路由需精确匹配指定的 Header 条件 | `:`分隔的键值对。示例：`Key:value` |

## 配置负载均衡

| Annotation名 | 作用 | 值 |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/balance.weight][] | 配置多 Service 之间的负载均衡 | JSON 字符串。示例：`{"svc": {"sub-svc1":80, "sub-svc2":20}}` |

## 配置重定向

### Response Location相关

| Annotation名 | 作用 | 值 |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/redirect.url-set][] | 设置重定向Location为指定值 | 字符串。示例：`https://www.baidu.com` |
| [bfe.ingress.kubernetes.io/redirect.url-from-query][] | 设置重定向Location为指定请求Query值 | 字符串。需要取值的Query的key。 |
| [bfe.ingress.kubernetes.io/redirect.url-prefix-add][] | 设置重定向Location为原始URL增加指定前缀 | 字符串。示例：`https://www.baidu.com?prefixPath` |
| [bfe.ingress.kubernetes.io/redirect.scheme-set][] | 设置重定向Location为原始URL并修改协议(支持HTTP和HTTPS) | 字符串。示例：`https` |

### Response Status Code

| Annotation名 | 作用 | 值 |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/redirect.response-status][] | 设置重定向Response的状态码，该Annotation为可选项 | 数字形式的字符串。可选：`301`、`302`、`303`、`307`、`308`，默认为`302` |


## 系统保留

| Annotation名 | 作用 | 值 |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/bfe-ingress-status][] | 用于 BFE-Ingress 控制器反馈当前 Ingress 的生效情况 | 只读，不可设置。由 BFE-Ingress 控制器生成的 JSON 字符串，包含生效状态、错误原因|

[kubernetes.io/ingress.class]: https://kubernetes.io/zh-cn/docs/concepts/services-networking/ingress/#deprecated-annotation

[bfe.ingress.kubernetes.io/bfe-ingress-status]: ../ingress/validate-state.md

[bfe.ingress.kubernetes.io/router.cookie]: ../ingress/basic.md#cookie

[bfe.ingress.kubernetes.io/router.header]: ../ingress/basic.md#header

[bfe.ingress.kubernetes.io/balance.weight]: ../ingress/load-balance.md

[bfe.ingress.kubernetes.io/redirect.url-set]: ../ingress/redirect.md#静态URL

[bfe.ingress.kubernetes.io/redirect.url-from-query]:  ../ingress/redirect.md#从Query中获得URL

[bfe.ingress.kubernetes.io/redirect.url-prefix-add]: ../ingress/redirect.md#添加前缀

[bfe.ingress.kubernetes.io/redirect.scheme-set]: ../ingress/redirect.md#设置Scheme

[bfe.ingress.kubernetes.io/redirect.response-status]: ../ingress/redirect.md#重定向状态码