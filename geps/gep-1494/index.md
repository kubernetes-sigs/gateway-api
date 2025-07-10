# GEP-1494: HTTP Auth in Gateway API

* Issue: [#1494](https://github.com/kubernetes-sigs/gateway-api/issues/1494)
* Status: Implementable

(See [status definitions](../overview.md#gep-states).)


## TLDR

Provide a method for configuring **Gateway API implementations** to add HTTP Authentication for north-south traffic. The method may also include Authorization config if practical. At the time of writing, this authentication is only for ingress traffic to the Gateway.


## Goals

(Using the [Gateway API Personas](../../concepts/roles-and-personas.md))

* A way for Ana the Application Developer to configure a Gateway API implementation to perform Authentication (at least), with optional Authorization.

* A way for Chihiro the Cluster Admin to configure a default Authentication and/or Authorization config for some set of HTTP or GRPC matching criteria.

## Stretch Goals

* Optionally, a way for Ana to have the ability to disable Authentication and/or Authorization for specific routes when needed, allowing certain routes to not be protected. This would probably need to work something like a default enabling at Gateway level, that can be specifically set at lower levels, but further design is TBD. This goal comes from the relatively-common desire for Chihiro to be able to set reasonably-secure defaults, and for Ana or others to be able to _disable_ for specific paths for purposes of health checks. The fact that this is relatively undefined is why this goal is _optional_.


## Non-Goals

* Handling all possible authentication and authorization schemes. Handling a (preferably large) subset of authentication and authorization is acceptable.


## Deferred Goals

* (Not performed during the Provisional Phase) Defining the API or configuration to be used.
* Handling GRPC (We will handle plain HTTP first)
* Any decisions about doing Auth for non-HTTP protocols (this is a whole other problem that could significantly impact timelines)

## Introduction

A common request for Gateway API has been a way to direct implementations to automatically request various forms of authentication and authorization. (see the [GEP Issue](https://github.com/kubernetes-sigs/gateway-api/issues/1494) for some of the history.)

**Authentication (AuthN for short)** refers to proving the requesting party's identity in some fashion.

**Authorization (AuthZ for short)** refers to making decisions about what a party can access _based on their Authenticated identity_.

In this document, we'll use Auth* to represent "AuthN and maybe AuthZ as well".

This capability is useful for both Ana (who wants to have something for AuthN and AuthZ that ensures security without it needing to be built into her app), and Chihiro (who wants to be able to ensure that the platform as a whole conforms to some given level of security).

### Common authentication methods

Before discussing any proposed Auth* solution, it's important to discuss some AuthN methods that are often used in securing modern applications.

#### Basic HTTP Auth

In Basic HTTP Auth, a server asks for authentication as part of returning a `401` status response, and the client includes an `Authorization` header that includes a Base64-encoded username and password.

Use of passwords in this way is very vulnerable to replay attacks (since if you get the password, you can use
it as many times as you like), which is why Basic Auth is generally used for lower-security use cases.

However, even in the lowest-security use cases, using TLS to at least prevent man-in-the-middle password interception is very necessary.

Basic auth is defined in [RFC-7617](https://datatracker.ietf.org/doc/html/rfc7617).

#### TLS Client Certificate Authentication

TLS includes the possibility of having both the client and server present certificates for the other party to validate. (This is often called "mutual TLS", but is distinct from the use of that term in Service Mesh contexts, where it means something more like "mutual TLS with short-lifetime, automatically created and managed dynamic keypairs for both client and server").

In this case, the server also authenticates the client, based on the certificate chain presented by the client. Some implementations also allow details about the certificate to be passed through to backend clients, to be used in authorization decisions.

TLS v1.3 is defined in [RFC-8446](https://datatracker.ietf.org/doc/html/rfc8446), with v1.2 defined in [RFC-5246](https://datatracker.ietf.org/doc/html/rfc5246). Earlier versions of TLS (v1.1 and v1.0) were deprecated in [RFC-8996](https://datatracker.ietf.org/doc/html/rfc8996) and should no longer be used.

Gateway API already has some work in progress to handle this for straightforward use cases, in [GEP-91](../gep-91/index.md).

#### JWT

JWT is a format for representing _claims_, which are stored as fields in a JSON object. In common use, these claims represent various parameters that describe the properties of an authentication, with sample claims being `iss`, Issuer, `sub`, Subject, and `aud`, Audience.

JWT also specified ways in which JWTs can be _nested_, where a JWT is encoded inside another JWT, which allows the use of encryption and signing algorithms. In the Gateway API Auth* case, this is expected to be common, in a pattern looking something like this:

- a common authority issues keypairs to clients and servers
- clients include JWTs with requests that identify themselves and their intended uses
- servers unwrap encryption and validate signatures of these JWTs to validate the chain of trust
- servers use the authentication details to make authorization decisions.

JWT is defined in [RFC-75199](https://datatracker.ietf.org/doc/html/rfc7519).

#### Oauth2 and OIDC

Oauth2 is an _authorization framework_, which allows clients and servers to define ways to perform authentication and authorization in as secure a way as possible. It extensively uses TLS for encryption, and involves a third-party handling the authorization handshake with a client, which the third-party then provides to the server.

Open ID Connect (OIDC) is a protocol based on the OAuth 2 framework, that allows Users to talk to Identity Providers (IDPs), on behalf of a Relying Party (RP), and have the IDP give the user an Identity Token (which the User's browser can then provide as Authentication to the Relying Party), and also allows the RP to request Claims about the User, which can be used for Authorization.

Usually, the Identity Token is delivered using JWT, although that is not required.

## Auth* User Stories


* As Ana the Application Developer, I wish to be able to configure that some or all of my service exposed via Gateway API requires Authentication, and ideally to be able to make Authorization decisions about _which_ authenticated clients are allowed to access.
* As Ana the Application Developer, I wish to be able to redirect users to a login page when they lack authentication, while unauthenticated API access gets the proper 40x response.
* As Chihiro the Cluster Admin, I wish to be able to configure default Authentication settings (at least), with an option to enforce Authentication settings (preferable but not required) for some set of services exposed via Gateway API inside my cluster.
* More User Stories welcomed here!

## Currently implemented Auth mechanisms in implementations

Many Gateway API implementations have implemented some method of configuring Auth* that are either delivered outside of the Gateway API framework, or if they are inside, are currently Implementation Specific.

This section lays out some examples (updates with extra examples we've missed are very welcome).

#### HTTP Authentication
| Name | External Authentication | Self Authentication |
| -------- | -------- | -------- |
| [HAProxy Ingress](https://haproxy-ingress.github.io/docs/configuration/keys/)     | [Custom](https://haproxy-ingress.github.io/docs/configuration/keys/#auth-external) , [OAuth](https://haproxy-ingress.github.io/docs/configuration/keys/#oauth)    | [Basic](https://github.com/jcmoraisjr/haproxy-ingress/tree/master/examples/auth/basic), [mTLS](https://haproxy-ingress.github.io/docs/configuration/keys/#auth-tls)     |
|[GlooEdge](https://docs.solo.io/gloo-edge/latest/guides/security/auth/extauth/)| [Envoy](https://github.com/solo-io/gloo) ( [Basic](https://docs.solo.io/gloo-edge/latest/guides/security/auth/extauth/basic_auth/), [passthrough](https://docs.solo.io/gloo-edge/latest/guides/security/auth/extauth/passthrough_auth/), [OAuth](https://docs.solo.io/gloo-edge/latest/guides/security/auth/extauth/oauth/), [ApiKey](https://docs.solo.io/gloo-edge/latest/guides/security/auth/extauth/apikey_auth/), [LDAP](https://docs.solo.io/gloo-edge/latest/guides/security/auth/extauth/ldap/), [Plugin](https://docs.solo.io/gloo-edge/latest/guides/security/auth/extauth/plugin_auth/)), [Custom](https://docs.solo.io/gloo-edge/latest/guides/security/auth/custom_auth/) | Envoy ([JWT](https://docs.solo.io/gloo-edge/latest/guides/security/auth/jwt/))
|[traefik](https://doc.traefik.io/traefik/middlewares/http/forwardauth/)|[Custom(ForwardAuth middleware)](https://doc.traefik.io/traefik/middlewares/http/forwardauth/)|[Basic](https://doc.traefik.io/traefik/middlewares/http/basicauth/), [Digest Auth](https://doc.traefik.io/traefik/middlewares/http/digestauth/)|
|[Ambassador](https://www.getambassador.io/docs/edge-stack/latest/howtos/ext-filters)|[Envoy](https://github.com/emissary-ingress/emissary) ([Basic](https://www.getambassador.io/docs/edge-stack/latest/howtos/ext-filters#2-configure-aesambassador-edge-stack-authentication))|[SSO(OAuth, OIDC)](https://www.getambassador.io/docs/edge-stack/latest/howtos/oauth-oidc-auth) |
|[ingress-nginx](https://kubernetes.github.io/ingress-nginx/examples/customization/external-auth-headers/)|[httpbin](https://httpbin.org) ([Basic](https://kubernetes.github.io/ingress-nginx/examples/auth/external-auth/), [OAuth](https://kubernetes.github.io/ingress-nginx/examples/auth/oauth-external-auth/))|[Basic](https://kubernetes.github.io/ingress-nginx/examples/auth/basic/), [Client Certificate](https://kubernetes.github.io/ingress-nginx/examples/auth/client-certs/)|
|[Envoy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authzz/v3/ext_authzz.proto)|[External Authorization server (ext_authzz filter)](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authzz/v3/ext_authzz.proto) |[JWT](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/jwt_authn/v3/config.proto)|
|[Contour](https://projectcontour.io/docs/1.24/config/client-authorization/)|[Envoy](https://projectcontour.io/docs/1.24/config/client-authorization/)|[Envoy(JWT)](https://projectcontour.io/docs/1.24/config/jwt-verification/)|
|[Istio](https://istio.io/latest/docs/tasks/security/authorization/)|[mutual TLS ingress gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-mutual-tls-ingress-gateway), [External Authorization](https://istio.io/latest/docs/tasks/security/authorization/authz-custom/)|[JWT (RequestAuthentication)](https://istio.io/latest/docs/reference/config/security/request_authentication/)|
|[Envoy Gateway](https://gateway.envoyproxy.io/docs/tasks/security/ext-auth/)| [Envoy](https://gateway.envoyproxy.io/docs/tasks/security/ext-auth/#http-external-authorization-service) | [Envoy(JWT)](https://gateway.envoyproxy.io/docs/tasks/security/jwt-authentication/), [Basic](https://gateway.envoyproxy.io/docs/tasks/security/basic-auth/) |
|[Consul](https://developer.hashicorp.com/consul/docs/connect/gateways/api-gateway/secure-traffic/verify-jwts-k8s)|[Envoy](https://developer.hashicorp.com/consul/docs/connect/proxies/envoy-extensions/configuration/ext-authz)| [JWT](https://developer.hashicorp.com/consul/docs/connect/gateways/api-gateway/secure-traffic/verify-jwts-k8s#use-jwts-to-verify-requests-to-api-gateways-on-kubernetes)|


## Outstanding Questions and Concerns (TODO)

From @ongy, some additional goals to keep in mind:
* The API of the proposed implementation provides enough flexibility to integrate with an authorization mechanism and protect resources entirely in the gateway
* The API allows to inject information about the authentication result into the requests and allows backend application to make authorization decisions based on this.

## API

This GEP proposes a two-part solution to this problem:

* We introduce a new HTTPRoute Filter, `ExternalAuth`, that allows the
  specification of an external source to connect to using Envoy's `ext_authz` protocol.
* We introduce a new Policy object that can be targeted at either the
  Gateway or HTTPRoute levels. In either of these cases, it _defaults_ the settings
  for the HTTPRoute Filter across all HTTPRoute matches that roll up to the object.

These two parts will be done in two separate changes - Filter first, then
Policy after.

Both of these additions use the same underlying struct for the config, so that
changes or additions in one place add them in the other as well.

This plan has some big things that need explaining before we get to the API details:

* Why a Filter plus Policy approach?
* Why two changes?
* Why Envoy's `ext_authz`?

### Why a Filter plus Policy approach?

We have two main requirements: Ana needs to be able to configure auth at least at
the smallest possible scope, and Ana, Ian and Chihiro need to be able to configure
defaults at larger scopes.

The smallest possible scope for this config is the HTTPRoute Rule level, where
you can match a single set of matchers - like a path, or a path and header
combination.

At this level, the inline tool we have available to perform changes is the HTTPRoute
Filter, which also has the property that it's designed to _filter_ requests. This
matches the overall pattern here, which is to _filter_ some requests, allowing
or denying them based on the presence of Authentication and the passing of
Authorization checks.

A Policy _can_ be targeted at this level, using the Route rule as a `sectionName`,
but that leaves aside that Filters are exactly designed to handle this use case.

Policy attachment includes defaulting fields like Filters in its scope already,
so we are allowed to use a combination in this way. 

Using a Filter also has the advantage that, at the tightest possible scope (the
object being defaulted) you can _explicitly_ override any configured defaults.

Using a Filter also includes ordering (because Filters are an ordered list),
although this exact behavior is currently underspecified. This change will also
need to clarify. Ordering is particularly important for Auth use cases, because
sometimes Auth will expect certain properties that may need to be introduced
by things like header modification.

Lastly, using a Filter means that, for the simplest possible case, where Ana
wants to enable Auth* for a single path, then there is only a single object to
edit, and a single place to configure.

Using a Policy for the simplest case immediately brings in all the discovery
problems that Policy entails.

There are two important caveats here that must be addressed, however:
* Firstly, whatever is in the filter either must be accepted, or the rule
  is not accepted. Overrides from anywhere else, including if we add an Override
  Policy later, must not override explicit config - that would
  violate one of the key declarative principles, that what is requested in the
  spec is either what ends up in the state, or that config is rejected.
* Secondly, filter ordering is particularly important for Auth use cases, so we
  must ensure that when we add Policy defaulting we have a way to indicate at
  what position in a filter list the Auth policy should fit.

### Why two phases?

In short: In the interest of getting something, even if incomplete, into our
user's hands as quickly as possible.

Policy design is complex, and needs to be done carefully. Doing a first
pass using only a Filter to get the basic config correct while we discuss
how to make the Policy handling work means that we can get some work out to the
community without needing to complete the whole design.

In particular, the design for the Filter plus Policy will need to answer at
least the following questions:

* How to set where in a list of Filters a defaulted Auth filter sits;
  and what happens if no Filters are specified in a HTTPRoute? Does it go first,
  last, or should there be a way to specify the order?
* What Policy types are possible? Defaults? (Definitely useful for setting a
  baseline expectation for a cluster, which is desirable for security constructs
  like Auth) Overrides? (Also very useful for ensuring that exceptions meet
  certain requirements - like only allowing the disabling of Auth on `/healthz`
  endpoints or similar use cases.)
* Should Policy have a way to describe rules around when it should take effect?
  That's in addition to the usual hierarchical rules, should the Policy have ways
  to include or exclude specific matches? This would require agreement in the
  Policy Attachment spec as well.

All of these changes have costs in complexity and troubleshooting difficulty, so
it's important to ensure that the design consciously makes these tradeoffs.

In particular, the last two items in the above list seem likely to require a fair
amount of discussion, and including a Policy in the initial release of this
seems likely to make this change miss its current release window.


### Why Envoy's ext_authz?

#### What is ext_authz?

Envoy's External Authorization filter is a filter that calls out to an authorization
service to check if the incoming request is authorized or not. Note that, in
order to check _authorization_, it must also be able to determine _authentication_ - 
this is one of the reasons why we've chosen this approach.

Envoy's implementation of this filter allows both a
[gRPC, protobuf API](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authzz/v3/ext_authzz.proto)
and configuration of a HTTP based API (which, as it's not defined using a
specification like protobuf, requires more configuration).

The important thing to remember here is that the actual authorization process
is delegated to the authorization service, and the authentication process _may
optionally_ also be delegated there - which is why the ext_authz approach allows
handling many Auth methods - most of the work is performed by external services
which implement various methods (like Basic Auth, OAuth, JWT validation, etc).

#### Why use it over other options?

The community discussed Auth extensively in-person at Kubecon London in early 2025,
and got broad agreement from multiple dataplanes that:

* something like ext_authz was a good idea, because it's flexible and allows the
  implementation of many types of Auth without specific protocol implementation
  in upstream
* Envoy's ext_authz protocol has no major problems that would stop us using it
* Envoy-based implementations mostly already have support for it

At that meeting, those present agreed that ext_authz was a good place to start.

Most non-Envoy dataplanes also already have similar methods, so the maintainers
of projects using other dataplanes were okay with this idea.

The alternative here would be to add a Filter type _per auth method_, which, given
the large number of options, could quickly become very complex.

This GEP is, however, explicitly _not_ ruling out the possibility of adding
specific Filters for specific Auth methods in the future, if users of this API
find the overhead of running a compatible implementation to be too much.

### API Design

#### Phase 1: Adding a Filter

This config mainly takes inspiration from Envoy's ext_authz filter config, while
also trying to maintain compatibility with other HTTP methods.

This design is also trying to start with a minimum feature set, and add things
as required, rather than adding everything configurable in all implementations
immediately.

There is some difference between data planes, based on the links above, but
these fields should be broadly supportable across all the listed implementations.

Some design comments are included inline.

The intent for Phase 2 is that this struct will be included in an eventual Policy
so that additions only need to be made in one place.

Additionally, as per other added Filters, the config is included in HTTPRoute,
and is not an additional CRD.

##### Go Structs

```go

// HTTPRouteExtAuthProtcol specifies what protocol should be used
// for communicating with an external authorization server.
//
// Valid values are supplied as constants below.
type HTTPRouteExtAuthProtocol string

const (
	HTTPRouteExtAuthGRPCProtocol HTTPRouteExtAuthProtocol = "GRPC"
	HTTPRouteExtAuthHTTPProtocol HTTPRouteExtAuthProtocol = "HTTP"
)
// HTTPExtAuthFilter defines a filter that modifies requests by sending
// request details to an external authorization server.
//
// Support: Extended
// Feature Name: HTTPRouteExtAuth
type HTTPExtAuthFilter struct {

	// ExtAuthProtocol describes which protocol to use when communicating with an
	// ext_authz authorization server.
	//
	// When this is set to GRPC, each backend must use the Envoy ext_authz protocol
	// on the port specified in `backendRefs`. Requests and responses are defined
	// in the protobufs explained at:
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto
	//
	// When this is set to HTTP, each backend must respond with a `200` status
    // code in on a successful authorization. Any other code is considered
	// an authorization failure.
	//
	// Feature Names:
	// GRPC Support - HTTPRouteExtAuthGRPC
	// HTTP Support - HTTPRouteExtAuthHTTP
	//
	// +unionDiscriminator
	// +kubebuilder:validation:Enum=HTTP;GRPC
	ExtAuthProtocol HTTPRouteExtAuthProtocol `json:"protocol"`

	// BackendRefs is a reference to a backend to send authorization
	// requests to.
	//
	// The backend must speak the selected protocol (GRPC or HTTP) on the
	// referenced port.
	//
	// If the backend service requires TLS, use BackendTLSPolicy to tell the
	// implementation to supply the TLS details to be used to connect to that
	// backend.
	//
	BackendRef BackendObjectReference `json:"backendRef"`

	// GRPCAuthConfig contains configuration for communication with ext_authz
	// protocol-speaking backends.
	//
	// If unset, implementations must assume the default behavior for each
	// included field is intended.
	//
	// +optional
	GRPCAuthConfig *GRPCAuthConfig `json:"grpc,omitempty"`

	// HTTPAuthConfig contains configuration for communication with HTTP-speaking
	// backends.
	//
	// If unset, implementations must assume the default behavior for each
	// included field is intended.
	//
	// +optional
	HTTPAuthConfig *HTTPAuthConfig `json:"http,omitempty"`

	// ForwardBody controls if requests to the authorization server should include
	// the body of the client request; and if so, how big that body is allowed
	// to be.
	//
	// It is expected that implementations will buffer the request body up to
	// `forwardBody.maxSize` bytes. Bodies over that size must be rejected with a
	// 4xx series error (413 or 403 are common examples), and fail processing
	// of the filter.
	//
	// If unset, or `forwardBody.maxSize` is set to `0`, then the body will not
	// be forwarded.
	//
	// Feature Name: HTTPRouteExtAuthForwardBody
	//
	// GEP Review Notes:
	// Both Envoy and Traefik show support for this feature, but HAProxy and
	// ingress-nginx do not. So this has a separate feature flag for it.
	//
	// +optional
	ForwardBody *ForwardBodyConfig `json:"forwardBody,omitempty"`
}

// GRPCAuthConfig contains configuration for communication with ext_authz
// protocol-speaking backends.
type GRPCAuthConfig struct {

	// AllowedRequestHeaders specifies what headers from the client request
	// will be sent to the authorization server.
	//
	// If this list is empty, then the following headers must be sent:
	//
	// - `Authorization`
	// - `Location`
	// - `Proxy-Authenticate`
	// - `Set-Cookie`
	// - `WWW-Authenticate`
	//
	// If the list has entries, only those entries must be sent.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=64
	AllowedRequestHeaders []string `json:"allowedHeaders,omitempty"`
}

// HTTPAuthConfig contains configuration for communication with HTTP-speaking
// backends.
type HTTPAuthConfig struct {
	// Path sets the prefix that paths from the client request will have added
	// when forwarded to the authorization server.
	//
	// When empty or unspecified, no prefix is added.
	// +optional
	Path string `json:"path,omitempty"`

	// AllowedRequestHeaders specifies what additional headers from the client request
	// will be sent to the authorization server.
	//
	// The following headers must always be sent to the authorization server,
	// regardless of this setting:
	// 
	// * `Host`
	// * `Method`
	// * `Path`
	// * `Content-Length`
	// * `Authorization`
	//
	// If this list is empty, then only those headers must be sent.
    // 
    // Note that `Content-Length` has a special behavior, in that the length
    // sent must be correct for the actual request to the external authorization
    // server - that is, it must reflect the actual number of bytes sent in the
    // body of the request to the authorization server.
    //
    // So if the `forwardBody` stanza is unset, or `forwardBody.maxSize` is set
    // to `0`, then `Content-Length` must be `0`. If `forwardBody.maxSize` is set
    // to anything other than `0`, then the `Content-Length` of the authorization
    // request must be set to the actual number of bytes forwarded.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=64
	AllowedRequestHeaders []string `json:"allowedHeaders,omitempty"`

	// AllowedResponseHeaders specifies what headers from the authorization response
	// will be copied into the request to the backend.
	//
	// If this list is empty, then all headers from the authorization server
	// except Authority or Host must be copied.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=64
	AllowedResponseHeaders []string `json:"allowedResponseHeaders,omitempty"`

}

// ForwardBody configures if requests to the authorization server should include
// the body of the client request; and if so, how big that body is allowed
// to be.
//
// If empty or unset, do not forward the body.
type ForwardBodyConfig struct {

	// MaxSize specifies how large in bytes the largest body that will be buffered
    // and sent to the authorization server. If the body size is larger than
    // `maxSize`, then the body sent to the authorization server must be
    // truncated to `maxSize` bytes.
	//
	// If 0, the body will not be sent to the authorization server.
	MaxSize uint16 `json:"maxSize,omitempty"`
}

```
#### YAML Examples

Coming soon.

#### Phase 2: Adding more complex configuration with Policy

This phase is currently undefined until we reach agreement on the Filter + Policy
approach.

## Conformance Details

(from https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-2162/index.md#standardize-features-and-conformance-tests-names)

#### Feature Names

For this feature as a base:

`HTTPRouteExtAuth`

For supporting talking to ext_authz servers using the gRPC ext_authz protocol:

`HTTPRouteExtAuthGRPC`

For supporting talking to ext_authz servers using HTTP:

`HTTPRouteExtAuthHTTP`

For forwarding the body of the client request to the authorization server

`HTTPRouteExtAuthForwardBody`


### Conformance tests 

Conformance tests file names should try to follow the `pascal-case-name.go` format.
For example for `HTTPRoutePortRedirect` - the test file would be `httproute-port-redirect.go`.
Treat this guidance as "best effort" because we might have test files that check the combination of several features and can't follow the same format.
In any case, the conformance tests file names should be meaningful and easy to understand.

## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References
