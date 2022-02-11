@sig-network @conformance @release-1.22
Feature: Ingress class
  Ingresses can be implemented by different controllers, often with different configuration.
  Each Ingress definition could specify a class, a reference to an IngressClass resource that contains
  additional configuration including the name of the controller that should implement the class.
  
  https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class

  Scenario: An Ingress with an invalid ingress class should not send traffic to the matching backend service
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: test-ingress-class
    spec:
      ingressClassName: some-invalid-class-name
      rules:
        - host: "ingress-class"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: ingress-class-prefix
                    port:
                      number: 8080

    """
    Then The Ingress status should not contain the IP address or FQDN
