@annotations @balance.weight @release-1.22
Feature: Load balance rules

  Scenario: An Ingress with load balance rule 50-50 should send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":50, \"whoami2\":50}}"
    spec:
      rules:
        - host: "balance-weight-50-50"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send 100 "GET" requests to "http://balance-weight-50-50/"
    Then the response status-code must be 200 the response body should contain the IP address of 2 different Kubernetes pods
    And the response must be served by one of "whoami1|whoami2" service

  Scenario: An Ingress with load balance rule 1-1 should send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":1, \"whoami2\":1}}"
    spec:
      rules:
        - host: "balance-weight-1-1"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send 100 "GET" requests to "http://balance-weight-1-1/"
    Then the response status-code must be 200 the response body should contain the IP address of 2 different Kubernetes pods
    And the response must be served by one of "whoami1|whoami2" service

  Scenario: An Ingress with load balance rule 3-1 should send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":3, \"whoami2\":1}}"
    spec:
      rules:
        - host: "balance-weight-3-1"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send 100 "GET" requests to "http://balance-weight-3-1/"
    Then the response status-code must be 200 the response body should contain the IP address of 2 different Kubernetes pods
    And the response must be served by one of "whoami1|whoami2" service

  Scenario: An Ingress with load balance rule 1-1-1 should send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000|whoami3:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":1, \"whoami2\":1, \"whoami3\":1}}"
    spec:
      rules:
        - host: "balance-weight-1-1-1"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send 100 "GET" requests to "http://balance-weight-1-1-1/"
    Then the response status-code must be 200 the response body should contain the IP address of 3 different Kubernetes pods
    And the response must be served by one of "whoami1|whoami2|whoami3" service

  Scenario: An Ingress with load balance rule 200-177-121 should send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000|whoami3:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":200, \"whoami2\":177, \"whoami3\":121}}"
    spec:
      rules:
        - host: "balance-weight-200-177-121"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send 100 "GET" requests to "http://balance-weight-200-177-121/"
    Then the response status-code must be 200 the response body should contain the IP address of 3 different Kubernetes pods
    And the response must be served by one of "whoami1|whoami2|whoami3" service

  Scenario: An Ingress with load balance rule format error should not send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":50,, \"whoami2\":50}}"
    spec:
      rules:
        - host: "balance-weight-50-50"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with load balance rule 0-0-0 should not send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000|whoami3:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":0, \"whoami2\":0, \"whoami3\":0}}"
    spec:
      rules:
        - host: "balance-weight-0-0-0"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with load balance rule 1-1-(-1) should not send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000|whoami3:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":1, \"whoami2\":1, \"whoami3\":-1}}"
    spec:
      rules:
        - host: "balance-weight-1-1-mimus-1"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with load balance rule 0.33-0.33-0.34 should not send traffic to the matching backend service
  (exact / and cookie matches request /)
    Given an Ingress with service info "whoami1:3000|whoami2:3000|whoami3:3000" resource in a new random namespace
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: balance-rules
      annotations:
        bfe.ingress.kubernetes.io/balance.weight: "{\"service\":{\"whoami1\":0.33, \"whoami2\":0.33, \"whoami3\":0.34}}"
    spec:
      rules:
        - host: "balance-weight-0.33-0.33-0.34"
          http:
            paths:
              - path: /
                pathType: Exact
                backend:
                  service:
                    name: service
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

