@sig-network @conformance @release-1.22
Feature: Host rules
  An Ingress may define routing rules based on the request host.

  If the HTTP request host matches one of the hosts in the
  Ingress objects, the traffic is routed to its backend service.

  Background:
    Given a new random namespace
    Given a self-signed TLS secret named "conformance-tls" for the "foo.bar.com" hostname
    Given an Ingress resource
    """
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: host-rules
    spec:
      tls:
        - hosts:
            - foo.bar.com
          secretName: conformance-tls
      rules:
        - host: "*.foo.com"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: wildcard-foo-com
                    port:
                      number: 8080
    
        - host: foo.bar.com
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: foo-bar-com
                    port:
                      number: 80

    """
    Then The Ingress status shows the IP address or FQDN where it is exposed


  Scenario: An Ingress with a host rule should send TLS traffic to the matching backend service
  (host foo.bar.com matches request foo.bar.com)

    When I send a "GET" request to "https://foo.bar.com"
    Then the secure connection must verify the "foo.bar.com" hostname
    And the response status-code must be 200
    And the response must be served by the "foo-bar-com" service
    And the request host must be "foo.bar.com"

  Scenario: An Ingress with a host rule should send traffic to the matching backend service
  (host foo.bar.com matches request foo.bar.com)

    When I send a "GET" request to "http://foo.bar.com"
    And the response status-code must be 200
    And the response must be served by the "foo-bar-com" service
    And the request host must be "foo.bar.com"

  Scenario: An Ingress with a host rule should not route traffic when hostname does not match
  (host foo.bar.com does not match request subdomain.bar.com)

    When I send a "GET" request to "http://subdomain.bar.com"
    Then the response status-code must be 500

  Scenario: An Ingress with a wildcard host rule should send traffic to the matching backend service
  (Matches based on shared suffix)

    When I send a "GET" request to "http://bar.foo.com"
    Then the response status-code must be 200
    And the response must be served by the "wildcard-foo-com" service
    And the request host must be "bar.foo.com"

  Scenario: An Ingress with a wildcard host rule should not route traffic matching on more than a single dns label
  (No match, wildcard only covers a single DNS label)

    When I send a "GET" request to "http://baz.bar.foo.com"
    Then the response status-code must be 500

  Scenario: An Ingress with a wildcard host rule should not route traffic matching no dns label
  (No match, wildcard only covers a single DNS label)

    When I send a "GET" request to "http://foo.com"
    Then the response status-code must be 500
