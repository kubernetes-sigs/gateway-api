# GEP-735: TCP and UDP addresses matching

* Issue: [#735](https://github.com/kubernetes-sigs/gateway-api/issues/735)
* Status: Declined

## Notes about declined status

At one point before the release of `v0.5.0` we did have an implementation
of this GEP in `main`, but we decided to pull back on it for multiple
reasons:

- operated too much like WAF/firewall functionality, which is not in scope
- no implementations championing the use case

It should also be noted that the maintainers have at least considered the
idea of an `IPRoute` API which would help differentiate this from firewall
functionality, however there haven't been any strong champions for such a
use case for this either.

As such this GEP is marked as `Declined` to make it clear to readers that
presently we don't have plans to include this in any future release. Keep
in mind that this doesn't mean that we wouldn't consider it again as a
future feature however: if you're interested in this functionality please
feel free to start a new [github discussion][disc] and/or feel free to
create a new PR updating this GEP with your use case(s) and context.

[disc]:https://github.com/kubernetes-sigs/gateway-api/discussions

## TLDR

Spec for matching source and destination addresses on L4 APIs.

## Goals

- add matching rules for address to `TCPRoute`
- add matching rules for address to `UDPRoute`
- intentionally avoid type definitions that would make it hard to expand later

## Non-Goals

- define rules for port matching

## Introduction

While `TCPRoute` and `UDPRoute` currently support custom matching extensions,
there is desire among the community to include some "fundamental" matching
options in the spec that cover the most common requirements. In this GEP we
request address matching for these APIs in order to support a standard
for some of the commonplace setups of gateway implementations. Matching is
intended to be covered for both _source_ and _destination_ to enable a finer
level of tuning options for L4 traffic routing at a level below the `Gateway`.

## API

The API changes include the following new types:

- `AddressMatch` to indicate the IP for address matching
- `AddressRouteMatches` to configure matching according to network address

These types enable the address matching required, with some active
considerations about how to leave these open ended for later expansion.

### AddressMatch Type

A new `AddressMatch` type provides the targeting mechanism for match inclusion
of a given network address:

```go
type AddressMatch struct {
	// Type of the address, either IPAddress or NamedAddress.
	//
	// If NamedAddress is used this is a custom and specific value for each
	// implementation to handle (and add validation for) according to their
	// own needs.
	//
	// For IPAddress the implementor may expect either IPv4 or IPv6.
	//
	// Support: Core (IPAddress)
	// Support: Implementation-specific (NamedAddress)
	//
	// +optional
	// +kubebuilder:validation:Enum=IPAddress;NamedAddress
	// +kubebuilder:default=IPAddress
	Type *AddressType `json:"type,omitempty"`

	// Value of the address. The validity of the values will depend
	// on the type and support by the controller.
	//
	// If implementations support proxy-protocol (see:
	// https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt) they
	// must respect the connection metadata from proxy-protocol
	// in the match logic implemented for these address values.
	//
	// Examples: `1.2.3.4`, `128::1`, `my-named-address`.
	//
	// Support: Core
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Value string `json:"value"`
}
```

### AddressRouteMatches Type

Using the new `AddressMatch` type matches can be expressed in topical lists on
`TCPRoute` and `UDPRoute` using the new `AddressRouteMatches` type:

```go
type AddressRouteMatches struct {
	// SourceAddresses indicates the originating (source) network
	// addresses which are valid for routing traffic.
	//
	// Support: Core
	SourceAddresses []AddressMatch `json:"sourceAddresses"`

	// DestinationAddresses indicates the destination network addresses
	// which are valid for routing traffic.
	//
	// Support: Core
	DestinationAddresses []AddressMatch `json:"destinationAddresses"`
}
```

This type becomes an optional field and shared by both `TCPRouteRule` and
`UDPRouteRule` as a list:

```go
type TCPRouteRule struct {
	// Matches add rules for filtering traffic to backends based on addresses.
	//
	// +optional
	Matches []AddressRouteMatches `json:"matches,omitempty"`
}
```

Each element in `[]AddressRouteMatches` should be implemented as an `OR` style
match (e.g. the inbound traffic matches as long as at least one of the separate
`AddressRouteMatches` rules is matched).

The above would make the following YAML examples possible:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: matching
spec:
  rules:
  - matches:
    - sourceAddresses:
      - value: "192.168.1.1"
      - value: "FE80::0202:B3FF:FE1E:8329"
      - type: NamedAddress
        value: "my-custom-name"
      destinationAddresses:
      - value: "10.96.0.1"
    backendRefs:
    - name: my-service
      port: 8080
```

## Alternatives

### Pure Gateway Mode

Technically the existing specification supported this kind of matching through
`Gateway` object `Listeners` and it was considered to simply document that
further and expand upon it, but in a desire to better support more complex
setups that are becoming commonplace in the ecosystem (e.g. service mesh) there
was sufficient cause to add this functionality at the route level.

### Copying NetworkPolicy

After the first draft of this document we consulted the `NetworkPolicy` API to
determine if there were enough similarities to copy some of the semantics there
to here. Both the [existing API][k8s-net] and (at the time of writing) the
[upcoming API][pol-new] were reviewed. Ultimately some influence was taken from
`NetworkPolicyPort` to define the `PolicyMatch` structure here, but some ideas
such as binding ports and network addresses together in a single struct did not
seem necessary as the `RuleAction` present in policy did not seem applicable
for this work at the time. We may want to revisit this as the new policy work
merges and matures.

[k8s-net]:https://github.com/kubernetes/kubernetes/blob/1e6f3b5cd68049a3501782af8ff3ddd647d0b408/pkg/apis/networking/types.go#L95
[pol-new]:https://github.com/kubernetes/enhancements/pull/2522

### Port Matching

While we were able to think of some cases for port matching, the constraints of
listeners for the Gateway make it much harder to understand the value at this
stage. We're deferring port matching to focus on address matching for this
iteration so that we can come back around to it separately once we've gathered
more use case information.

### CIDR AddressType

When using `AddressType` as a component to `AddressMatch` it was considered to
add a new type `CIDRAddress` which would allow matching against an entire
subnet. This sounds good, but given a lack of concrete feedback on an
immediate need for this in the original issue [#727][issue-727] it was decided
that this could wait for now and just as easily be added later in a backwards
compatible manner.

[issue-727]:https://github.com/kubernetes-sigs/gateway-api/issues/727

## References

A related conversation in [#727][issue-727] ultimately instigated these
new requirements and may be helpful to review.

[issue-727]:https://github.com/kubernetes-sigs/gateway-api/issues/727
