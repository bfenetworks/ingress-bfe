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
      name: path-rules
    spec:
      rules:
        - host: "path-rules-single-slash"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: single-slash
                    port:
                      number: 3000

        - host: "path-rules-single-slash-exact"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: single-slash-exact
                    port:
                      number: 3000

        - host: "path-rules-single-slash-prefix"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: single-slash-prefix
                    port:
                      number: 3000

        - host: "path-rules-slash-foo"
          http:
            paths:
              - path: /foo
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: slash-foo
                    port:
                      number: 3000

        - host: "path-rules-slash-foo-exact"
          http:
            paths:
              - path: /foo
                pathType: Exact
                backend:
                  service:
                    name: slash-foo-exact
                    port:
                      number: 3000

        - host: "path-rules-slash-foo-prefix"
          http:
            paths:
              - path: /foo
                pathType: Prefix
                backend:
                  service:
                    name: slash-foo-prefix
                    port:
                      number: 3000

        - host: "path-rules-slash-foo-slash"
          http:
            paths:
              - path: /foo/
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: slash-foo-slash
                    port:
                      number: 3000

        - host: "path-rules-slash-foo-slash-exact"
          http:
            paths:
              - path: /foo/
                pathType: Exact
                backend:
                  service:
                    name: slash-foo-slash-exact
                    port:
                      number: 3000

        - host: "path-rules-slash-foo-slash-prefix"
          http:
            paths:
              - path: /foo/
                pathType: Prefix
                backend:
                  service:
                    name: slash-foo-slash-prefix
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bb"
          http:
            paths:
              - path: /aaa/bb
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: slash-aaa-slash-bb
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bb-exact"
          http:
            paths:
              - path: /aaa/bb
                pathType: Exact
                backend:
                  service:
                    name: slash-aaa-slash-bb-exact
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bb-prefix"
          http:
            paths:
              - path: /aaa/bb
                pathType: Prefix
                backend:
                  service:
                    name: slash-aaa-slash-bb-prefix
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bbb"
          http:
            paths:
              - path: /aaa/bbb
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: slash-aaa-slash-bbb
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bbb-exact"
          http:
            paths:
              - path: /aaa/bbb
                pathType: Exact
                backend:
                  service:
                    name: slash-aaa-slash-bbb-exact
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bbb-prefix"
          http:
            paths:
              - path: /aaa/bbb
                pathType: Prefix
                backend:
                  service:
                    name: slash-aaa-slash-bbb-prefix
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bbb-slash"
          http:
            paths:
              - path: /aaa/bbb/
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: slash-aaa-slash-bbb-slash
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bbb-slash-exact"
          http:
            paths:
              - path: /aaa/bbb/
                pathType: Exact
                backend:
                  service:
                    name: slash-aaa-slash-bbb-slash-exact
                    port:
                      number: 3000

        - host: "path-rules-slash-aaa-slash-bbb-slash-prefix"
          http:
            paths:
              - path: /aaa/bbb/
                pathType: Prefix
                backend:
                  service:
                    name: slash-aaa-slash-bbb-slash-prefix
                    port:
                      number: 3000
    """
    Then The Ingress status shows the IP address or FQDN where it is exposed

  Scenario: An Ingress with default path rules / should send traffic to the matching backend service
  (default / matches request /)

    When I send a "GET" request to "http://path-rules-single-slash/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash" service

  Scenario: An Ingress with default path rules / should send traffic to the matching backend service
  (default / matches request /a)

    When I send a "GET" request to "http://path-rules-single-slash/a"
    Then the response status-code must be 200
    And the response must be served by the "single-slash" service

  Scenario: An Ingress with default path rules / should send traffic to the matching backend service
  (default / matches request /a/b/c/d/e/f/g)

    When I send a "GET" request to "http://path-rules-single-slash/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "single-slash" service

  Scenario: An Ingress with Exact path rules / should send traffic to the matching backend service
  (Exact / matches request /)

    When I send a "GET" request to "http://path-rules-single-slash-exact/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-exact" service

  Scenario: An Ingress with Exact path rules / should not send traffic to the matching backend service
  (Exact / dose not matches request /a)
    When I send a "GET" request to "http://path-rules-single-slash-exact/a"
    Then the response status-code must be 500

  Scenario: An Ingress with Exact path rules / should not send traffic to the matching backend service
  (Exact / dose not matches request /a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-single-slash-exact/a/b/c/d/e/f/g"
    Then the response status-code must be 500

  Scenario: An Ingress with Exact path rules / should send traffic to the matching backend service
  (Prefix / matches request /)

    When I send a "GET" request to "http://path-rules-single-slash-prefix/"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-prefix" service

  Scenario: An Ingress with Exact path rules / should send traffic to the matching backend service
  (Prefix / dose not matches request /a)
    When I send a "GET" request to "http://path-rules-single-slash-prefix/a"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-prefix" service

  Scenario: An Ingress with Exact path rules / should send traffic to the matching backend service
  (Prefix / dose not matches request /a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-single-slash-prefix/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "single-slash-prefix" service

  Scenario: An Ingress with path rules /foo should not send traffic to the matching backend service
  (default /foo dose not matches request /)
    When I send a "GET" request to "http://path-rules-slash-foo/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (default /foo matches request /foo)
    When I send a "GET" request to "http://path-rules-slash-foo/foo"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo" service

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (default /foo matches request /foo/)
    When I send a "GET" request to "http://path-rules-slash-foo/foo/"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo" service

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (default /foo matches request /foo/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-foo/foo/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo" service

  Scenario: An Ingress with path rules /foo should not send traffic to the matching backend service
  (Exact /foo dose not matches request /)
    When I send a "GET" request to "http://path-rules-slash-foo-exact/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (Exact /foo matches request /foo)
    When I send a "GET" request to "http://path-rules-slash-foo-exact/foo"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-exact" service

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (Exact /foo dose not matches request /foo/)
    When I send a "GET" request to "http://path-rules-slash-foo-exact/foo/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (Exact /foo dose not matches request /foo/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-foo-exact/foo/a/b/c/d/e/f/g"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo should not send traffic to the matching backend service
  (Prefix /foo dose not matches request /)
    When I send a "GET" request to "http://path-rules-slash-foo-prefix/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (Prefix /foo matches request /foo)
    When I send a "GET" request to "http://path-rules-slash-foo-prefix/foo"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-prefix" service

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (Prefix /foo matches request /foo/)
    When I send a "GET" request to "http://path-rules-slash-foo-prefix/foo/"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-prefix" service

  Scenario: An Ingress with path rules /foo should send traffic to the matching backend service
  (Prefix /foo matches request /foo/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-foo-prefix/foo/"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-prefix" service

  Scenario: An Ingress with path rules /foo/ should not send traffic to the matching backend service
  (default /foo/ dose not matches request /)
    When I send a "GET" request to "http://path-rules-slash-foo-slash/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (default /foo/ matches request /foo)
    When I send a "GET" request to "http://path-rules-slash-foo-slash/foo"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash" service

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (default /foo/ matches request /foo/)
    When I send a "GET" request to "http://path-rules-slash-foo-slash/foo/"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash" service

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (default /foo/ matches request /foo/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-foo-slash/foo/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash" service

  Scenario: An Ingress with path rules /foo/ should not send traffic to the matching backend service
  (Exact /foo/ dose not matches request /)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-exact/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo/ should not send traffic to the matching backend service
  (Exact /foo/ dose not matches request /foo)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-exact/foo"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (Exact /foo/ matches request /foo/)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-exact/foo/"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash-exact" service

  Scenario: An Ingress with path rules /foo/ should not send traffic to the matching backend service
  (Exact /foo/ dose not matches request /foo/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-exact/foo/a/b/c/d/e/f/g"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo/ should not send traffic to the matching backend service
  (Prefix /foo/ dose not matches request /)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-prefix/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (Prefix /foo/ matches request /foo)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-prefix/foo"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash-prefix" service

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (Prefix /foo/ matches request /foo/)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-prefix/foo/"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash-prefix" service

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (Prefix /foo/ matches request /foo/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-prefix/foo/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash-prefix" service

  Scenario: An Ingress with path rules /foo/ should send traffic to the matching backend service
  (Prefix /foo/ matches request /foo/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-foo-slash-prefix/foo/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "slash-foo-slash-prefix" service

  Scenario: An Ingress with path rules /aaa/bb should not send traffic to the matching backend service
  (default /aaa/bb dose not matches request /aaa/b)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb/aaa/b"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules /aaa/bb should send traffic to the matching backend service
  (default /aaa/bb matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb/aaa/bb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bb" service

  Scenario: An Ingress with path rules /aaa/bb should send traffic to the matching backend service
  (default /aaa/bb matches request /aaa/bb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb/aaa/bb/"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bb" service

  Scenario: An Ingress with path rules /aaa/bb should send traffic to the matching backend service
  (default /aaa/bb matches request /aaa/bb/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb/aaa/bb/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bb" service

  Scenario: An Ingress with path rules /aaa/bb should not send traffic to the matching backend service
  (default /aaa/bb dose not matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb/aaa/bbb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bb should not send traffic to the matching backend service
  (exact /aaa/bb dose not matches request /aaa/b)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-exact/aaa/b"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bb should send traffic to the matching backend service
  (exact /aaa/bb matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-exact/aaa/bb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bb-exact" service

  Scenario: An Ingress with path rules exact /aaa/bb should not send traffic to the matching backend service
  (exact /aaa/bb dose not matches request /aaa/bb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-exact/aaa/bb/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bb should not send traffic to the matching backend service
  (exact /aaa/bb dose not matches request /aaa/bb/a/b/c/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-exact/aaa/bb/a/b/c/e/f/g"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bb should not send traffic to the matching backend service
  (exact /aaa/bb dose not matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-exact/aaa/bbb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bb should not send traffic to the matching backend service
  (prefix /aaa/bb dose not matches request /aaa/b)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-prefix/aaa/b"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bb should send traffic to the matching backend service
  (prefix /aaa/bb matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-prefix/aaa/bb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bb-prefix" service

  Scenario: An Ingress with path rules prefix /aaa/bb should send traffic to the matching backend service
  (prefix /aaa/bb matches request /aaa/bb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-prefix/aaa/bb/"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bb-prefix" service

  Scenario: An Ingress with path rules prefix /aaa/bb should send traffic to the matching backend service
  (prefix /aaa/bb matches request /aaa/bb/a/b/c/d/e/f/g)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-prefix/aaa/bb/a/b/c/d/e/f/g"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bb-prefix" service

  Scenario: An Ingress with path rules prefix /aaa/bb should send traffic to the matching backend service
  (prefix /aaa/bb matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bb-prefix/aaa/bbb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules default /aaa/bbb should not send traffic to the matching backend service
  (default /aaa/bbb dose not matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb/aaa/bb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules default /aaa/bbb should send traffic to the matching backend service
  (default /aaa/bbb matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb" service

  Scenario: An Ingress with path rules default /aaa/bbb should not send traffic to the matching backend service
  (default /aaa/bbb dose not matches request /aaa/bbbxyz)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb/aaa/bbbxyz"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules default /aaa/bbb should send traffic to the matching backend service
  (default /aaa/bbb matches request /aaa/bbb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb/aaa/bbb/"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb" service

  Scenario: An Ingress with path rules default /aaa/bbb should send traffic to the matching backend service
  (default /aaa/bbb matches request /aaa/bbb/ccc)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb/aaa/bbb/ccc"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb" service

  Scenario: An Ingress with path rules exact /aaa/bbb should not send traffic to the matching backend service
  (exact /aaa/bbb dose not matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-exact/aaa/bb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bbb should send traffic to the matching backend service
  (exact /aaa/bbb matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-exact/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-exact" service

  Scenario: An Ingress with path rules exact /aaa/bbb should not send traffic to the matching backend service
  (exact /aaa/bbb dose not matches request /aaa/bbb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-exact/aaa/bbb/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bbb should not send traffic to the matching backend service
  (exact /aaa/bbb dose not matches request /aaa/bbb/ccc)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-exact/aaa/bbb/ccc"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bbb should not send traffic to the matching backend service
  (prefix /aaa/bbb dose not matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-prefix/aaa/bb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bbb should send traffic to the matching backend service
  (prefix /aaa/bbb matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-prefix/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-prefix" service

  Scenario: An Ingress with path rules prefix /aaa/bbb should not send traffic to the matching backend service
  (prefix /aaa/bbb dose not matches request /aaa/bbbxyz)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-prefix/aaa/bbbxyz"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bbb should not send traffic to the matching backend service
  (prefix /aaa/bbb dose not matches request /aaa/bbb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-prefix/aaa/bbb/"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-prefix" service

  Scenario: An Ingress with path rules default /aaa/bbb/ should not send traffic to the matching backend service
  (default /aaa/bbb/ dose not matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash/aaa/bb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules default /aaa/bbb/ should send traffic to the matching backend service
  (default /aaa/bbb/ matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-slash" service

  Scenario: An Ingress with path rules default /aaa/bbb/ should send traffic to the matching backend service
  (default /aaa/bbb/ matches request /aaa/bbbxyz)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash/aaa/bbbxyz"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules default /aaa/bbb/ should send traffic to the matching backend service
  (default /aaa/bbb/ matches request /aaa/bbb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash/aaa/bbb/"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-slash" service

  Scenario: An Ingress with path rules default /aaa/bbb/ should send traffic to the matching backend service
  (default /aaa/bbb/ matches request /aaa/bbb/ccc)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash/aaa/bbb/ccc"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-slash" service

  Scenario: An Ingress with path rules exact /aaa/bbb/ should send not traffic to the matching backend service
  (exact /aaa/bbb/ dose not matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-exact/aaa/bb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bbb/ should send not traffic to the matching backend service
  (exact /aaa/bbb/ dose not matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-exact/aaa/bbb"
    Then the response status-code must be 500
    #And the response must be served by the "slash-aaa-slash-bbb-slash-exact" service

  Scenario: An Ingress with path rules exact /aaa/bbb/ should send not traffic to the matching backend service
  (exact /aaa/bbb/ dose not matches request /aaa/bbbxyz)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-exact/aaa/bbbxyz"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules exact /aaa/bbb/ should send not traffic to the matching backend service
  (exact /aaa/bbb/ dose not matches request /aaa/bbb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-exact/aaa/bbb/"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-slash-exact" service

  Scenario: An Ingress with path rules exact /aaa/bbb/ should send not traffic to the matching backend service
  (exact /aaa/bbb/ dose not matches request /aaa/bbb/ccc)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-exact/aaa/bbb/ccc"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bbb/ should send not traffic to the matching backend service
  (prefix /aaa/bbb/ dose not matches request /aaa/bb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-prefix/aaa/bb"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bbb/ should send traffic to the matching backend service
  (prefix /aaa/bbb/ matches request /aaa/bbb)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-prefix/aaa/bbb"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-slash-prefix" service

  Scenario: An Ingress with path rules prefix /aaa/bbb/ should send traffic to the matching backend service
  (prefix /aaa/bbb/ matches request /aaa/bbbxyz)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-prefix/aaa/bbbxyz"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules prefix /aaa/bbb/ should send traffic to the matching backend service
  (prefix /aaa/bbb/ matches request /aaa/bbb/)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-prefix/aaa/bbb/"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-slash-prefix" service

  Scenario: An Ingress with path rules prefix /aaa/bbb/ should send traffic to the matching backend service
  (prefix /aaa/bbb/ matches request /aaa/bbb/ccc)
    When I send a "GET" request to "http://path-rules-slash-aaa-slash-bbb-slash-prefix/aaa/bbb/ccc"
    Then the response status-code must be 200
    And the response must be served by the "slash-aaa-slash-bbb-slash-prefix" service
