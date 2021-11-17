# Deployment Guide

## Installation
Install BFE Ingress Controller in either of two ways:
* Apply a configure file
* Install helm charts of controller

### Configure file

``` shell script
kubectl apply -f https://raw.githubusercontent.com/bfenetworks/ingress-bfe/develop/examples/controller-all.yaml
```

- Above configure file uses latest version of [BFE Ingress Controller  image](https://hub.docker.com/r/bfenetworks/bfe-ingress-controller) in Docker Hub. You can edit configure file to specify other version of the image.

- For details of permission configuration, please find more information in [Role-Based Access Control](rbac.md)

### Helm

```
helm upgrade --install bfe-ingress-controller bfe-ingress-controller --repo https://github.com/bfenetworks/ingress-bfe  --namespace ingress-bfe --create-namespace
```

- helm3 is required.

## Testing

* Create a testing service

  ``` shell script
  kubectl apply -f https://raw.githubusercontent.com/bfenetworks/ingress-bfe/develop/examples/whoami.yaml
  ```

* Create ingress resource for testing service to verify the installation

   ``` shell script
   kubectl apply -f https://raw.githubusercontent.com/bfenetworks/ingress-bfe/develop/examples/ingress.yaml  
   ```

