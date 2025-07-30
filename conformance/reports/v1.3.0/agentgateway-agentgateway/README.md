# Agent Gateway (with kgateway)

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|experimental|[v0.6.0]|default|[report](./experimental-0.6.0-report.yaml)|

## Reproduce

```
go test./conformance -run TestConformance -args \
  --report-output /tmp/report.yaml \
  --conformance-profiles=GATEWAY-HTTP \
  --gateway-class agentgateway \
  --all-features \
  --organization agentgateway \
  --project agentgateway \
  --url http://agentgateway.dev/ \
  --version v0.6.0-dev \
  --contact "github.com/agentgateway/agentgateway/issues/new/choose"
```
