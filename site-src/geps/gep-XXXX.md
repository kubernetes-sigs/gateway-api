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
        grpcService: helloworld.Greeter
        grpcMethod:  SayHello
      headers:
      - type: Exact
        name: magic
        value: foo

    filters:
    - type: GRPCRequestHeaderModifierFilter
      add:
        - name: my-header
          value: foo

    - type: GRPCRequestMirrorPolicyFilter
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
