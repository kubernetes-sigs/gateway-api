# GEP-957: Destination Port Matching

* Issue: [#957](https://github.com/kubernetes-sigs/gateway-api/issues/957)
* Status: Standard

## TLDR

Add a new `port` field to ParentRef to support port matching in Routes.

## Goals

* Support port matching in routes based on the destination port number of the
  request.

## Non-Goals

* Support port matching based on port name.

## Introduction

Port matching is a common service mesh use case where traffic policies/rules
need to be applied to traffic to certain destination ports. For ingress, while
the API today already supports attaching a route to a specific listener, it may
be useful to support attaching routes to listener(s) on a specified port. This
allows route authors to apply networking behaviors on a fixed port.

## API

The proposal is to add a new field `Port` to `ParentRef`:

```go
type ParentRef struct {
  ...
  // Port is the network port this Route targets. It can be interpreted
  // differently based on the type of parent resource:
  //
  // Gateway: All listeners listening on the specified port that also support
  // this kind of Route(and select this Route). It's not recommended to set
  // `Port` unless the networking behaviors specified in a Route must
  // apply to a specific port as opposed to a listener(s) whose port(s) may
  // be changed.
  // When both Port and SectionName are specified, the name and port of the
  // selected listener must match both specified values.
  //
  // Implementations MAY choose to support other parent resources.
  // Implementations supporting other types of parent resources MUST clearly
  // document how/if Port is interpreted.
  //
  // For the purpose of status, an attachment is considered successful as
  // long as the parent resource accepts it partially. For example, Gateway
  // listeners can restrict which Routes can attach to them by Route kind,
  // namespace, or hostname. If 1 of 2 Gateway listeners accept attachment from
  // the referencing Route, the Route MUST be considered successfully
  // attached. If no Gateway listeners accept attachment from this Route, the
  // Route MUST be considered detached from the Gateway.
  //
  // Support: Core
  //
  // +optional
  Port *PortNumber `json:"port,omitempty"`
  ...
}
```

The following example shows how an HTTPRoute could be applied to port 8000. In
this example, the HTTPRoute will be attached to listeners foo and bar on port
8000 but not listener baz on port 8080.
```yaml
kind: HTTPRoute
metadata:
  name: example
  namespace: example
spec:
  parentRef:
  - name: my-gateway
    port: 8000
  ...
---
kind: Gateway
metadata:
  name: my-gateway
  namespace: example
spec:
  listeners:
  - name: foo
    port: 8000
    protocol: HTTP
    ...
  - name: bar
    port: 8000
    protocol: HTTP
    ...
  - name: baz
    port: 8080
    ...
```

The following example shows how a TCPRoute could be attached to an Mesh CRD to
route all traffic in a service mesh whose original destination port is 8000 to
port 8080 of service foo.
```yaml
kind: TCPRoute
metadata:
  name: example
  namespace: example
spec:
  parentRef:
  - name: my-mesh
    group: example.io
    kind: Mesh
    port: 8000
  rules:
  - backendRefs
    - name: foo
      port: 8080
```

## Alternatives
### 1. Use SectionName in ParentRef for port matching
Port matching can be supported if SectionName accepts port numbers in addition
to listener names. This approach results in a less explicit API when a ParentRef
points to a resource that is not `Gateway`. For example, an implementation may
attach a route to an `Mesh` CRD. In this case, it's less inituitive to set
`ParentRef.SectionName` to `443` to express `route all traffic whose destination
port is 443 to ...`. It also complicates the validation on SectionName in order
to differentiate between a listener name and a port number.

### 2. Update TrafficMatches to support port matching
TrafficMatches was proposed in
[gep-735](https://gateway-api.sigs.k8s.io/geps/gep-735/) to support TCP and UDP
address matching. TrafficMatches can be extended to support port matching.
TrafficMatches will need to be added to HTTPRoute/TLSRoute if the feature is
desired there.

While this proposal works for mesh, it may be confusing for ingress because a
user can specify port matching behavior in a route that is incompatible with
the listeners the route attaches to. For example, a user can specify a match
on port 443 in a route while the route only attaches to a listener on port 80.
