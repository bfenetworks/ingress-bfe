# Rewrite

The BFE engine provides multiple ways to modify the url of HTTP request, including modifying host, path and query.

The BFE Ingress Controller supports parsing the related annotations and modifying the url of matched requests. 

## How to Config

* In the Ingress object,
  - `spec.rules` define the route rules
  - `metadata.annotations` define the rewrite actions for modifying the matched requests

Reference format:

```yaml
metadata:
  annotations:
    bfe.ingress.kubernetes.io/rewrite-url.host: '[{"params": "baidu.com", "when": "AfterLocation"}]'
    bfe.ingress.kubernetes.io/rewrite-url.path: '[{"params": "/bar"}]'
    bfe.ingress.kubernetes.io/rewrite-url.query-rename: >-
     [
       {
          "params": {"name": "user"}, 
          "when": "AfterLocation", 
          "order": 1
       }
     ]
spec:
  rules:
  - ...
```

### Description of Rewrite Rule

| Config Item           | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| callbackList[]        | A list of rewrite rule, each rule must have an unique callback point. |
| callbackList[].params | Params of rewrite rule.                                      |
| callbackList[].order  | The order of rewrite rule,  effective on the same callback point. |
| callbackList[].when   | The callback point of rewrite url, default is `AfterLocation` (only support this callback point) |

## Rewrite Host

The BFE Ingress Controller supports for configuring host information by annotations, including set static host or dynamic host.

### Static Host

Use `bfe.ingress.kubernetes.io/rewrite-url.host` to configure the rewrite action of setting host to specified value.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.host: '[{"params": "baidu.com", "when": "AfterLocation"}]'
```

Corresponding scenario:

* Original url: http://host/path?query-key=value
* Modified url: http://www.baidu.com/path?query-key=value

### Dynamic Host

Use `bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix` to set host to specified path prefix.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix: '[{"params": "true"}]'
```

Corresponding scenario:

- Original url: https://old-host/new-host/path?query-key=value
- Modified url: https://new-host/path?query-key=value

## Rewrite Path

The following are allowed ways to configure the rewrite actions of url path:  

* Static path

* Dynamic path, including add, delete and strip path's prefix.

### Static Path

Use `bfe.ingress.kubernetes.io/rewrite-url.path`  to set path to specified value.

For example:

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path: '[{"params": "/index"}]'
```

Corresponding scenario:

- Original url: http://host/path?query-key=value
- Modified url: http://host/index?query-key=value

### Dynamic Path

Modify the prefix of url path, including add, delete and strip prefix. Be careful about setting the `order` value in the rewrite rule.

#### Add Path Prefix 

Use `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add` to add prefix to original path.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/foo/"}]'
```

Corresponding scenario:

- Original url: https://host/path?query-key=value
- Modified url:  https://host/foo/path?query-key=value

#### Delete Path Prefix

Use `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim` to trim prefix from original path.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/foo/"}]'
```

Corresponding scenario:

* Original url: https://host/foo/path?query-key=value

- Modified url: https://host/path?query-key=value

#### Strip Path Prefix

Use `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip` to strip the segments of original path.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip: '[{"params": "1"}]'
```

Corresponding scenario:

* Original url: https://host/foo/path?query-key=value

- Modified url: https://host/path?query-key=value

#### Order of Rewrite Rule

You can define multiple annotations to modify traffic's path, but be careful about the order of rules.

For example:

```yaml
# case1
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/index/", "order": 1}]'
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/bar", "order": 2}]'
```

```yaml
# case2
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/index/", "order": 2}]'
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/bar", "order": 1}]'
```

The matched url is https://host/bar/other-path?query-key=value.

For config of `case1` :

1. Add `/index`  prefix, the url is modified to https://host/index/bar/other-path?query-key=value.
2. Because of adding a new path prefix, the path is not matched to `/bar`  prefix, so the engine won't trim prefix. the url still is https://host/index/bar/other-path?query-key=value.

 For config of `case2`：

1. Because of matching to `/bar`  prefix, trim the prefix at first, then the url is modified to https://host/other-path?query-key=value.
2. Add `/index`  prefix, the url is modified to https://host/index/other-path?query-key=value.

## Rewrite Query

The BFE Ingresss Controller supports multiple ways to modify query params. be careful about the order of rewrite rules. The supported ways are：

* add specified query
* rename specified query
* delete specified query
* delete all queries except the specified query.

### Add Query

Use `bfe.ingress.kubernetes.io/rewrite-url.query-add` to add new query, the type of  `params`  should be dict.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-add: '[{"params": {"b": "2"}}]'
```

Corresponding scenario：

- Original url: https://host/path?a=1
- Modified url:  https://host/path?a=1&b=2

### Delete Query

Use  `bfe.ingress.kubernetes.io/rewrite-url.query-delete`  to delete query params, the type of  `params`  should be list.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-delete: '[{"params": ["a"]}]'
```

Corresponding scenario：

- Original url: https://host/path?a=1
- Modified url:  https://host/path

### Rename Query

Use `bfe.ingress.kubernetes.io/rewrite-url.query-rename` to rename query, the type of `params` should be dict.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-add: '[{"params": {"a": "b"}}]'
```

Corresponding scenario：

- Original url: https://host/path?a=1
- Modified url: https://host/path?b=1

### Delete All Queries Except

Use `bfe.ingress.kubernetes.io/rewrite-url.query-rename` to delete all queries except specified query,  the type of `params` should be string.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-add: '[{"params": "a"}]'
```

Corresponding scenario：

- Original url: https://host/path?a=1&b=2&c=3
- Modified url: https://host/path?a=1