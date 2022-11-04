@annotations @router.cookie @release-1.22
Feature: Redirect
  An Ingress may define redirection rules in its annotations.

  If the a Request matches the rule defined in a Ingress which have redirection annotations,
  BFE should return a Response whose header has the corresponding Location and Status Code according to the annotations in the Ingress.

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-set` but
  no annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-set: "https://www.baidu.com"
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 302
    And the response location must be "https://www.baidu.com"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-set` and
  annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-set: "https://www.baidu.com"
        bfe.ingress.kubernetes.io/redirect.response-status: 301
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 301
    And the response location must be "https://www.baidu.com"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-from-query` but
  no annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-from-query
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-from-query: "location"
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar?location=https://www.baidu.com"
    Then the response status-code must be 302
    And the response location must be "https://www.baidu.com"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-from-query` and
  annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-from-query
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-from-query: "location"
        bfe.ingress.kubernetes.io/redirect.response-status: 301
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar?location=https://www.baidu.com"
    Then the response status-code must be 301
    And the response location must be "https://www.baidu.com"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-prefix-add` but
  no annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-prefix-add: "https://new.org/api"
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 302
    And the response location must be "https://new.org/api/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-prefix-add` and
  annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-prefix-add: "https://new.org/api"
        bfe.ingress.kubernetes.io/redirect.response-status: 301
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 301
    And the response location must be "https://new.org/api/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.scheme-set` but
  no annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.scheme-set: https
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 302
    And the response location must be "https://foo.com/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.scheme-set` and
  annotation `bfe.ingress.kubernetes.io/redirect.response-status`.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.scheme-set: https
        bfe.ingress.kubernetes.io/redirect.response-status: 301
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 301
    And the response location must be "https://foo.com/bar"

  Scenario: An Ingress with redirect annotations is applied and
  then the ingress is updated by removing the redirect annotations.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-set: "https://www.baidu.com"
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 302
    And the response location must be "https://www.baidu.com"
    Then update the ingress by removing the redirect annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://foo.com/bar"
    Then the response status-code must be 200


  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-set` whose value
  is invalid
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-set: aaa-bbb
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-prefix-add` whose value
  is invalid url
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-prefix-add: invalid-url
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.url-prefix-add` whose value
  is valid url with fragment
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.url-prefix-add: https://example.org/path#fragment
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/redirect.scheme-set` whose value
  is not http or https
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.scheme-set: ftp
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with multiple redirection location annotations should not be created successfully.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.scheme-set: https
        bfe.ingress.kubernetes.io/redirect.url-prefix-add: "https://new.org/api"
        bfe.ingress.kubernetes.io/redirect.response-status: 301
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with redirection response-status annotation but no redirection location annotations
  should not be created successfully.
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.response-status: 301
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with redirection response-status annotation whose value is not 3XX
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.scheme-set: https
        bfe.ingress.kubernetes.io/redirect.response-status: 404
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success

  Scenario: An Ingress with redirection response-status annotation whose value type is not string
    Given an Ingress resource with redirection annotations
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: redirect-url-set
      annotations:
        bfe.ingress.kubernetes.io/redirect.scheme-set: https
        bfe.ingress.kubernetes.io/redirect.response-status: some-code
    spec:
      rules:
        - host: "foo.com"
          http:
            paths:
              - path: /bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status should not be success
