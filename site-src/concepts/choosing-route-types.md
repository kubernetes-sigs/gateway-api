# Choosing the Right Route Type for Your Application

When deploying an application on Kubernetes, one of the first questions developers face is:

**“Which Route type should I use to expose my service?”**

Gateway API provides multiple Route kinds—HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, and UDPRoute—each designed for different traffic patterns and application protocols.

This guide helps you understand:

- [What to verify on your Gateway before selecting a Route](#1-before-choosing-a-route-check-your-gateway)
- [How each Route type works and what problems it solves](#2-route-types-explained)
- [How to choose the right Route type for your application](#3-choosing-the-right-route--decision-table)
- [When to use HTTPRoute versus TLSRoute for HTTPS traffic](#4-correct-rule-for-https)
- [What information to confirm with your cluster administrator](#5-what-to-ask-your-administrator)
- [Examples of each Route type in real deployments](#2-route-types-explained)


---

## 1. Before Choosing a Route: Check Your Gateway

Each Gateway declares its listeners, which define the protocols and ports it supports:

```yaml
spec:
  listeners:
    - name: http
      protocol: HTTP
      port: 80
    - name: https
      protocol: HTTPS
      port: 443
    - name: tls
      protocol: TLS
      port: 8443
    - name: tcp
      protocol: TCP
      port: 5432
```
Before creating a Route, verify:

- Does the Gateway support your protocol?
- Does the Gateway allow the Route kind you want to attach?
- Does the GatewayClass implementation support that Route type?

Your Route must match the listener protocol.

If you're unsure, check:

- The Gateway YAML
- Your GatewayClass documentation
- Your administrator’s networking policies

---

## 2. Route Types Explained

### HTTPRoute (Standard)

Use this when your application uses HTTP or HTTPS and you want features such as:

- Path-based routing
- Hostname or header routing
- Redirects, rewrites, filters

Most HTTPS workloads use HTTPRoute because the Gateway terminates TLS.

Example:

```yaml
kind: HTTPRoute
spec:
  parentRefs:
  - name: gateway
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    backendRefs:
    - name: api-backend
      port: 8080
```
### GRPCRoute (Standard)

Use this when your application uses gRPC and requires routing at the gRPC service or method level.

This Route type requires HTTP/2 support on the Gateway.

Example:

```yaml
kind: GRPCRoute
spec:
  parentRefs:
  - name: gateway
  rules:
  - matches:
    - method:
        service: payments.Service
        method: Charge
    backendRefs:
    - name: payments-backend
      port: 50051
```
### TLSRoute (Promoted)

Use this when your application uses TLS and you want the Gateway to pass encrypted traffic directly to the backend without decrypting it.

Common scenarios include:

- End-to-end TLS encryption
- Applications that manage their own certificates
- Routing based only on the SNI (TLS hostname)

TLSRoute does not support HTTP-level routing such as paths or headers.  
Use this only when the Gateway should not terminate TLS.

Example:

```yaml
kind: TLSRoute
spec:
  parentRefs:
  - name: gateway
  rules:
  - matches:
    - snis: ["secure.example.com"]
    backendRefs:
    - name: secure-app
      port: 8443
```
### TCPRoute (Experimental)

Use this when your application communicates using raw TCP.

Common examples include:

- Databases such as Postgres or MySQL
- Message brokers such as MQTT or AMQP
- Mail protocols like SMTP or IMAP

TCPRoute provides simple Layer-4 forwarding without HTTP or TLS inspection.

Example:

```yaml
kind: TCPRoute
spec:
  parentRefs:
  - name: gateway
  backendRefs:
  - name: postgres
    port: 5432
```
### UDPRoute (Experimental)

Use this when your application uses UDP.

Common examples include:

- DNS servers
- Game servers
- RTP or other streaming workloads

UDPRoute forwards packets at Layer 4 without connection state or protocol inspection.

Example:

```yaml
kind: UDPRoute
spec:
  parentRefs:
  - name: gateway
  backendRefs:
  - name: dns-backend
    port: 53
```
## 3. Choosing the Right Route — Decision Table

The table below provides a quick reference for selecting the appropriate Route type based on your application's protocol and routing needs.

| Application Type       | Route Type | Reason                        |
|------------------------|------------|-------------------------------|
| Website / REST API     | HTTPRoute  | Layer-7 routing, TLS termination |
| gRPC service           | GRPCRoute  | Service and method-based routing |
| HTTPS passthrough      | TLSRoute   | Backend terminates TLS          |
| Database               | TCPRoute   | Raw TCP protocol                 |
| DNS                    | UDPRoute   | Uses UDP                        |
| Game server            | UDPRoute   | Uses UDP                        |

## 4. Correct Rule for HTTPS

Choosing between HTTPRoute and TLSRoute for HTTPS traffic depends on how TLS is handled by the Gateway.

### When the Gateway terminates TLS  
Use **HTTPRoute**.

This allows the Gateway to inspect the HTTP request and apply features such as:

- Path matching
- Hostname-based routing
- Header-based routing
- Redirects, rewrites, and filters

### When the Gateway passes TLS through  
Use **TLSRoute**.

In this case:

- The Gateway does not decrypt the TLS traffic  
- Only SNI-based routing is possible  
- The backend service is responsible for terminating TLS  

TLSRoute is not used for plain-text traffic.

## 5. What to Ask Your Administrator

Before creating a Route, confirm the following details with your administrator or platform team:

### Gateway capabilities
- Which listener protocols are enabled?
- Does HTTPS terminate TLS or use passthrough?
- Are TCP or UDP listeners exposed?

### Allowed Route types
Check whether the Gateway restricts which Route kinds can attach to its listeners, using the `allowedRoutes` field.

### Implementation support
- Does the GatewayClass implementation support GRPCRoute?
- Are experimental Route types such as TCPRoute or UDPRoute enabled?
- Are there restrictions on using TLS passthrough in production environments?

## 6. Summary

Choosing the correct Route type depends on understanding how your application communicates and how your Gateway is configured.

Key factors include:

1. The protocol your application uses  
2. Whether TLS is terminated at the Gateway or passed through  
3. The listener protocols exposed by the Gateway  
4. The capabilities of your GatewayClass implementation  
5. Any networking or security policies enforced in your cluster  

By matching your workload to the appropriate Route type, you ensure correct traffic handling and take advantage of the routing features provided by Gateway API.

