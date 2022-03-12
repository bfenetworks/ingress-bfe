@annotations @route.header @router.cookie @release-1.22
Feature: Priority rules
  An Ingress may define routing rules based on the request path.

  If the HTTP request path matches the paths and cookie info in the
  Ingress objects, the traffic is routed to its backend service.

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact /, cookie, header matches request / with cookie and header)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-diff
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
        bfe.ingress.kubernetes.io/router.header: "h_key:h_value"
    spec:
      rules:
        - host: "same-host-same-path-num-diff"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service-match-2
                    port:
                      number: 3000
    """
    And an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-diff
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
    spec:
      rules:
        - host: "same-host-same-path-num-diff"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service-match-1
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://same-host-same-path-num-diff/" with header
    """
    {"Cookie": ["uid=abcd"],"h_key": ["h_value"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "service-match-2" service

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact /, cookie matches request / with cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-diff
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
        bfe.ingress.kubernetes.io/router.header: "h_key:h_value"
    spec:
      rules:
        - host: "same-host-same-path-num-diff"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-match-2
                    port:
                      number: 3000
    """
    And an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-diff
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
    spec:
      rules:
        - host: "same-host-same-path-num-diff"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-match-1
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://same-host-same-path-num-diff/" with header
    """
    {"Cookie": ["uid=abcd"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "service-match-1" service

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact /, cookie matches request / with cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-diff
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
        bfe.ingress.kubernetes.io/router.header: "h_key:h_value"
    spec:
      rules:
        - host: "same-host-same-path-num-diff"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-match-2
                    port:
                      number: 3000
    """
    And an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-diff
      annotations:
        bfe.ingress.kubernetes.io/router.header: "h_key:h_value"
    spec:
      rules:
        - host: "same-host-same-path-num-diff"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-match-1
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://same-host-same-path-num-diff/" with header
    """
    {"h_key": ["h_value"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "service-match-1" service

  Scenario: An Ingress with path rules and cookie should send traffic to the matching backend service
  (exact /, cookie matches request / with cookie)
    Given an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-same
      annotations:
        bfe.ingress.kubernetes.io/router.cookie: "uid:abcd"
    spec:
      rules:
        - host: "same-host-same-path-num-same"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-match-2
                    port:
                      number: 3000
    """
    And an Ingress resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: same-host-same-path-num-same1
      annotations:
        bfe.ingress.kubernetes.io/router.header: "h_key:h_value"
    spec:
      rules:
        - host: "same-host-same-path-num-same"
          http:
            paths:
              - path: /
                pathType: ImplementationSpecific
                backend:
                  service:
                    name: service-match-1
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://same-host-same-path-num-same/" with header
    """
    {"Cookie": ["uid=abcd"],"h_key": ["h_value"]}
    """
    Then the response status-code must be 200
    And the response must be served by the "service-match-2" service