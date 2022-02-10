# GEP-XXXX: GRPCRoute

* Issue: [TODO](https://github.com/kubernetes-sigs/gateway-api/issues/696)
* Status: Implementable

(See definitions in [Kubernetes KEP][kep-status].

[kep-status]: https://github.com/kubernetes/enhancements/blob/master/keps/NNNN-kep-template/kep.yaml#L9

## Goal

Add an idiomatic GRPCRoute for routing gRPC traffic.

## Non-Goals

While certain gRPC implementations support multiple transports and multiple
interface definition languages (IDLs), this proposal limits itself to
[HTTP/2](https://developers.google.com/web/fundamentals/performance/http2) as
the transport and [Protocol Buffers](https://developers.google.com/protocol-buffers)
as the IDL, which makes up the vast majority of gRPC traffic in the wild.

## Introduction

At the time of writing, the only official Route resource within the Gateway APIs
is HTTPRoute. It _is_ possible to support other protocols via CRDs and
controllers taking advantage of this have started to pop up. However, in the
long run, this leads to a fragmented ecosystem.

gRPC is a [popular RPC framework adopted widely across the industry](https://grpc.io/about/#whos-using-grpc-and-why).
The protocol is used pervasively within the Kubernetes project itself as the basis for
many interfaces, including:

- [the CSI](https://github.com/container-storage-interface/spec/blob/5b0d4540158a260cb3347ef1c87ede8600afb9bf/spec.md),
- [the CRI](https://github.com/kubernetes/cri-api/blob/49fe8b135f4556ea603b1b49470f8365b62f808e/README.md),
- [the device plugin framework](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/)

Given gRPC's importance in the application-layer networking space and to
the Kubernetes project in particular, we must ensure that the gRPC control plane
configuration landscape does not Balkanize.

### Encapsulated Network Protocols

At the time of writing, the only kind of route officially defined by the Gateway
APIs is `HttpRoute`. This GEP is novel not only in that it introduces a second
protocol to route, but also in that it introduces the first protocol
encapsulated in a protocol already supported by the API.

That is, it _is_ theoretically possible to route gRPC traffic using only `HTTPRoute`
resources, but there are several serious problems with forcing gRPC users to route traffic at
the level of HTTP. This is why we propose a new resource.

In setting this precendent, we must also introduce a coherent policy for _when_
to introduce a custom `Route` resource for an encapsulated protocol for which a
lower layer protocol already exists. We propose the following criteria for such
an addition.

- Users of the encapsulated protocol would miss out on significant conventional features from their ecosystem if forced to route at a lower layer.
- Users of the enapsulated protocol would experience a degraded user experience if forced to route at a lower layer.
- The encapsulated protocol has a significant user base, particularly in the Kubernetes community.

gRPC meets _all_ of these criteria and is therefore, we contend, a strong
candidate for inclusion in the Gateway APIs.

#### HTTP/2 Cleartext

gRPC allows HTTP/2 cleartext communication (H2C). This is conventionally deployed for
testing. Many control plane implementations do not support this by default and
would require special configuration to work properly.

#### Content-Based Routing

While not included in the scope of this initial GEP, a common use case cited for
routing gRPC is payload-aware routing. That is, routing rules which determine a
backend based on the contents of the protocol buffer payload.

#### User Experience

The user experience would also degrade significantly if forced to route at the level of HTTP.

- Encoding services and methods as URIs (an implementation detail of gRPC)
- The [Transfer Encoding header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Transfer-Encoding) for trailers
- Many features supported by HTTP/2 but not by gRPC, such as
  - Query parameters
  - Methods besides `POST`
  - CORS

### Cross Serving

A new route type (and especially one encapsulated in an already supported
protocol) raises questions about precedence of routing rules. Today, there is
already a complicated algorithm dictating which routing rule to apply to a
particular HTTP request when two `HTTPRoutes` both apply to the request's
hostname. What semantics should apply when a `GRPCRoute` and an `HTTPRoute` both apply to
a request?

We propose that, in this case, all rules defined in the `HTTPRoute` should be
applied. Only then will the `GRPCRoute` rules be evaluated. This supports the
common use case of routing gRPC traffic at a particular URI prefix to gRPC
backends while routing all other HTTP traffic to a default HTTP backend.

More generally, we propose that when a new protocol encapsulated in an already
supported one is added, if traffic applies to both a Route resource of the lower
layer and a Route resource of the higher layer, the rules of the lower layer
will be applied before the rules of the higher layer.

#### Proxyless Service Mesh

The gRPC library supports proxyless service mesh, a system by which routing
configuration is received not by an in-line proxy or sidecar proxy but by the client
itself. Eventually, `GRPCRoute` in the Gateway APIs should support this feature.
However, to date, there are no HTTP client libraries capable of participating
in a proxyless service mesh.

---

## API

The API deviates from `HTTPRoute` where it results in a better UX for gRPC
users, while mirroring it in all other cases.

### Example `GRPCRoute`

```yaml
kind: GRPCRoute
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: foo-grpcroute
spec:
  parentRefs:
  - name: my-gateway
  hostnames:
  - foo.com
  - bar.com
  rules:
  - matches:
      method:
        service: helloworld.Greeter
        method:  SayHello
      headers:
      - type: Exact
        name: magic
        value: foo

    filters:
    - type: RequestHeaderModifierFilter
      add:
        - name: my-header
          value: foo

    - type: RequestMirrorPolicyFilter
      destination:
        backendRef:
          name: mirror-svc

    backendRefs:
    - name: foo-v1
      weight: 90
    - name: foo-v2
      weight: 10
```

### Structs

```go
type GRPCRouteSpec struct {
	CommonRouteSpec `json:",inline"`

	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Hostnames []Hostname `json:"hostnames,omitempty"`

	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:default={{matches: {{method: {type: "RegularExpression", service: ".*", method: ".*"}}}}}
	Rules []GRPCRouteRule `json:"rules,omitempty"`
}

type GRPCRouteRule struct {
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{method: {type: "RegularExpression", service: ".*", method: ".*"}}}
	Matches []GRPCRouteMatch `json:"matches,omitempty"`

	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []GRPCRouteFilter `json:"filters,omitempty"`

	// Support: Core for Kubernetes Service
	// Support: Custom for any other resource
	//
	// Support for weight: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	BackendRefs []GRPCBackendRef `json:"backendRefs,omitempty"`
}

type GRPCRouteMatch struct {
	// +optional
	// +kubebuilder:default={type: "RegularExpression", service: ".*", method: ".*"}
	Method *GRPCMethodMatch `json:"path,omitempty"`

	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []GRPCHeaderMatch `json:"headers,omitempty"`
}

type GRPCMethodMatch struct {
	// Support: Core (Exact)
	//
	// Support: Custom (RegularExpression)
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *GRPCMethodMatchType `json:"type,omitempty"`

	// +optional
	// +kubebuilder:default=""
	// +kubebuilder:validation:MaxLength=1024
	Service *string `json:"value,omitempty"`

	// +optional
	// +kubebuilder:default=""
	// +kubebuilder:validation:MaxLength=1024
	Method *string `json:"value,omitempty"`

	// +optional
	// +kubebuilder:default=true
	CaseSensitive *bool `json:"value,omitempty"`
}

// +kubebuilder:validation:Enum=Exact;RegularExpression
type GRPCMethodMatchType string

type GRPCHeaderMatch struct {
	// +optional
	// +kubebuilder:default=Exact
	Type *HeaderMatchType `json:"type,omitempty"`

	Name GRPCHeaderName `json:"name"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`
}

// +kubebuilder:validation:Enum=Exact;RegularExpression
type HeaderMatchType string

// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`
type GRPCHeaderName string

type GRPCBackendRef struct {
	// +optional
	BackendRef `json:",inline"`

	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []GRPCRouteFilter `json:"filters,omitempty"`
}

type GRPCRouteFilter struct {
	// +unionDiscriminator
	// +kubebuilder:validation:Enum=RequestHeaderModifier;RequestMirror;ExtensionRef
	// <gateway:experimental:validation:Enum=RequestHeaderModifier;RequestMirror;ExtensionRef>
	Type GRPCRouteFilterType `json:"type"`

	// Support: Core
	//
	// +optional
	RequestHeaderModifier *GRPCRequestHeaderFilter `json:"requestHeaderModifier,omitempty"`

	// Support: Extended
	//
	// +optional
	RequestMirror *GRPCRequestMirrorFilter `json:"requestMirror,omitempty"`

	// Support: Implementation-specific
	//
	// +optional
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`
}

type GRPCRequestHeaderFilter struct {
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Set []GRPCHeader `json:"set,omitempty"`

	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Add []GRPCHeader `json:"add,omitempty"`

	// +optional
	// +kubebuilder:validation:MaxItems=16
	Remove []string `json:"remove,omitempty"`
}

type GRPCHeader struct {
	Name GRPCHeaderName `json:"name"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`
}

type GRPCRequestMirrorFilter struct {
	// Support: Extended for Kubernetes Service
	// Support: Custom for any other resource
	BackendRef BackendObjectReference `json:"backendRef"`
}
```
