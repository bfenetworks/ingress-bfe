@ingress.rule @release-1.22
Feature: Same host in multiple ingresses
  An Ingress may define routing rules based on the request path and host.

  If the HTTP request path matches one of the paths in the
  Ingress objects, the traffic is routed to its backend service.

  Background:
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-diff-path-1
    spec:
      rules:
        - host: "same-host"
          http:
            paths:
              - path: /test/foo
                pathType: Prefix
                backend:
                  service:
                    name: service-same-host-1
                    port:
                      number: 8080
    """
    Then The Ingress status shows the IP address or FQDN where it is exposed

    And an Ingress resource
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-diff-path-2
    spec:
      rules:
        - host: "same-host"
          http:
            paths:
              - path: /test
                pathType: Prefix
                backend:
                  service:
                    name: service-same-host-2
                    port:
                      number: 8080
    """
    Then The Ingress status shows the IP address or FQDN where it is exposed

  Scenario: An Ingress with path rules slash and host should not send traffic to the matching backend service
  (path /test and host same-host dose not matches request / and host same-host)

    When I send a "GET" request to "http://same-host/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path /test and host same-host matches request /test and host same-host)

    When I send a "GET" request to "http://same-host/test"
    Then the response status-code must be 200
    And the response must be served by the "service-same-host-2" service

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path /test and host same-host matches request /test/ and host same-host)

    When I send a "GET" request to "http://same-host/test/"
    Then the response status-code must be 200
    And the response must be served by the "service-same-host-2" service

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path /test and host same-host matches request /test/fo and host same-host)

    When I send a "GET" request to "http://same-host/test/fo"
    Then the response status-code must be 200
    And the response must be served by the "service-same-host-2" service

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path /test/foo and host same-host matches request /test/foo and host same-host)

    When I send a "GET" request to "http://same-host/test/foo"
    Then the response status-code must be 200
    And the response must be served by the "service-same-host-1" service

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path /test/foo and host same-host matches request /test/foo/ and host same-host)

    When I send a "GET" request to "http://same-host/test/foo/"
    Then the response status-code must be 200
    And the response must be served by the "service-same-host-1" service

  Scenario: An Ingress with path rules slash and host should send traffic to the matching backend service
  (path /test/foo and host same-host matches request /test/foo/a/b/c/e/f/g and host same-host)

    When I send a "GET" request to "http://same-host/test/foo/a/b/c/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "service-same-host-1" service
