@sig-network @conformance @release-1.22
Feature: Load Balancing
  An Ingress exposing a backend service with multiple replicas should use all the pods available
  The feature sessionAffinity is not configured in the backend service https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#service-v1-core

  Background:
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: path-rules 
    spec:
      rules:
        - host: "load-balancing"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: echo-service
                    port:
                      number: 8080

    """
    Then The Ingress status shows the IP address or FQDN where it is exposed
    Then The backend deployment "echo-service" for the ingress resource is scaled to 10

  Scenario Outline: An Ingress should send all requests to the backend
    When I send 100 requests to "http://load-balancing"
    Then all the responses status-code must be 200 and the response body should contain the IP address of 10 different Kubernetes pods
