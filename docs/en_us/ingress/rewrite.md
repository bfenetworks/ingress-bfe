# Rewrite

The BFE engine provides multiple ways to [modify the url of HTTP request](https://www.bfe-networks.net/en_us/modules/mod_rewrite/mod_rewrite/), including modifying host, path and query.

The BFE Ingress Controller supports modifying the url of matched requests by annotations. 

## How to Config

* In the Ingress object,
  - `spec.rules` define the route rules
  - `metadata.annotations` define the rewrite actions for the matched requests

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

The BFE Ingress Controller supports configuring rewrite actions of "host" by annotations, including set static host or dynamic host.

### Static Host

Use `bfe.ingress.kubernetes.io/rewrite-url.host` to rewrite host to a specified value. 

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.host: '[{"params": "baidu.com", "when": "AfterLocation"}]'
```

Set `params` field to the expected host and the type of `params` field should be string. 

Corresponding scenario:

* Original url: http://host/path?query-key=value
* Modified url: http://www.baidu.com/path?query-key=value

### Dynamic Host

Use `bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix` to set host to specified path prefix, value of `params` field is fixed to `true`.

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

Set `params` field to the expected path and type of `params` field should be string.

Corresponding scenario:

- Original url: http://host/path?query-key=value
- Modified url: http://host/index?query-key=value

### Dynamic Path

Modify the prefix of url path, including add, delete and strip prefix. Be careful with the `order` value in the rewrite rule.

#### Add Path Prefix 

Use `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add` to add prefix to original path.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/foo/"}]'
```

Set `params` field to the added prefix and type of `params` field should be string.

Corresponding scenario:

- Original url: https://host/path?query-key=value
- Modified url:  https://host/foo/path?query-key=value

#### Delete Path Prefix

Use `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim` to trim prefix from original path.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/foo/"}]'
```

Set `params` field to the trimmed prefix and type of `params` field should be string.

Corresponding scenario:

* Original url: https://host/foo/path?query-key=value

- Modified url: https://host/path?query-key=value

#### Strip Path Prefix

Use `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip` to strip the indicated number of prefix segments from original path.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip: '[{"params": 1}]'
```

Set `params` field to a positive number and type of `params` field should be string of number.

Corresponding scenario:

* Original url: https://host/foo/path?query-key=value

- Modified url: https://host/path?query-key=value

#### Order of Rewrite Rule

You can define multiple annotations to modify path of url, but be careful about the order of rules.

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

The url of the request is https://host/bar/other-path?query-key=value.

For  `case1` :

1. Prefix  `/index`  will be added and the url will be rewrited to https://host/index/bar/other-path?query-key=value.
2. Because a new prefix added, the path prefix is not `/bar`, thus the prefix trim will not be performed. The url will still be https://host/index/bar/other-path?query-key=value.

 For  `case2`：

1. With the prefix trim rule configured with order=1, the path will be trimmed first. So the url will be rewrited to https://host/other-path?query-key=value.
2. Prefix `/index` will be added and the url will be rewrited to https://host/index/other-path?query-key=value.

## Rewrite Query

The BFE Ingresss Controller supports multiple ways to modify query params. be careful about the order of rewrite rules. The supported ways are：

* add specified query
* rename specified query
* delete specified query
* delete all queries except the specified query.

### Add Query

Use `bfe.ingress.kubernetes.io/rewrite-url.query-add` to add new query.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-add: '[{"params": {"b": "2"}}]'
```

Value of `params` field should be the added queries and the type of  `params` field should be dict.

Corresponding scenario：

- Original url: https://host/path?a=1
- Modified url:  https://host/path?a=1&b=2

### Delete Query

Use  `bfe.ingress.kubernetes.io/rewrite-url.query-delete`  to delete query params.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-delete: '[{"params": ["a"]}]'
```

Value of `params` field should be the keys of removed queries, and the type of  `params`  should be list.

Corresponding scenario：

- Original url: https://host/path?a=1
- Modified url:  https://host/path

### Rename Query

Use `bfe.ingress.kubernetes.io/rewrite-url.query-rename` to rename query.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-rename: '[{"params": {"a": "b"}}]'
```

Value of `params` field should be key-key mappings of renamed queries and the type of `params` should be dict.

Corresponding scenario：

- Original url: https://host/path?a=1
- Modified url: https://host/path?b=1

### Delete All Queries Except

Use `bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except` to delete all queries except specified query.

For example：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except: '[{"params": "a"}]'
```

Value of `params` field should be the key of the only reserved query and the type of `params` should be string.

Corresponding scenario：

- Original url: https://host/path?a=1&b=2&c=3
- Modified url: https://host/path?a=1