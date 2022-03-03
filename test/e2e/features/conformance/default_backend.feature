@sig-network @conformance @release-1.22
Feature: Default backend
  An Ingress with no rules sends all traffic to the single default backend.
  The default backend is part of the Ingress resource spec field `defaultBackend`.

  If none of the hosts or paths match the HTTP request in the
  Ingress objects, the traffic is routed to your default backend.

  Background:
    Given a new random namespace
    Given an Ingress resource named "default-backend" with this spec:
    """
      defaultBackend:
        service:
          name: echo-service
          port:
            number: 3000
    """
    Then The Ingress status shows the IP address or FQDN where it is exposed

  Scenario Outline: An Ingress with no rules should send all requests to the default backend
    When I send a "<method>" request to http://"<host>"/"<path>"
    Then the response status-code must be 200
    And the response must be served by the "echo-service" service
    And the response proto must be "HTTP/1.1"
    And the response headers must contain <key> with matching <value>
      | key            | value |
      | Content-Length | *     |
      | Content-Type   | *     |
      | Date           | *     |
      | Server         | *     |
    And the request method must be "<method>"
    And the request path must be "<path>"
    And the request proto must be "HTTP/1.1"
    And the request headers must contain <key> with matching <value>
      | key        | value              |
      | User-Agent | Go-http-client/1.1 |

    Examples:
      | method | host      | path     |
      | GET    | my-host   |          |
      | GET    | my-host   | sub-path |
      | POST   | some-host |          |
      | PUT    |           | resource |
      | DELETE | some-host | resource |
      | PATCH  | my-host   | resource |
