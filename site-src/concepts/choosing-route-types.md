# Choosing the Right Route Type for Your Application

When deploying an application on Kubernetes using Gateway API, one of the first questions developers face is:

**“Which Route type should I use to expose my service?”**

Gateway API defines multiple Route kinds—`HTTPRoute`, `GRPCRoute`, `TLSRoute`, `TCPRoute`, and `UDPRoute`—each designed for specific protocols and traffic behaviors.  
Choosing the correct Route ensures your traffic is handled correctly, securely, and efficiently.

This guide explains:

- How each Route type works and the problems they solve  
- How Route types differ in routing discriminators, OSI layers, and TLS behavior  
- How to select the right Route for your application    
- When to choose `HTTPRoute` vs `TLSRoute` for HTTPS
- What to verify in your Gateway and with your administrator before attaching a Route  

---

## 1. Route Types at a Glance

The table below summarizes Route types, from the official Route Summary Table.

### Route Summary Table

| Object      | OSI Layer                           | Routing Discriminator         | TLS Support                 | Purpose                                                                 |
|-------------|--------------------------------------|-------------------------------|-----------------------------|-------------------------------------------------------------------------|
| UDPRoute    | Layer 4                              | destination port              | None                        | Allows forwarding of a UDP stream from the Listener to the Backends.   |
| TLSRoute    | Somewhere between Layer 4 and 7      | SNI or other TLS properties   | Passthrough or Terminated   | Routing of TLS protocols including HTTPS where HTTP inspection is not required. |
| TCPRoute    | Layer 4                              | destination port              | Terminated                  | Allows forwarding of a TCP stream from the Listener to the Backends.   |
| HTTPRoute   | Layer 7                              | Anything in the HTTP protocol | Terminated only             | HTTP and HTTPS routing.                                                 |
| GRPCRoute   | Layer 7                              | Anything in the gRPC protocol | Terminated only             | gRPC routing over HTTP/2 and HTTP/2 cleartext.                          |

This provides a quick mental model before diving deeper.

---

## 2. Route Types Explained

### HTTPRoute (Standard)

Use `HTTPRoute` when your workload uses HTTP or HTTPS **and you want L7 routing features**, such as:

- Path-based routing  
- Hostname or header-based matching  
- Redirects and rewrites  
- Filters, timeouts, retries  

Most HTTPS workloads use `HTTPRoute` because **the Gateway terminates TLS** and can inspect HTTP requests.

**Example:**

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

Use `GRPCRoute` for workloads using gRPC. It routes based on:

- gRPC service names
- gRPC method names

This Route type requires **HTTP/2** support on the Gateway.

**Example:**

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

Use `TLSRoute` when your application uses TLS and the Gateway should **not** decrypt the traffic.  
Routing is based only on **SNI (Server Name Indication)**.

Common scenarios include:

- End-to-end TLS encryption
- Applications that handle their own certificates
- Routing TLS protocols without HTTP parsing (including HTTPS passthrough)

`TLSRoute` cannot match HTTP paths or headers because the traffic remains encrypted.

**Example:**

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

Use `TCPRoute` for applications that communicate using raw TCP.  
This provides simple Layer-4 forwarding with no HTTP or TLS inspection.

Common use cases include:

- Databases (PostgreSQL, MySQL)
- Message brokers (MQTT, AMQP)
- Mail protocols (SMTP, IMAP)

**Example:**

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

Use `UDPRoute` for applications that use UDP.  
UDP traffic is connectionless and forwarded at Layer 4 without protocol inspection.

Common examples include:

- DNS servers
- Game servers
- RTP or other media streaming workloads

**Example:**

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

| Application Type       | Route Type | Reason                            |
|------------------------|------------|-----------------------------------|
| Website / REST API     | HTTPRoute  | Full L7 routing with TLS termination |
| gRPC service           | GRPCRoute  | Service/method-level routing        |
| HTTPS passthrough      | TLSRoute   | Backend terminates TLS              |
| Database               | TCPRoute   | Raw TCP protocol                    |
| DNS                    | UDPRoute   | UDP-based protocol                  |
| Game server            | UDPRoute   | UDP workload                        |

## 4. Understanding TLS Termination and SNI Routing

Choosing between `HTTPRoute` and `TLSRoute` for HTTPS traffic depends on how TLS is handled by the Gateway.

### When the Gateway Terminates TLS

If the Gateway decrypts the TLS session:

- It can inspect the HTTP request.
- It can match paths, headers, hostnames, and apply filters.
- Full Layer-7 routing features become available.

**Use `HTTPRoute` in this case.**

### When the Gateway Uses TLS Passthrough

If the Gateway does **not** decrypt TLS:

- It only sees the TLS ClientHello.
- Routing is based solely on **SNI (Server Name Indication)**.
- The backend service is responsible for terminating TLS.

**Use `TLSRoute` in this case.**

### Why SNI Matters

SNI is the hostname sent by the client before encryption begins.  
Because the Gateway cannot inspect HTTP headers or paths inside encrypted TLS traffic, SNI becomes the only routing discriminator available for passthrough scenarios.

`TLSRoute` **cannot** perform path or header matching.

### Summary of HTTPS Routing

| Scenario                         | Gateway Behavior | Route Type  |
|----------------------------------|------------------|-------------|
| Gateway needs HTTP inspection    | Terminates TLS   | HTTPRoute   |
| Gateway should not inspect HTTP  | TLS passthrough  | TLSRoute    |
| Non-HTTP TLS protocols           | Passthrough      | TLSRoute    |


# 5. Prerequisites Before Selecting a Route

Before attaching a Route, verify the Gateway and GatewayClass support it.

Gateway Configuration

Check listener protocols:

```yaml
spec:
  listeners:
    - protocol: HTTP
    - protocol: HTTPS
    - protocol: TLS
    - protocol: TCP
    - protocol: UDP
```
Routes must match listener protocols.

**Allowed Route Types**

Gateways may restrict Route kinds:
```yaml
allowedRoutes:
  kinds:
    - kind: HTTPRoute
```
### GatewayClass Implementation Support

These checks ensure the implementation you are using actually supports the Route type you want to attach.

Confirm:

- HTTPS termination rules
- TLS passthrough support
- GRPCRoute support
- TCPRoute and UDPRoute support

### Administrator Policies

These checks ensure your platform or networking team allows the protocol behavior your application requires.

Check:

- TLS termination model
- Allowed protocols
- Security requirements
- Restrictions on TLS passthrough

## 6. Summary

Choosing the correct Route type depends on:

- The protocol your application uses
- Whether TLS is terminated or passed through
- Gateway listener configuration
- GatewayClass capabilities
- Cluster networking policies

Selecting the correct Route ensures correct traffic handling and optimal use of Gateway API features.
