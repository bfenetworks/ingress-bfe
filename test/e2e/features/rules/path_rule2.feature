@ingress.rule @release-1.22
Feature: Path test
  An Ingress may define routing rules based on the request path.

  If the HTTP request path matches one of the paths in the
  Ingress objects, the traffic is routed to its backend service.

  Background:
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: multi-path
    spec:
      rules:
        - host: "all-prefix"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: single-slash-type-prefix
                    port:
                      number: 3000
              - path: /aaa
                pathType: Prefix
                backend:
                  service:
                    name: single-slash-aaa-type-prefix
                    port:
                      number: 3000

        - host: "prefix-exact"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: single-slash-type-prefix-mix
                    port:
                      number: 3000
              - path: /aaa
                pathType: Exact
                backend:
                  service:
                    name: single-slash-aaa-type-exact-mix
                    port:
                      number: 3000

        - host: "exact-prefix"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: single-slash-type-exact-mix
                    port:
                      number: 3000
              - path: /aaa
                pathType: Prefix
                backend:
                  service:
                    name: single-slash-aaa-type-prefix-mix
                    port:
                      number: 3000

        - host: "all-exact"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: single-slash-type-exact
                    port:
                      number: 3000
              - path: /aaa
                pathType: Exact
                backend:
                  service:
                    name: single-slash-aaa-type-exact
                    port:
                      number: 3000

        - host: "multi-path-route"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: single-slash
                    port:
                      number: 3000
              - path: /aaa
                pathType: Prefix
                backend:
                  service:
                    name: single-slash-aaa
                    port:
                      number: 3000
              - path: /aaa/bbb
                pathType: Prefix
                backend:
                  service:
                    name: single-slash-aaa-bbb
                    port:
                      number: 3000

    """
    Then The Ingress status shows the IP address or FQDN where it is exposed

  Scenario: An Ingress with prefix path rules / and /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request / and route to single-slash-type-prefix)

    When I send a "GET" request to "http://all-prefix/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix" service

  Scenario: An Ingress with prefix path rules / and /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /a and route to single-slash-type-prefix)

    When I send a "GET" request to "http://all-prefix/a"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix" service

  Scenario: An Ingress with prefix path rules / and /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa and route to single-slash-aaa-type-prefix)

    When I send a "GET" request to "http://all-prefix/aaa"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix" service

  Scenario: An Ingress with prefix path rules / and /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa/ and route to single-slash-aaa-type-prefix)

    When I send a "GET" request to "http://all-prefix/aaa/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix" service

  Scenario: An Ingress with prefix path rules / and /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa/ and route to single-slash-type-prefix)

    When I send a "GET" request to "http://all-prefix/aaaxyz"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix" service

  Scenario: An Ingress with prefix path rules / and /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa/ccc and route to single-slash-aaa-type-prefix)

    When I send a "GET" request to "http://all-prefix/aaa/ccc"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix" service

  Scenario: An Ingress with prefix path rules / and /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa/ccc/a/b/c/d/e/f/g and route to single-slash-aaa-type-prefix)

    When I send a "GET" request to "http://all-prefix/aaa/ccc/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix" service

  Scenario: An Ingress with prefix path rules / and exact path rules /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request / and route to single-slash-type-prefix-mix)

    When I send a "GET" request to "http://prefix-exact/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix-mix" service

  Scenario: An Ingress with prefix path rules / and exact path rules /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /a and route to single-slash-type-prefix-mix)

    When I send a "GET" request to "http://prefix-exact/a"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix-mix" service

  Scenario: An Ingress with prefix path rules / and exact path rules /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa and route to single-slash-type-prefix)

    When I send a "GET" request to "http://prefix-exact/aaa"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-exact-mix" service

  Scenario: An Ingress with prefix path rules / and exact path rules /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa and route to single-slash-type-prefix-mix)

    When I send a "GET" request to "http://prefix-exact/aaa/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix-mix" service

  Scenario: An Ingress with prefix path rules / and exact path rules /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaaxyz and route to single-slash-type-prefix-mix)

    When I send a "GET" request to "http://prefix-exact/aaaxyz"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix-mix" service

  Scenario: An Ingress with prefix path rules / and exact path rules /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa/ccc and route to single-slash-type-prefix-mix)

    When I send a "GET" request to "http://prefix-exact/aaa/ccc"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix-mix" service

  Scenario: An Ingress with prefix path rules / and exact path rules /aaa should send traffic to the matching backend service
  (prefix / and /aaa matches request /aaa/ccc/a/b/c/d/e/f/g and route to single-slash-type-prefix-mix)

    When I send a "GET" request to "http://prefix-exact/aaa/ccc/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-prefix-mix" service

  Scenario: An Ingress with exact path rules / and prefix path rules /aaa should send traffic to the matching backend service
  (exact / and prefix /aaa matches request / and route to single-slash-type-exact-mix)

    When I send a "GET" request to "http://exact-prefix/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-exact-mix" service

  Scenario: An Ingress with exact path rules / and prefix path rules /aaa should send traffic to the matching backend service
  (exact / and prefix /aaa matches request /a and not route to backend)

    When I send a "GET" request to "http://exact-prefix/a"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules / and prefix path rules /aaa should send traffic to the matching backend service
  (exact / and prefix /aaa matches request /aaa and route to single-slash-aaa-type-prefix-mix)

    When I send a "GET" request to "http://exact-prefix/aaa"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix-mix" service

  Scenario: An Ingress with exact path rules / and prefix path rules /aaa should send traffic to the matching backend service
  (exact / and prefix /aaa matches request /aaa and route to single-slash-aaa-type-prefix-mix)

    When I send a "GET" request to "http://exact-prefix/aaa"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix-mix" service

  Scenario: An Ingress with exact path rules / and prefix path rules /aaa should send traffic to the matching backend service
  (exact / and prefix /aaa matches request /aaaxyz and not route to backend)

    When I send a "GET" request to "http://exact-prefix/aaaxyz"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules / and prefix path rules /aaa should send traffic to the matching backend service
  (exact / and prefix /aaa matches request /aaa/ccc and route to single-slash-aaa-type-prefix-mix)

    When I send a "GET" request to "http://exact-prefix/aaa/ccc"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix-mix" service

  Scenario: An Ingress with exact path rules / and prefix path rules /aaa should send traffic to the matching backend service
  (exact / and prefix /aaa matches request /aaa/ccc/a/b/c/d/e/f/g and route to single-slash-aaa-type-prefix-mix)

    When I send a "GET" request to "http://exact-prefix/aaa/ccc/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-prefix-mix" service

  Scenario: An Ingress with exact path rules / and path rules /aaa should send traffic to the matching backend service
  (exact / and exact /aaa matches request / and route to single-slash-type-exact)

    When I send a "GET" request to "http://all-exact/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-type-exact" service

  Scenario: An Ingress with exact path rules / and path rules /aaa should send traffic to the matching backend service
  (exact / and exact /aaa matches request /a and not route to backend)

    When I send a "GET" request to "http://all-exact/a"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules / and path rules /aaa should send traffic to the matching backend service
  (exact / and exact /aaa matches request /aaa and route to single-slash-aaa-type-exact)

    When I send a "GET" request to "http://all-exact/aaa"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-type-exact" service

  Scenario: An Ingress with exact path rules / and path rules /aaa should send traffic to the matching backend service
  (exact / and exact /aaa matches request /aaa/ and not route to backend)

    When I send a "GET" request to "http://all-exact/aaa/"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules / and path rules /aaa should send traffic to the matching backend service
  (exact / and exact /aaa matches request /aaaxyz and not route to backend)

    When I send a "GET" request to "http://all-exact/aaazyx"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules / and path rules /aaa should send traffic to the matching backend service
  (exact / and exact /aaa matches request /aaa/ccc and not route to backend)

    When I send a "GET" request to "http://all-exact/aaa/ccc"
    Then the response status-code must be 500

  Scenario: An Ingress with exact path rules / and path rules /aaa should send traffic to the matching backend service
  (exact / and exact /aaa matches request /aaa/ccc/a/b/c/d/e/f/g and not route to backend)

    When I send a "GET" request to "http://all-exact/aaa/ccc/a/b/c/d/e/f/g"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request / and route to single-slash)

    When I send a "GET" request to "http://multi-path-route/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /a and route to single-slash)

    When I send a "GET" request to "http://multi-path-route/a"
    Then the response status-code must be 200
    And the response must be served by the "single-slash" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /aaa and route to single-slash-aaa)

    When I send a "GET" request to "http://multi-path-route/aaa"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /aaa/ and route to single-slash-aaa)

    When I send a "GET" request to "http://multi-path-route/aaa/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /aaa/bb and route to single-slash-aaa)

    When I send a "GET" request to "http://multi-path-route/aaa/bb"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /aaa/bbb and route to single-slash-aaa-bbb)

    When I send a "GET" request to "http://multi-path-route/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-bbb" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /aaa/bbb/ and route to single-slash-aaa-bbb)

    When I send a "GET" request to "http://multi-path-route/aaa/bbb/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-bbb" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /aaa/bbb/ccc and route to single-slash-aaa-bbb)

    When I send a "GET" request to "http://multi-path-route/aaa/bbb/ccc"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa-bbb" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /ccc and route to single-slash)

    When I send a "GET" request to "http://multi-path-route/ccc"
    Then the response status-code must be 200
    And the response must be served by the "single-slash" service

  Scenario: An Ingress with path rules / and path rules /aaa and /aaa/bbb should send traffic to the matching backend service
  (prefix /, /aaa, /aaa/bbb  matches request /aaa/ccc and route to single-slash-aaa)

    When I send a "GET" request to "http://multi-path-route/aaa/ccc"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-aaa" service
