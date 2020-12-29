@sig-network @conformance @v1alpha1
Feature: HTTP Gateway
  A Gateway may define routing rules based on the request host.
  
  If the HTTP request host matches one of the hosts in the Gateway objects, the
  traffic is routed through a selected HTTPRoute to a backend service.

  Background:
    Given a new random namespace
    Given a self-signed TLS secret named "conformance-tls" for the "foo.bar.com" hostname
    Given the "httpgateway" YAML scenario
    Then The Gateway status should include Scheduled and Ready conditions set to "True"
    Then The Gateway status should include 1 Listener with a Ready condition set to "True"

  Scenario: A request to the specified host should be routed to the designated Service
    (host foo.bar.com matches request foo.bar.com)

    When I send a "GET" request to "https://foo.bar.com"
    Then the secure connection must verify the "foo.bar.com" hostname
    And the response status-code must be 200
    And the response must be served by the "foo-bar-com" service
    And the request host must be "foo.bar.com"
