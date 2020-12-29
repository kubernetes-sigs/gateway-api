@sig-network @conformance @v1alpha1
Feature: HTTP Gateway
  A HTTP Gateway should have appropriate status conditions set.

  Background:
    Given a new random namespace
    Given the "httpgateway" scenario
  
  Scenario:
    Then Gateway "gateway-conformance" should have "Scheduled" condition should be set to "True" within 3 minutes
    Then Gateway "gateway-conformance" should have "Ready" condition should be set to "True" within 3 minutes
    Then Gateway "gateway-conformance" should have an address in status within 3 minutes
