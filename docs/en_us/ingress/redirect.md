# Redirect

The BFE Ingress Controller supports redirecting traffic matched by the Ingress by using `metadata.annotations` of the Ingress object.

## How to Config

In the Ingress object,

- `spec.rules` defines the route rules;
- `metadata.annotations` defines the behavior of redirecting traffic matched by the Ingress.

Reference format:

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

## Redirect Location

The BFE Ingress Controller supports 4 ways to configure the redirect location, and only one of them can be set in an Ingress object.

### Static URL

Use `bfe.ingress.kubernetes.io/redirect.url-set` to config the static redirect location。

For example:

```yaml
bfe.ingress.kubernetes.io/redirect.target: "https://www.baidu.com"
```

Corresponding scenario:

- Request: http://host/path?query-key=value
- Response: https://www.baidu.com

### Fetch URL from Query

Redirect location is fetched from specific Query of request URL, query key is specified by `bfe.ingress.kubernetes.io/redirect.url-from-query`.

For example:

```yaml
bfe.ingress.kubernetes.io/redirect.url-from-query: url
```

Corresponding scenario:

- Request: https://host/path?url=https%3A%2F%2Fwww.baidu.com
- Response: https://www.baidu.com

### Add Prefix

Redirect location is a combination of given prefix and the `Path` of request URL, the prefix is set by `bfe.ingress.kubernetes.io/redirect.url-prefix-add`.

For example:

```yaml
bfe.ingress.kubernetes.io/redirect.url-prefix-add: "http://www.baidu.com/redirect"
```

Corresponding scenario:

- Request: https://host/path?query-key=value
- Response: http://www.baidu.com/redirect/path?query-key=value

### Set Scheme

Change the scheme of the request。Only HTTP and HTTPS are supported.

For example:

```yaml
bfe.ingress.kubernetes.io/redirect.scheme-set: http
```

Corresponding scenario:

- Request: https://host/path?query-key=value
- Response: http://host/path?query-key=value

## Response Status Code

By default, the status code of the redirect response is `302`. Users can manually specify the redirect status code by setting `bfe.ingress.kubernetes.io/redirect.response-status`.

For example:

```yaml
bfe.ingress.kubernetes.io/redirect.response-status: 301
```

The supported redirection status codes are: 301, 302, 303, 307, 308.