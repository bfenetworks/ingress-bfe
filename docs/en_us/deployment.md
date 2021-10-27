# Deployment Guide

## Install

* To deploy BFE Ingress Controller and configure related access control:

    ``` shell script
    kubectl apply -f controller.yaml
    ```
    - Config file example: [controller.yaml](../../examples/controller.yaml)
        - This config file uses [BFE Ingress Controller latest image on Docker Hub](https://hub.docker.com/r/bfenetworks/bfe-ingress-controller). If you want to use your customized version of the image, edit the config file to specify it.
        - Or you can run `make docker` in root folder of this project to create your own local image and use it.

* To config role-based access control:
    ``` shell script
    kubectl apply -f rbac.yaml
    ```

   - Config file example: [rbac.yaml](../../examples/rbac.yaml)
   - See detailed instructions in [Role-Based Access Control](rbac.md)

## Test

* Create a test service

  ``` shell script
  kubectl apply -f whoami.yaml
  ```


   test service config file example：[whoami](../../examples/whoami.yaml)

* Create ingress resource，configure route for the test service and verify

   ``` shell script
   kubectl apply -f ingress.yaml  
   ```
   
   - Refer to [ingress.yaml](../../examples/ingress.yaml) for basic Ingress configuration.
   
   - Refer to [Summary](SUMMARY.md) for more Ingress configuration options that BFE Ingress Controller support.
