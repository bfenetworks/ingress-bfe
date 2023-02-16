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

## Redirect

### Response Location

| Annotation Name | Function | Value |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/redirect.url-set][] | Redirect to specified URL | String. i.e. `https://www.baidu.com` |
| [bfe.ingress.kubernetes.io/redirect.url-from-query][] | Redirect to URL parsed from specified query in request | String. The key of the query. |
| [bfe.ingress.kubernetes.io/redirect.url-prefix-add][] | Redirect to URL concatenated by specified prefix and the original URL | String. i.e. `https://www.baidu.com?prefixPath` |
| [bfe.ingress.kubernetes.io/redirect.scheme-set][] | Redirect to the original URL but with scheme changed. supported scheme: http|https | String. i.e. `https` |

### Response Status Code

| Annotation Name | Function | Value |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/redirect.response-status][] | (Optional) Set the Status Code of the Redirect Response | Number String. Optional `301`、`302`、`303`、`307`、`308`,default is `302` |

## Rewrite URL

### Host

| Annotation Name                                              | Function                           | Value                                         |
| :----------------------------------------------------------- | :--------------------------------- | :-------------------------------------------- |
| [bfe.ingress.kubernetes.io/rewrite-url.host][]               | Set the host to specified value.   | JSON string. i.e. `[{"params": "baidu.com"}]` |
| [bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix][] | Set host to specified path prefix. | JSON string.  i.e. `[{"params": true}]`       |

### Path

| Annotation Name                                             | Function                                                     | Value                                        |
| :---------------------------------------------------------- | :----------------------------------------------------------- | :------------------------------------------- |
| [bfe.ingress.kubernetes.io/rewrite-url.path][]              | Set path to specified value.                                 | JSON string. i.e. `[{"params": "/foo/bar"}]` |
| [bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add][]   | Add prefix to original path.                                 | JSON string. i.e. `[{"params": "/bar"}]`     |
| [bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim][]  | Trim prefix from original path.                              | JSON string. i.e. `[{"params": "/bar"}]`     |
| [bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip][] | Strip the indicated number of prefix segments from original path. | JSON string. i.e.  `[{"params": 1}]`         |

### Query

| Annotation Name                                              | Function                                   | Value                                                |
| :----------------------------------------------------------- | :----------------------------------------- | :--------------------------------------------------- |
| [bfe.ingress.kubernetes.io/rewrite-url.query-add][]          | Add query.                                 | JSON string. i.e.  `[{"params": {"name": "alice"}}]` |
| [bfe.ingress.kubernetes.io/rewrite-url.query-delete][]       | Delete query.                              | JSON string. i.e.  `[{"params": ["name"]}]`          |
| [bfe.ingress.kubernetes.io/rewrite-url.query-rename][]       | Rename query.                              | JSON string. i.e. `[{"params": {"name": "user"} }]`  |
| [bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except][] | Delete all queries except specified query. | JSON string. i.e. `[{"params": "name"}]`             |

## BFE-Reserved

| Annotation Name | Function | Value |
|:---|:---|:---|
| [bfe.ingress.kubernetes.io/bfe-ingress-status][] | Feedback ingress status | `Read-only` JSON string, which contains ingress status, error message |

[kubernetes.io/ingress.class]: https://kubernetes.io/docs/concepts/services-networking/ingress/#deprecated-annotation

[bfe.ingress.kubernetes.io/bfe-ingress-status]: ../ingress/validate-state.md

[bfe.ingress.kubernetes.io/router.cookie]: ../ingress/basic.md#cookie

[bfe.ingress.kubernetes.io/router.header]: ../ingress/basic.md#header

[bfe.ingress.kubernetes.io/balance.weight]: ../ingress/load-balance.md

[bfe.ingress.kubernetes.io/redirect.url-set]: ../ingress/redirect.md#static-url

[bfe.ingress.kubernetes.io/redirect.url-from-query]:  ../ingress/redirect.md#

[bfe.ingress.kubernetes.io/redirect.url-prefix-add]: ../ingress/redirect.md#add-prefix

[bfe.ingress.kubernetes.io/redirect.scheme-set]: ../ingress/redirect.md#set-scheme

[bfe.ingress.kubernetes.io/redirect.response-status]: ../ingress/redirect.md#response-status-code
[bfe.ingress.kubernetes.io/rewrite-url.host]: ../ingress/rewrite.md#static-host
[bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix]: ../ingress/rewrite.md#dynamic-host
[bfe.ingress.kubernetes.io/rewrite-url.path]: ../ingress/rewrite.md#static-path
[bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add]: ../ingress/rewrite.md#add-path-prefix
[bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim]: ../ingress/rewrite.md#delete-path-prefix
[bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip]: ../ingress/rewrite.md#strip-path-prefix
[bfe.ingress.kubernetes.io/rewrite-url.query-add]: ../ingress/rewrite.md#add-query
[bfe.ingress.kubernetes.io/rewrite-url.query-delete]: ../ingress/rewrite.md#delete-query
[bfe.ingress.kubernetes.io/rewrite-url.query-rename]: ../ingress/rewrite.md#rename-query
[bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except]: ../ingress/rewrite.md#delete-all-queries-except
