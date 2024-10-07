# GEP-3171: Percentage-based Request Mirroring

* Issue: [#3171](https://github.com/kubernetes-sigs/gateway-api/issues/3171)
* Status: **Experimental**

(See status definitions [here](/geps/overview/#gep-states).)

## TLDR

Enhance the existing [Request Mirroring](https://gateway-api.sigs.k8s.io/guides/http-request-mirroring/) feature by allowing users to specify a percentage of requests they'd like mirrored.

## Goals

Successfully implement the feature.

## Introduction

[Request Mirroring](https://gateway-api.sigs.k8s.io/guides/http-request-mirroring/) is a feature that allows a user to mirror requests going to some backend A along to some other specified backend B. Right now Request Mirroring is an all or nothing feature – either 100% of request are mirrored, or 0% are. Percentage-based Request Mirroring will allow users to specify a percentage of requests they'd like mirrored as opposed to every single request.   

This feature is already [supported by Envoy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-routeaction-requestmirrorpolicy), so adding it for the Gateway API would enable better integration between the two products. There's also an existing user desire for this feature on the [HAProxy side](https://www.haproxy.com/blog/haproxy-traffic-mirroring-for-real-world-testing) and [NGINX side](https://alex.dzyoba.com/blog/nginx-mirror/). Since Request Mirroring is already supported by the Gateway API, Percentage-based Request Mirroring would a clear improvement on this pre-existing feature.

## Existing Support in Implementations

| Implementation | Support |
|----------------|------------|
| Envoy | [config.route.v3.RouteAction.RequestMirrorPolicy](config.route.v3.RouteAction.RequestMirrorPolicy) |
| HAProxy | [HAProxy SPOP](https://github.com/haproxytech/spoa-mirror) |
| NGINX | [ngx_http_mirror_module](https://nginx.org/en/docs/http/ngx_http_mirror_module.html) |
| gCloud | [RequestMirrorPolicy](https://cloud.google.com/python/docs/reference/compute/latest/google.cloud.compute_v1.types.RequestMirrorPolicy) |

## API

This GEP proposes the following API changes:

* Add utility type `Fraction` to [v1/shared_types.go](https://github.com/kubernetes-sigs/gateway-api/blob/cb5bf1541fa70f0692aebde8c64bba434cf331b6/apis/v1/shared_types.go):


```go
type Fraction struct {
        // +optional
        // +kubebuilder:default=100
        // +kubebuilder:validation:Minimum=0
        Numerator int32 `json:"numerator"`

        // +optional
        // +kubebuilder:default=100
        // +kubebuilder:validation:Minimum=1
        Denominator int32 `json:"denominator"`
}
```


* Update the `HTTPRequestMirrorFilter` struct to include a `Percent` field of type `int32`, and a `Fraction` field of type `Fraction`:


```go
// HTTPRequestMirrorFilter defines configuration for the RequestMirror filter.
type HTTPRequestMirrorFilter struct {
        // BackendRef references a resource where mirrored requests are sent.
        //
        // Mirrored requests must be sent only to a single destination endpoint
        // within this BackendRef, irrespective of how many endpoints are present
        // within this BackendRef.
        //
        // If the referent cannot be found, this BackendRef is invalid and must be
        // dropped from the Gateway. The controller must ensure the "ResolvedRefs"
        // condition on the Route status is set to `status: False` and not configure
        // this backend in the underlying implementation.
        //
        // If there is a cross-namespace reference to an *existing* object
        // that is not allowed by a ReferenceGrant, the controller must ensure the
        // "ResolvedRefs"  condition on the Route is set to `status: False`,
        // with the "RefNotPermitted" reason and not configure this backend in the
        // underlying implementation.
        //
        // In either error case, the Message of the `ResolvedRefs` Condition
        // should be used to provide more detail about the problem.
        //
        // Support: Extended for Kubernetes Service
        //
        // Support: Implementation-specific for any other resource
        BackendRef BackendObjectReference `json:"backendRef"`
      
        // Percent represents the percentage of requests that should be
        // mirrored to BackendRef. Its minimum value is 0 (indicating 0% of
        // requests) and its maximum value is 100 (indicating 100% of requests).
        //
        // If both Percent and Fraction are specified, Fraction will take
        // priority. If Percent is unspecified, it will have a default value of
        // 100. If Fraction is unspecified, it will have a default value of
        // 100/100. This means that if neither field is specified, 100% of
        // requests will be mirrored.
        //
        // +optional
        // +kubebuilder:default=100
        // +kubebuilder:validation:Minimum=0
        // +kubebuilder:validation:Maximum=100
        Percent int32 `json:"percent,omitempty"`

        // Fraction represents the fraction of requests that should be
        // mirrored to BackendRef.
        //
        // If both Percent and Fraction are specified, Fraction will take
        // priority. If Percent is unspecified, it will have a default value of
        // 100. If Fraction is unspecified, it will have a default value of
        // 100/100. This means that if neither field is specified, 100% of
        // requests will be mirrored.
        //
        // +optional
        Fraction Fraction `json:"fraction,omitempty"`
}
```

## Example

An example with Percent:


```
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-mirror
  labels:
    gateway: mirror-gateway
spec:
  parentRefs:
  - name: mirror-gateway
  hostnames:
  - mirror.example
  rules:
  - backendRefs:
    - name: foo-v1
      port: 8080
    filters:
    - type: RequestMirror
      requestMirror:
        backendRef:
          name: foo-v2
          port: 8080
        percent: 42
```
This would result in 42% of requests going to `foo-v1` to be mirrored to `foo-v2`.    

An example with Fraction:


```
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-mirror
  labels:
    gateway: mirror-gateway
spec:
  parentRefs:
  - name: mirror-gateway
  hostnames:
  - mirror.example
  rules:
  - backendRefs:
    - name: foo-v1
      port: 8080
    filters:
    - type: RequestMirror
      requestMirror:
        backendRef:
          name: foo-v2
          port: 8080
        fraction:
          numerator: 5
          denominator: 1000
```
This would result in 0.5% of requests going to `foo-v1` to be mirrored to `foo-v2`.
