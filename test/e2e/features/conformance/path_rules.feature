@sig-network @conformance @release-1.22
Feature: Path rules
  An Ingress may define routing rules based on the request path.
  
  If the HTTP request path matches one of the paths in the
  Ingress objects, the traffic is routed to its backend service.

  Background:
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: path-rules  
    spec:
      rules:
        - host: "exact-path-rules"
          http:
            paths:
              - path: /foo
                pathType: Exact
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 8080
    
        - host: "prefix-path-rules"
          http:
            paths:
              - path: /foo
                pathType: Prefix
                backend:
                  service:
                    name: foo-prefix
                    port:
                      number: 8080
    
              - path: /aaa/bbb
                pathType: Prefix
                backend:
                  service:
                    name: aaa-slash-bbb-prefix
                    port:
                      number: 8080
    
              - path: /aaa
                pathType: Prefix
                backend:
                  service:
                    name: aaa-prefix
                    port:
                      number: 8080
    
        - host: "mixed-path-rules"
          http:
            paths:
              - path: /foo
                pathType: Prefix
                backend:
                  service:
                    name: foo-prefix
                    port:
                      number: 8080
    
              - path: /foo
                pathType: Exact
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 8080

    
        - host: "trailing-slash-path-rules"
          http:
            paths:
              - path: /aaa/bbb/
                pathType: Prefix
                backend:
                  service:
                    name: aaa-slash-bbb-slash-prefix
                    port:
                      number: 8080
              - path: /foo/
                pathType: Exact
                backend:
                  service:
                    name: foo-slash-exact
                    port:
                      number: 8080

    """
    Then The Ingress status shows the IP address or FQDN where it is exposed

  Scenario: An Ingress with exact path rules should send traffic to the matching backend service
    (exact /foo matches request /foo)

    When I send a "GET" request to "http://exact-path-rules/foo"
    Then the response status-code must be 200
    And the response must be served by the "foo-exact" service

  Scenario: An Ingress with exact path rules should not match requests with trailing slash
    (exact /foo does not match request /foo/)

    When I send a "GET" request to "http://exact-path-rules/foo/"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules should be case sensitive
    (exact /foo does not match request /FOO)

    When I send a "GET" request to "http://exact-path-rules/FOO"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules should not match any other label
    (exact /foo does not match request /bar)

    When I send a "GET" request to "http://exact-path-rules/bar"
    Then the response status-code must be 500

  Scenario: An Ingress with prefix path rules should send traffic to the matching backend service
    (prefix /foo matches request /foo)

    When I send a "GET" request to "http://prefix-path-rules/foo"
    Then the response status-code must be 200
    And the response must be served by the "foo-prefix" service

  Scenario: An Ingress with prefix path rules should ignore the request trailing slash and send traffic to the matching backend service
    (prefix /foo matches request /foo/)

    When I send a "GET" request to "http://prefix-path-rules/foo/"
    Then the response status-code must be 200
    And the response must be served by the "foo-prefix" service

  Scenario: An Ingress with prefix path rules should be case sensitive
    (prefix /foo does not match request /FOO)

    When I send a "GET" request to "http://prefix-path-rules/FOO"
    Then the response status-code must be 500

  Scenario: An Ingress with prefix path rules should match multiple labels, match the longest path, and send traffic to the matching backend service
    (prefix /aaa/bbb matches request /aaa/bbb)

    When I send a "GET" request to "http://prefix-path-rules/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "aaa-slash-bbb-prefix" service

  Scenario: An Ingress with prefix path rules should match multiple labels, match the longest path, and subpaths and send traffic to the matching backend service
    (prefix /aaa/bbb matches request /aaa/bbb/ccc)

    When I send a "GET" request to "http://prefix-path-rules/aaa/bbb/ccc"
    Then the response status-code must be 200
    And the response must be served by the "aaa-slash-bbb-prefix" service

  Scenario: An Ingress with prefix path rules should and send traffic to the matching backend service
    (prefix /aaa matches request /aaa/ccc)

    When I send a "GET" request to "http://prefix-path-rules/aaa/ccc"
    Then the response status-code must be 200
    And the response must be served by the "aaa-prefix" service

  Scenario: An Ingress with prefix path rules should match each labels string prefix
    (prefix /aaa does not match request /aaaccc)

    When I send a "GET" request to "http://prefix-path-rules/aaaccc"
    Then the response status-code must be 500

  Scenario: An Ingress with prefix path rules should ignore the request trailing slash and send traffic to the matching backend service
    (prefix /foo matches request /foo/)

    When I send a "GET" request to "http://prefix-path-rules/foo/"
    Then the response status-code must be 200
    And the response must be served by the "foo-prefix" service

  Scenario: An Ingress with mixed path rules should send traffic to the matching backend service where Exact is preferred
    (exact /foo matches request /foo)

    When I send a "GET" request to "http://mixed-path-rules/foo"
    Then the response status-code must be 200
    And the response must be served by the "foo-exact" service

  Scenario: An Ingress with a trailing slashes in a prefix path rule should ignore the trailing slash and send traffic to the matching backend service
    (prefix /aaa/bbb/ matches request /aaa/bbb)

    When I send a "GET" request to "http://trailing-slash-path-rules/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "aaa-slash-bbb-slash-prefix" service

  Scenario: An Ingress with a trailing slashes in a prefix path rule should ignore the trailing slash and send traffic to the matching backend service
    (prefix /aaa/bbb/ matches request /aaa/bbb/)

    When I send a "GET" request to "http://trailing-slash-path-rules/aaa/bbb/"
    Then the response status-code must be 200
    And the response must be served by the "aaa-slash-bbb-slash-prefix" service

  Scenario: An Ingress with a trailing slashes in an exact path rule should not match requests without a trailing slash
    (exact /foo/ does not match request /foo)

    When I send a "GET" request to "http://trailing-slash-path-rules/foo"
    Then the response status-code must be 500
