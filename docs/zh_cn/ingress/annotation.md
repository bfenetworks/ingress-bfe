# Annotation

## 用途
BFE Ingress Annotation 用于支持高级配配规则。

目前支持`Cookie`和`Header`两种，格式和优先级如下:


## Cookie
- 优先级：0
``` yaml
bfe.ingress.kubernetes.io/router.cookie: "key: value"
```
BFE将执行 `req.Cookies["Key"]==value` 的判断




## Header
- 优先级：1
``` yaml
bfe.ingress.kubernetes.io/router.header: "key: value"
```
BFE将执行 `req.Headers["Key"]==value` 的判断



## 注意
- 一个类型的Annotation下仅支持设置一个值；
    ```yaml
    # 例
    annotation:
      bfe.ingress.kubernetes.io/router.header: "key1: value1" # 不生效
      bfe.ingress.kubernetes.io/router.header: "key2: value2" # 生效
    ```
- 优先级数组越小，其优先级越高