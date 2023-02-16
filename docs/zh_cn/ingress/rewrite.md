# URL重写配置

BFE引擎提供了丰富的[URL重写能力](https://www.bfe-networks.net/en_us/modules/mod_rewrite/mod_rewrite/)，包含对host、path、query等三部分url信息的修改操作。

BFE Ingress Controller可以根据Ingress配置中的相关注解（Annotations），对Ingress匹配的流量进行URL重写。

## 配置方式

Ingress资源内

* `spec.rules`定义路由规则
* `metadata.annotations`定义对符合路由规则的流量进行URL重写的行为

参考格式：

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

### URL重写param格式

| 配置项                | 描述                                                         |
| --------------------- | ------------------------------------------------------------ |
| callbackList[]        | 不同回调点的URL重写动作列表                                  |
| callbackList[].params | URL重写动作参数值                                            |
| callbackList[].order  | URL重写动作顺序，仅作用于同一回调点                          |
| callbackList[].when   | URL重写动作回调点，默认为`AfterLocation`（目前只支持该回调点） |

## 重写Host

BFE Ingress Controller支持用户通过annotation配置Host重写，可选择静态Host或动态Host。

### 静态Host

设置`bfe.ingress.kubernetes.io/rewrite-url.host`，配置静态Host。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.host: '[{"params": "baidu.com", "when": "AfterLocation"}]'
```

`params`字段的值为指定的host，类型需为字符串。

对应实例：

* 重写前: http://host/path?query-key=value
* 重写后: http://www.baidu.com/path?query-key=value

### 动态Host

设置`bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix`，从URL 路径前缀中设置动态Host。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix: '[{"params": "true"}]'
```

 `params`字段的值仅能设置为`true`。

对应示例：

- 重写前: https://old-host/new-host/path?query-key=value
- 重写后: https://new-host/path?query-key=value

## 重写Path

通过特定的annotation可配置Path重写，可设置为以下的模式其中之一：

* 静态Path

* 动态Path，包含Path前缀的添加、删除与剥离。

### 静态Path

设置`bfe.ingress.kubernetes.io/rewrite-url.path`，将Path设置为指定值。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path: '[{"params": "/index"}]'
```

`params`字段的值为指定的路径，类型需为字符串。

对应示例

- 重写前: http://host/path?query-key=value
- 重写后: http://host/index?query-key=value

### 动态Path

支持租户对路由的path进行动态设置，如移除、增加、截断前缀等。使用时，需注意顺序字段参数设置。

#### 添加Path前缀

设置`bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add`，向Path添加指定前缀。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/foo/"}]'
```

`params`字段的值为需添加的路径前缀，类型需为字符串。

对应示例

- 重写前: https://host/path?query-key=value
- 重写后: https://host/foo/path?query-key=value

#### 删除Path前缀

设置`bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim`，从Path中删除指定的前缀。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/foo/"}]'
```

`params`字段的值为被移除的路径前缀，类型需为字符串。

对应示例

* 重写前: https://host/foo/path?query-key=value

- 重写后: https://host/path?query-key=value

#### 剥离Path前缀

设置`bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip`，从Path中剥离指定数量的前缀片段。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip: '[{"params": "1"}]'
```

`params`字段的值为需剥离的前缀片段数量，类型需为数字形式字符串。

对应示例

* 重写前: https://host/foo/path?query-key=value

- 重写后: https://host/path?query-key=value

#### 动态Path注解顺序

租户可在当前Ingress资源中定义多个动态Path设置注解，需注意注解顺序。

例如：

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

对于流量：https://host/bar/other-path?query-key=value

在`case1`的配置下：

1. 增加`/index`前缀，流量重写为：https://host/index/bar/other-path?query-key=value。
2. 前缀移除时，path已新增前缀，不匹配`/bar`前缀，所以不会移除`/bar`，流量仍为：https://host/index/bar/other-path?query-key=value。

在`case2`的配置下：

1. 前缀移除，匹配`/bar`前缀，流量重写为：https://host/other-path?query-key=value。
2. 增加`/index`前缀，流量重写为：https://host/index/other-path?query-key=value。

## 重写Query

BFE Ingresss Controller支持多种query修改，使用时需注意顺序字段设置，具体如下：

* 添加指定query
* 重命名指定query
* 删除指定query
* 仅保留指定query

### 新增Query

设置`bfe.ingress.kubernetes.io/rewrite-url.query-add`，添加指定query参数。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-add: '[{"params": {"b": "2"}}]'
```

`params`字段的值为需添加Query的键值对，类型为字典。

对应示例：

- 重写前: https://host/path?a=1
- 重写后: https://host/path?a=1&b=2

### 删除指定Query

设置`bfe.ingress.kubernetes.io/rewrite-url.query-delete`，删除指定query参数。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-delete: '[{"params": ["a"]}]'
```

`params`字段的值为需删除Query的键，类型为字符串数组。

对应示例：

- 重写前: https://host/path?a=1
- 重写后: https://host/path

### 重命名指定Query

设置`bfe.ingress.kubernetes.io/rewrite-url.query-rename`，重命名指定query参数。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-rename: '[{"params": {"a": "b"}}]'
```

`params`字段的值为需重命名Query的键-键映射关系，类型为字典。

对应示例：

- 重写前: https://host/path?a=1
- 重写后: https://host/path?b=1

### 仅保留指定Query

设置`bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except`，重命名指定query参数。

例如：

```yaml
bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except: '[{"params": "a"}]'
```

`params`字段值为仅保留Query的键，类型为字符串。

对应示例：

- 重写前: https://host/path?a=1&b=2&c=3
- 重写后: https://host/path?a=1
