# GEP-1494: HTTP Auth in Gateway API

* Issue: [#1494](https://github.com/kubernetes-sigs/gateway-api/issues/1494)
* Status: Provisional

(See status definitions [here](/geps/overview/#gep-states).)


## TLDR

Provide a method for configuring **Gateway API implementations** to add HTTP Authentication for north-south traffic. The method may also include Authorization config if practical. At the time of writing, this authentication is only for ingress traffic to the Gateway.


## Goals

(Using the [Gateway API Personas](https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/))

* A way for Ana the Application Developer to configure a Gateway API implementation to perform Authentication (at least), with optional Authorization.

* A way for Chihiro the Cluster Admin to configure a default Authentication and/or Authorization config for some set of HTTP or GRPC matching criteria.

## Stretch Goals

* Optionally, a way for Ana to have the ability to disable Authentication and/or Authorization for specific routes when needed, allowing certain routes to not be protected. This would probably need to work something like a default enabling at Gateway level, that can be specifically set at lower levels, but further design is TBD. This goal comes from the relatively-common desire for Chihiro to be able to set reasonably-secure defaults, and for Ana or others to be able to _disable_ for specific paths for purposes of healthchecking. The fact that this is relatively undefined is why this goal is _optional_.


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

Gateway API already has some work in progress to handle this for straightforward use cases, in [GEP-91](https://gateway-api.sigs.k8s.io/geps/gep-91/).

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

Open ID Conect (OIDC) is a protocol based on the OAuth 2 framework, that allows Users to talk to Identity Providers (IDPs), on behalf of a Relying Party (RP), and have the IDP give the user an Identity Token (which the User's browser can then provide as Authentication to the Relying Party), and also allows the RP to request Claims about the User, which can be used for Authorization.

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
|[Envoy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto)|[External Authorization server (ext_authz filter)](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto) |[JWT](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/jwt_authn/v3/config.proto)|
|[Contour](https://projectcontour.io/docs/1.24/config/client-authorization/)|[Envoy](https://projectcontour.io/docs/1.24/config/client-authorization/)|[Envoy(JWT)](https://projectcontour.io/docs/1.24/config/jwt-verification/)|
|[Istio](https://istio.io/latest/docs/tasks/security/authorization/)|[mutual TLS ingress gateway](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-mutual-tls-ingress-gateway), [External Authorization](https://istio.io/latest/docs/tasks/security/authorization/authz-custom/)|[JWT (RequestAuthentication)](https://istio.io/latest/docs/reference/config/security/request_authentication/)|
|[Envoy Gateway](https://gateway.envoyproxy.io/docs/tasks/security/ext-auth/)| [Envoy](https://gateway.envoyproxy.io/docs/tasks/security/ext-auth/#http-external-authorization-service) | [Envoy(JWT)](https://gateway.envoyproxy.io/docs/tasks/security/jwt-authentication/), [Basic](https://gateway.envoyproxy.io/docs/tasks/security/basic-auth/) |
|[Consul](https://developer.hashicorp.com/consul/docs/connect/gateways/api-gateway/secure-traffic/verify-jwts-k8s)|[Envoy](https://developer.hashicorp.com/consul/docs/connect/proxies/envoy-extensions/configuration/ext-authz)| [JWT](https://developer.hashicorp.com/consul/docs/connect/gateways/api-gateway/secure-traffic/verify-jwts-k8s#use-jwts-to-verify-requests-to-api-gateways-on-kubernetes)|


## Outstanding Questions and Concerns (TODO)

From @ongy, some additional goals to keep in mind:
* The API of the proposed implementation provides enough flexibility to integrate with an authorization mechanism and protect resources entirely in the gateway
* The API allows to inject information about the authentication result into the requests and allows backend application to make authorization decisions based on this.

## API

(... details, can point to PR with changes)

## Conformance Details

(from https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-2162/index.md#standardize-features-and-conformance-tests-names)

#### Feature Names

Every feature should:

1. Start with the resource name. i.e HTTPRouteXXX
2. Follow the PascalCase convention. Note that the resource name in the string should come as is and not be converted to PascalCase, i.e HTTPRoutePortRedirect and not HttpRoutePortRedirect.
3. Not exceed 128 characters.
4. Contain only letters and numbers

### Conformance tests 

Conformance tests file names should try to follow the the `pascal-case-name.go` format.
For example for `HTTPRoutePortRedirect` - the test file would be `httproute-port-redirect.go`.
Treat this guidance as "best effort" because we might have test files that check the combination of several features and can't follow the same format.
In any case, the conformance tests file names should be meaningful and easy to understand.

## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References
