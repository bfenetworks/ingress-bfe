@annotations @router.cookie @release-1.22
Feature: Rewrite
  An Ingress may define rewrite rules in its annotations.

  If the a Request matches the rule defined in a Ingress which have rewrite annotations,
  BFE should rewrite the host, path and query param according to the annotations in the Ingress.

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.host`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-host-static
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.host: '[{"params": "bar.com", "when": "AfterLocation"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar"
    Then the response status code must be 200
    And the request host must be "bar.com"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-host-dynamic
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix: '[{"params": "true"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
          http:
            paths:
              - path: /baidu.com/bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://rewrite-url.com/baidu.com/bar"
    Then the response status code must be 200
    And the request host must be "baidu.com"
    And the request path must be "/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix` and the route path doesn't has host prefix
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-host-dynamic
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix: '[{"params": "true"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar"
    Then the response status code must be 200
    And the request host must be "rewrite-url.com"
    And the request path must be "/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-static
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path: '[{"params": "/index"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar"
    Then the response status code must be 200
    And the request path must be "/index"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-add
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/foo/"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar"
    Then the response status code must be 200
    And the request path must be "/foo/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-delete
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/foo"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
          http:
            paths:
              - path: /foo/bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://rewrite-url.com/foo/bar"
    Then the response status code must be 200
    And the request path must be "/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim`
  and the route path doesn't have the specific prefix to be removed.
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-delete
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/foo"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar"
    Then the response status code must be 200
    And the request path must be "/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-strip
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip: '[{"params": "1"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
          http:
            paths:
              - path: /foo/bar
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://rewrite-url.com/foo/bar"
    Then the response status code must be 200
    And the request path must be "/bar"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip`,
  but the length of route path segments is less than the strip length
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-strip
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-strip: '[{"params": "3"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: foo-exact
                    port:
                      number: 3000
    """
    And The Ingress status shows the IP address or FQDN where it is exposed
    When I send a "GET" request to "http://rewrite-url.com/"
    Then the response status code must be 200
    And the request path must be "/"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.query-add`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-query-add
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.query-add: '[{"params": {"user": "bob", "age": "18"}}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar"
    Then the response status code must be 200
    And the request path must be "/bar?age=18&user=bob"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.query-delete`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-query-delete
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.query-delete: '[{"params": ["user"]}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar?user=bob&age=18"
    Then the response status code must be 200
    And the request path must be "/bar?age=18"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.query-rename`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-query-rename
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.query-rename: '[{"params": {"name": "user"}}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar?name=bob"
    Then the response status code must be 200
    And the request path must be "/bar?user=bob"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-query-delete-all-except
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except: '[{"params": "name"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar?name=bob&age=18&gender=man"
    Then the response status code must be 200
    And the request path must be "/bar?name=bob"

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.host` and callback point is not AfterLocation
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-callback-illegal
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.host: '[{"params": "bar.com", "when": "BeforeForward"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.host` and `bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-host-multiple
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.host: '[{"params": "bar.com"}]'
        bfe.ingress.kubernetes.io/rewrite-url.host-from-path-prefix: '[{"params": "bar.com"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path` and `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-illegal
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path: '[{"params": "/index"}]'
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/foo/"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path`, `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-compound
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path: '[{"params": "/index"}]'
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/foo"}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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

  Scenario: An Ingress with annotation `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add`, `bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim`
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-path-compound
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-add: '[{"params": "/index/", "order": 1}]'
        bfe.ingress.kubernetes.io/rewrite-url.path-prefix-trim: '[{"params": "/bar", "order": 2}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar"
    Then the response status code must be 200
    And the request path must be "/index/bar"

  Scenario: An Ingress with multiple rewrite annotations with "order" param.
    Given an Ingress resource with rewrite annotation
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: rewrite-query-compound
      annotations:
        bfe.ingress.kubernetes.io/rewrite-url.query-delete-all-except: '[{"params": "name", "order": 1}]'
        bfe.ingress.kubernetes.io/rewrite-url.query-add: '[{"params": {"age": "18"}, "order": 3}]'
        bfe.ingress.kubernetes.io/rewrite-url.query-rename: '[{"params": {"name": "user"}, "order": 2}]'
    spec:
      rules:
        - host: "rewrite-url.com"
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
    When I send a "GET" request to "http://rewrite-url.com/bar?name=bob&gender=man"
    Then the response status code must be 200
    And the request path must be "/bar?age=18&user=bob"