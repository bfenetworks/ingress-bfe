@ingress.rule @release-1.22
Feature: host rule test
  An Ingress may define routing rules based on the request path and host.

  If the HTTP request path matches one of the paths in the
  Ingress objects, the traffic is routed to its backend service.

  Background:
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: diff-host-rules
    spec:
      rules:
        - host: "diff-host-diff-path-slash"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-diff-host-slash
                    port:
                      number: 8080

        - host: "diff-host-diff-path-slash-whoami"
          http:
            paths:
              - path: /whoami
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-diff-host-slash-whoami
                    port:
                      number: 8080
    """
    Then The Ingress status shows the IP address or FQDN where it is exposed

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path / and host diff-host-diff-path-slash matches request / and host diff-host-diff-path-slash)

    When I send a "GET" request to "http://diff-host-diff-path-slash/"
    Then the response status-code must be 200
    And the response must be served by the "service-diff-host-slash" service

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path / and host diff-host-diff-path-slash matches request /whoami and host diff-host-diff-path-slash)

    When I send a "GET" request to "http://diff-host-diff-path-slash/whoami"
    Then the response status-code must be 200
    And the response must be served by the "service-diff-host-slash" service

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path / and host diff-host-diff-path-slash matches request /whoami/a/b/c/d/e/f/g and host diff-host-diff-path-slash)

    When I send a "GET" request to "http://diff-host-diff-path-slash/whoami/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "service-diff-host-slash" service

  Scenario: An Ingress with path rules slash whoami and host should not send traffic to the matching backend service
  (path /whoami and diff-host-diff-path-slash-whoami dose not matches request / and host diff-host-diff-path-slash-whoami)

    When I send a "GET" request to "http://diff-host-diff-path-slash-whoami/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules slash whoami and host should send traffic to the matching backend service
  (path /whoami and diff-host-diff-path-slash-whoami matches request /whoami and host diff-host-diff-path-slash-whoami)

    When I send a "GET" request to "http://diff-host-diff-path-slash-whoami/whoami"
    Then the response status-code must be 200
    And the response must be served by the "service-diff-host-slash-whoami" service

  Scenario: An Ingress with path rules slash whoami and host should send traffic to the matching backend service
  (path /whoami and diff-host-diff-path-slash-whoami matches request /whoami/a/b/c/d/e/f/g and host diff-host-diff-path-slash-whoami)

    When I send a "GET" request to "http://diff-host-diff-path-slash-whoami/whoami/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "service-diff-host-slash-whoami" service
