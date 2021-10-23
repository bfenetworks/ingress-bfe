# FAQ
1. Question：what arguments can be used to run BFE Ingress Controller, and how to define them?

   Answer: Arguments supported by BFE Ingress Controller：

|Argument | Default value | Description|
| --- | --- | --- |
| --namespace <br> -n | Null | Specifies in which namespaces the BFE Ingress Controller will monitor Ingress, seperate multiple namespaces by `,`. <br>Default value means monitor all namespaces  |
| --ingress-class| bfe | Specifies the `kubernetes.io/ingress.class` value of Ingress it monitors. <br>If not specified, BFE Ingress Controller monitors the Ingress with ingress class set as bfe. Usually you don't need to specify it. |
| --default-backend| Null | Specify name of default backend service, in the format of `namespace/name`.<br>If specified, requests that match no Ingress rule will be forwarded to the service specified. |

How to define：
Define in config file of BFE Ingress Controller, like [controller.yaml](../../../examples/controller.yaml). Example：

```yaml
...
      containers:
        - name: bfe-ingress-controller
          image: bfenetworks/bfe-ingress-controller:latest
          args: ["-n", "ns1,ns2", "--default-backend", "test/whoami"]
...
```
