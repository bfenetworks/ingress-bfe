# Configuration Guide

## Introduction
Configure Ingress resources to define routes for accessing Services in Kubernetes cluster from outside the cluster. For more information about Ingress, please refer to [Ingress][] .

Refer to [ingress.yaml](../../examples/ingress.yaml) when configuring Ingress resources in yaml files.

## Example
### Simple example
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: simple-ingress
spec:
  rules:
  - host: whoami.com
    http:
      paths:
      - path: /testpath
        pathType: Prefix
        backend:
          service:
            name: whoami
            port:
              number: 80
```
Above example defines a Ingress resource, and

- sets `kubernetes.io/ingress.class` to `bfe`, means this Ingress will be handled by BFE Ingress Controller

- defines a simple route rule. A request will be forwarded to port 80 of Service `whoami`, if it matches both below conditions:
  - hostname is `whoami.com` 
  - path has prefix `/testpath`

### Complicated example
```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: complex-ingress
  namespace: my-namespace
  annotations:
    bfe.ingress.kubernetes.io/loadbalance: '{"foo": {"sub-foo1":80, "sub-foo2":20}}'
    bfe.ingress.kubernetes.io/router.cookie: "Session: 123"
    bfe.ingress.kubernetes.io/router.header: "Content-Language: zh-cn"
spec:
  tls:
  - hosts:
      - foo.com
    secretName: secret-foo-com
  rules:
  - host: foo.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: foo
            port:
              number: 80
      - path: /bar
        pathType: Exact
        backend:
          service:
            name: bar
            port:
              number: 80
```
Above Ingress resource defines 2 advanced route rules, and configure TLS certificate for `foo.com`. Rule options supported by BFE are defined with annotations.

- Route rule 1：a request will be forwarded to port 80 of service`foo` , if it matches all below conditions. Service `foo` is composed of two Services: `sub-foo1` and `sub-foo2`, serving 80% and 20% of total requests to `foo`. See [Load balancing between Services](load-balance.md).

    - hostname is `foo.com`
    - path has prefix `/foo`
    - value of a Cookie named `Session` is `123`
    - value of a Header named `Content-Language` is `zh-cn`
  
- Route rule 2：a request will be forwarded to port 80 of service`bar` , if it matches all below conditions. 
    - hostname is `foo.com`
    - path has prefix `/bar`
    - value of a cookie named `Session` is `123`
  - value of a header named `Content-Language` is `zh-cn`
  

## Condition of route rules

### Hostname condition(host)

Specified by `host` in a rule

BFE Ingress Controller support [hostname conditions][hostname-wildcards] defined by Kubernetes.                

### Path condition(path)
Specified by `path` and `pathType` in a rule

BFE Ingress Controller support below pathType：

- Prefix: prefix match.
- Exact: exact match
- ImplementationSpecific: __default__，implemented by BFE Ingress Controller as prefix match

### Advanced match condition

#### Introduction

BFE Ingress Controller supports advanced conditions by configuring `annotation`.

Advanced conditions are shared in an Ingress resource. 
So all the rules in the same Ingress resource will be restrained by advanced conditions, if configured.

Currently, BFE Ingress Controller supports two types of advanced condition: cookie and header.

#### Cookie

Format：
``` yaml
bfe.ingress.kubernetes.io/router.cookie: "key: value"
```

Explanation：

Requests containing a cookie with name=`key` and value=`value` are considered as matching this condition.

#### Header      

Format：

``` yaml
bfe.ingress.kubernetes.io/router.header: "key: value"
```

Explanation：

Requests containing a header with name=`key` and value=`value` are considered match this condition.

#### Restriction

- In an Ingress resource, for each advanced condition type, no more than one `Annotation` can be configured.
  
- If more than one `Annotation`s of the same advanced condition type are configured in the same Ingress resource, the last one takes effect.
  
    ```yaml
    # example
    annotation:
      bfe.ingress.kubernetes.io/router.header: "key1: value1" # not take effect
      bfe.ingress.kubernetes.io/router.header: "key2: value2" # takes effect
    ```

## Ingress class

BFE Ingress Controller supports user to configure ingress class in two ways:

### Set in annotations

Set `kubernetes.io/ingress.class` in annotations of Ingress. Default value is `bfe`

```yaml
  annotations:
    kubernetes.io/ingress.class: bfe  
```

### Set in IngressClass

the format of set IngressClass in k8s are varies from the versions of K8S.

#### set IngressClass
For k8s versions from 1.19
```yaml
apiVersion: networking.k8s.io/v1
metadata:
  name: external-lb
  labels:
    app.kubernetes.io/component: controller
  annotations:
    ingressclass.kubernetes.io/is-default-class: 'true'
spec:
  controller: bfe-networks.com/ingress-controller
```
For K8S versions from 1.14 to 1.18, set controller to `bfe-networks.com/ingress-controller` in IngressClass of K8S Cluster. Example:

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: IngressClass
metadata:
  name: external-lb
  controller: bfe-networks.com/ingress-controller
```

#### Then Ingress
Then set `ingressClassName` to `external-lb` in Ingress:

```yaml
apiVersion: "networking.k8s.io/v1beta1"
kind: "Ingress"
metadata:
  name: "example-ingress"
spec:
  ingressClassName: "external-lb"
...
```

For information about IngressClass, refer to [IngressClass]


[Ingress]: https://kubernetes.io/docs/concepts/services-networking/ingress/#what-is-ingress
[pathType]: https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types
[hostname-wildcards]: https://kubernetes.io/docs/concepts/services-networking/ingress/#hostname-wildcards
[IngressClass]: https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/#extended-configuration-with-ingress-classes
