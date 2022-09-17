# 重定向配置

BFE Ingress Controller支持通过在声明在Ingress中使用注解（Annotations）的方式，对当前Ingress匹配的流量进行重定向。

## 配置方式

Ingress 资源内

- `spec.rules`定义路由规则
- `metadata.annotations`定义对符合路由规则的流量，重定向响应的行为

参考格式：

```yaml
metadata:
  annotations:
    bfe.ingress.kubernetes.io/redirect.url-set: "https://www.baidu.com"
spec:
  rules:
  - ...
```

```yaml
metadata:
  annotations:
    bfe.ingress.kubernetes.io/redirect.scheme-set: https
    bfe.ingress.kubernetes.io/redirect.status: 301
spec:
  rules:
  - ...
```

## 重定向Location

BFE Ingress Controller支持使用4种方式配置重定向目标URL，并且每个Ingress对象仅允许使用一种配置方式。

### 静态URL

通过设置 `bfe.ingress.kubernetes.io/redirect.url-set`，配置静态重定向目标URL。

例如：

```yaml
bfe.ingress.kubernetes.io/redirect.target: "https://www.baidu.com"
```

对应示例

- Request: http://host/path?query-key=value
- Response: https://www.baidu.com

### 从Query中获得URL

从请求URL的指定`Query`值中获取重定向目标URL。通过设置`bfe.ingress.kubernetes.io/redirect.url-from-query`指定Query名。

例如：

```yaml
bfe.ingress.kubernetes.io/redirect.url-from-query: url
```

对应示例

- Request: https://host/path?url=https%3A%2F%2Fwww.baidu.com
- Response: https://www.baidu.com

### 添加前缀

重定向目标URL由指定前缀和请求URL的`Path`拼接而成。通过`bfe.ingress.kubernetes.io/redirect.url-prefix-add`设置拼接的前缀字符串。

例如：

```yaml
bfe.ingress.kubernetes.io/redirect.url-prefix-add: "http://www.baidu.com/redirect"
```

对应示例

- Request: https://host/path?query-key=value
- Response: http://www.baidu.com/redirect/path?query-key=value

### 设置Scheme

修改请求的协议。目前仅支持HTTP和HTTPS。

例如：

```yaml
bfe.ingress.kubernetes.io/redirect.scheme-set: http
```

对应示例

- Request: https://host/path?query-key=value
- Response: http://host/path?query-key=value

## 重定向状态码

默认情况下，重定向Response的状态码为302。也可以通过设置 `bfe.ingress.kubernetes.io/redirect.response-status`，手动指定重定向的状态码。

例如：

```yaml
bfe.ingress.kubernetes.io/redirect.response-status: 301
```

目前支持的重定向状态码有：301、302、303、307、308。