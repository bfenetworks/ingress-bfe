@ingress.rule @release-1.22
Feature: Path error rules
  An Ingress status error when given error ingress.

  Scenario: An Ingress with path rules a status should be error and not send traffic to the matching backend service
  (path rule a)
    Given an Ingress resource in a new random namespace should not create
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: path-err-rules
    spec:
      rules:
        - host: "exact-path-rules-cookie-uid"
          http:
            paths:
              - path: a
                pathType: Exact
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """

  Scenario: An Ingress with path rules null status should be error and not send traffic to the matching backend service
  (path rule null)
    Given an Ingress resource in a new random namespace should not create
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: path-err-rules
    spec:
      rules:
        - host: "exact-path-rules-cookie-uid"
          http:
            paths:
              - path:
                pathType: Exact
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
