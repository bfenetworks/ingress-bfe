@annotations @router.header @release-1.22
Feature: Header rules
  An Ingress may define routing rules based on the request path and header.

  If the HTTP request path matches the paths and header info in the
  Ingress objects, the traffic is routed to its backend service.

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header matches request / with header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "h_key:h_value"
    spec:
      rules:
        - host: "exact-path-header-key"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: exact-h-key
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-path-header-key/" with header
    """
    {"h_key": ["h_value"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "exact-h-key" service

  Scenario: An Ingress with path rules and header should not send traffic to the matching backend service
  (exact / and header matches request / without header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "h_key:h_value"
    spec:
      rules:
        - host: "exact-path-no-header"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: exact-h-key
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://exact-path-no-header/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header null matches request / with header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: ""
    spec:
      rules:
        - host: "header-null"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: h-key-null
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-null/" with header
    """
    {"h_key": ["h_value"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "h-key-null" service

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header null matches request / without header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: ""
    spec:
      rules:
        - host: "header-null"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: h-key-null-without-header
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-null/"
    Then the response status-code must be 200
    And the response must be served by the "h-key-null-without-header" service

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header only key matches request / with header only key)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "h-key"
    spec:
      rules:
        - host: "header-only-key"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: header-only-key
                    port:
                      number: 3000
    """
    And The Ingress status should not be success


  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header only key matches request / without header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "h-key"
    spec:
      rules:
        - host: "header-only-key"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: header-only-key-no-header
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header normal matches request / with header right)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "User-Agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36"
    spec:
      rules:
        - host: "header-normal-user-agent"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: header-normal-right
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-normal-user-agent/" with header
    """
    {"User-Agent": ["Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "header-normal-right" service

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header normal dose not matches request / without header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "user-agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36"
    spec:
      rules:
        - host: "header-normal"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: header-normal-right
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-normal/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header a:b:c matches request / with header right)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "a:b:c"
    spec:
      rules:
        - host: "header-a-b-c"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: header-a-b-c-right
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-a-b-c/" with header
    """
    {"a": ["b:c"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "header-a-b-c-right" service

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header a:b:c dose not matches request / without header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "a:b:c"
    spec:
      rules:
        - host: "header-a-b-c-without"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: without-header-a-b-c
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-a-b-c-without/"
    Then the response status-code must be 500

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header a:+/?%#& matches request / with header right)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "a:+/?%#&"
    spec:
      rules:
        - host: "header-special-character"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: header-special-character-right
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-special-character/" with header
    """
    {"a": ["+/?%#&"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "header-special-character-right" service

  Scenario: An Ingress with path rules and header should send traffic to the matching backend service
  (exact / and header a:+/?%#& matches request / without header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: header-rules
      annotations:
        bfe.ingress.kubernetes.io/router.header: "a:+/?%#&"
    spec:
      rules:
        - host: "header-special-character-no"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: header-special-character-no
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://header-special-character-no/"
    Then the response status-code must be 500