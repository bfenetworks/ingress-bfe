@annotations @router.cookie @release-1.22
Feature: Cookie rules
  An Ingress may define routing rules based on the request path.

  If the HTTP request path matches the paths and cookie info in the
  Ingress objects, the traffic is routed to its backend service.

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact / and cookie matches request / with cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: cookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
    spec:
      rules:
        - host: "exact-path-rules-cookie-uid"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-path-rules-cookie-uid/" with header
    """
    {"Cookie": ["uid=abcd"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "foo-exact" service

  Scenario: An Ingress with path rules and cookie should not send traffic to the matching backend service
  (exact / and cookie dose not matches request / without cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: cookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
    spec:
      rules:
        - host: "exact-path-rules-cookie-uid-req-no-cookie"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-path-rules-cookie-uid-req-no-cookie/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact / and cookie config length 0 matches request / with cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: cookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: ""
    spec:
      rules:
        - host: "exact-path-rules-cookie-length0"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: foo-exact-no-cookie
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-path-rules-cookie-length0/" with header
    """
    {"Cookie": ["uid=abcd"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "foo-exact-no-cookie" service

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact / and cookie matches request / with cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: cookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: ""
    spec:
      rules:
        - host: "exact-path-rules-no-cookie"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: foo-exact-no-cookie
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-path-rules-no-cookie/"
    Then the response status-code must be 200
    And the response must be served by the "foo-exact-no-cookie" service

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact / and cookie "c_ref:https://www.baidu.com/link" matches request / with right cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: cookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "c_ref:https://www.baidu.com/link"
    spec:
      rules:
        - host: "exact-with-cookie-right"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: exact-cookie-right
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-with-cookie-right/" with header
    """
    {"Cookie": ["c_ref=https://www.baidu.com/link"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "exact-cookie-right" service

  Scenario: An Ingress with path rules and error cookie should not send traffic to the matching backend service
  (exact / and cookie "c_ref:https://www.baidu.com/" not matches request / with wrong cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: cookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "c_ref:https://www.baidu.com/"
    spec:
      rules:
        - host: "exact-with-cookie-wrong"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: exact-cookie-wrong
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-with-cookie-wrong/" with header
    """
    {"Cookie": ["c_ref=https://www.baidu.com/link"]}
    """
    Then the response status-code must be 500

  Scenario: An Ingress with path rules and cookie should not send traffic to the matching backend service
  (exact / and cookie "c_ref:https://www.baidu.com/" not matches request / with no cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: cookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "c_ref:https://www.baidu.com/"
    spec:
      rules:
        - host: "exact-with-no-cookie"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: exact-cookie-wrong
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-with-no-cookie/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules and multi cookie should send traffic to the matching backend service, use last config
  (exact / and multi cookie "cookie_key:cookie_value" and "cookie_key1:cookie_value1" matches request / with cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: multicookie-rules
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "cookie_key:cookie_value"
        bfe.ingress.kubernetes.io/router.cookie: "cookie_key1:cookie_value1"
    spec:
      rules:
        - host: "multi-coookie-rule"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: multi-cookie
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://multi-coookie-rule/" with header
    """
    {"Cookie": ["cookie_key1=cookie_value1"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "multi-cookie" service