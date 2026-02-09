# Hostnames in Gateway API

## Introduction/Purpose of this document

This document is intended to help both users of Gateway API and integrators who build systems that programmatically interact with Gateway API objects to better understand how Gateway API uses hostnames, and what are the most important things to know about these usages.

## Where and how can you configure a hostname?

There are a number of places you can configure a hostname in Gateway API, each of which with a different purpose. They are used for a number of purposes, including both **hostname intersection** between Routes and Gateways, which is used to determine if the Routes can attach, and also as a **routing discriminator**, to choose which Listener and Route should accept a particular request (more on both of these later in this document).

Each `hostname` field can accept either _precise_ hostnames (that is, a hostname like `www.example.com`), or _wildcard_ hostnames (that is, a hostname like `*.example.com`). The level of precision in a hostname also affects its effective order in the process of choosing which Listener will match particular traffic, with more precise beating less precise hostnames.

Note that IP addresses are _never_ valid hostnames in Gateway API, although, at the time of writing, the validation for those fields may allow them. This is a bug and will be fixed in the future. Gateway API **strongly** recommends not depending on this behavior.

??? example "Hostname Type details"

	There are actually two types of Hostname available in Gateway API - `Hostname` and `PreciseHostname`. `Hostname` has the behavior described below with wildcards, but `PreciseHostname` does _not_ allow wildcards. Otherwise the two are the same.

### Hostname wildcards

In Gateway API, a wildcard is supported _only_ as the leftmost character in the hostname, and must be immediately followed by a `.` (which further means that the wildcard will only match complete DNS Labels, as defined in the DNS RFCs like [RFC-9499](https://www.rfc-editor.org/rfc/rfc9499.html), [RFC-2308](https://www.rfc-editor.org/rfc/rfc2308), and others).

For example:
- `*.example.com` is a valid `hostname`

- `f*.example.com` is not a valid `hostname`, as the wildcard character `*` is not the leftmost character.
- `*oo.example.com` is also not a valid `hostname`, as the wildcard character is not followed by a `.`.

Additionally - and unlike many other systems - the wildcard is defined as matching _one or more_ DNS Labels, rather than only one. For example:

- `*.example.com` matches `www.example.com` and `sub.domain.example.com`, but not `example.com`.
- `*.com` matches  `example.com`, and also `www.example.com`.

This is important in both hostname intersection and routing discrimination.

### Available `hostname` fields

#### Listener (avilable in Gateway and ListenerSet)

In a Gateway or a ListenerSet, the Listener stanza includes a `hostname` field. Each Listener can have up to one `hostname`, although the `hostname`s can include wildcards. When the `hostname` field is not specified, then any hostname will match for both hostname intersection and routing discrimination.

ListenerSet is a relatively new resource that provides a way for users who do not own Gateways to introduce additional Listeners to a Gateway. In order to do this, it includes a Listener stanza like the Gateway object.

#### Routes (HTTPRoute, GRPCRoute, and TLSRoute)

Some Routes have a `hostnames` field. This field is treated as an `OR` for both hostname intersection and routing discrimination. `hostnames` may include wildcards.

The `hostname` field is optional, and when not supplied, any hostname will match the Route, for both hostname intersection and routing discrimination.

## What do the various hostnames actually do?

### Route attachment

**Route Attachment** is the process by which Routes and Gateways agree on whether a Route may attach to a Gateway. Routes specify a ParentRef, which may be a Gateway or a ListenerSet, and Gateways and ListenerSets may specify AllowedRoutes, which can choose which Route Kinds are allowed, or what namespaces the Routes can be in.

Most importantly for this document, Gateway Listeners and some types of Route both include `hostname` fields, and those two fields must **intersect** correctly for the Route to be **Accepted**, and attach to the Gateway. Routes which do not attach to a Gateway ParentRef do not do anything for that Gateway. So getting this right is important!

This process is referred to as **hostname intersection**, and it works regardless of which Route type you are talking about.

#### Hostname Intersection

In **hostname intersection**, the `hostname` fields on both a Listener and a Route are considered, and, if those hostnames overlap, then the intersection is a success, and the Listener allows the Route, subject to other Listener requirements.

This intersection has some rules:

* if both hostnames are **precise** (containing no wildcards) then the hostnames must match exactly. So, `foo.example.com` on the Listener, and `foo.example.com` on a HTTPRoute intersect.
* If the Listener has a wildcard hostname, and the Route has a precise hostname that matches that wildcard, then they intersect. So, `*.example.com` on the Listener and `www.example.com` or `sub.domain.example.com` on a HTTPRoute will intersect.
* If the Listener has a precise hostname, and the Route has a wildcard hostname that matches the precise hostname, then they intersect. So, `www.example.com` or `sub.domain.example.com` on the Listener, and `*.example.com` on a HTTPRoute intersect.
* If both Listener and Route have wildcard hostnames, they intersect as long as they overlap. So, `*.example.com` and `*.example.com` intersect, but so do `*.example.com` and `*.com`.
* The special wildcard `*` (with no other characters), matches any other hostname for hostname intersection purposes.
* Having no hostname set is equivalent to having a `hostname` field set to `*`.

If the hostnames intersect following these rules, then the attachment between the Listener and Route can proceed (assuming all other requirements are also successful).

The hostname that actually intersects is referred to as the **intersected** hostname. This is important for traffic and routing discrimination.

Some examples:

| Listener `hostname` | Route `hostname` | Intersected `hostname` |
|---------------------|------------------|--------------------|
| `www.example.com` | `www.example.com` | `www.example.com`  |
| `*.example.com` | `www.example.com` | `www.example.com` |
| `*.example.com` | `sub.domain.example.com` | `sub.domain.example.com` |
| `www.example.com` | `*.example.com` | `www.example.com` |
| `sub.domain.example.com` | `*.example.com` | `sub.domain.example.com` |
| `*.example.com` | `*.example.com` | `*.example.com` |
| `*.com` | `*.example.com` | `*.example.com` |
| `*` | `www.example.com` | `www.example.com` |
| `*` | `*` | `*` |


### Traffic and Routing discriminators

Hostnames are also used for _traffic_ and _routing_ discrimination. To put it another way, they are used to choose where traffic will be routed, whether that is choosing a Listener out of a set of Listeners, or choosing a Route out of the set of Routes attached to a Listener.

#### Listener ordering

When a Gateway has a set of Listeners that contains Listeners that have the same `port` and `protocol`, but different `hostname`s, then the Gateway is expected to send traffic that _could_ match multiple Listeners to the _most specific_ Listener.

Note that for this process, only the **intersected hostname** (the result of the hostname intersection calculation) is relevant.

This is important when considering `hostname`, because wildcards create a hierarchy of specificity. That is, a Listener with a hostname that contains a wildcard is _less precise_ and _less specific_ than one that only contains a precise hostname.

Broadly speaking, a hostname is more specific if it has _more_ labels that do not contain a wildcard than another.

Some examples:

* `www.example.com` (3 specific labels) is more specific than `*.example.com` (2 specific labels).
* `*.example.com` (2 specific labels) is more specific than `*.com` (1 specific label).
* `*.com` (1 specific label) is more specific than `*` (0 specific labels).

When choosing a Listener to accept, the exact hostname details to match depend on the protocol, but, all follow a general pattern:

* exact matches
* most specific wildcard match to request hostname
* general wildcard match to request hostname

#### SNI matching

For `protocol` values that use TLS, the intersected hostname is expected to match multiple details:

* The intersected hostname must be present on a certificate that is used for TLS termination, either in the CN or SAN fields, when the `tls.mode` is set to `Terminate`.
* A TLS request that arrives at a `HTTPS` or `TLS` listener must have a matching Server Name Indicator (SNI).

Note that Gateway API does _not_ require implementations to verify certificates used in connections on Listeners that have `tls.mode` set to `Terminate` have the intersected hostname present in those fields. (Implementations _may_ do that, but they are not required to).

For SNI matching, the "matching" part means that the SNI hostname must match the intersected hostname using the rules given in [RFC-2818](https://datatracker.ietf.org/doc/html/rfc2818#section-3.1) for Server Identity matching. Note also that an SNI may not include a wildcard, it must be a precise hostname (in Gateway API terms).

Quoted from RFC-2818:
> If more than one identity of a given type is present in the certificate (e.g., more than one dNSName name), a match in any one of the set is considered acceptable. Names may contain the wildcard character * which is considered to match any single domain name component or component fragment. E.g., `*.a.com` matches `foo.a.com` but not `bar.foo.a.com`. 

Because some values that are valid in certificates are not valid Gateway API `hostname`s, some matches are not possible - for example, `f*.com` is not a valid `hostname`, so cannot match `foo.com` as in RFC-2818.

Also note that RFC-2818 has the wildcard character `*` only match a _single_ DNS label, rather than multiple. So the SNI behavior is subtly different to the hostname intersection and Listener selection behavior.

Additionally, IP addresses are not valid `hostname` values for Gateway API, so they cannot match.

Examples per RFC-2818 SNI matching rules:

| Intersected hostname | Request SNI | Match |
|---|---|---|
| `www.example.com` | `www.example.com` | ✅ |
| `www.example.com` | `foo.example.com` | ❌ |
| `*.example.com` | `www.example.com` | ✅ |
| `*.example.com` | `foo.example.com` | ✅ |
| `*.example.com` | `foo.bar.example.com` | ❌ |

Note that SNI matching is relevant for the following cases:

* HTTPRoute with Listener `protocol` `HTTPS` or `TLS` and `tls.mode` `Terminate`.
* GRPCRoute with Listener `protocol` `HTTPS` or `TLS` and `tls.mode` `Terminate`.
* TLSRoute with Listener `protocol` `TLS` and `tls.mode` `Passthrough`.
* TLSRoute with Listener `protocol` `TLS` and `tls.mode` `Terminate`.


#### `Host` header matching

For `protocol` and Route combinations that use unencrypted HTTP connection metadata (that is, HTTPRoute and GRPCRoute),
it is also required that the `Host` or `:authority` header matches the intersected hostname. Similarly to SNI matching, a `Host` header must be a precise hostname in Gateway API terms, so the matching here is similar to the Listener Selection matching:

| Intersected hostname | `Host` header | Match |
|---|---|---|
| `www.example.com` | `www.example.com` | ✅ |
| `www.example.com` | `foo.example.com` | ❌ |
| `*.example.com` | `www.example.com` | ✅ |
| `*.example.com` | `foo.example.com` | ✅ |
| `*.example.com` | `foo.bar.example.com` | ✅ |

Note that `*` can match more than one label for `Host` header matching, _not_ a single DNS label as for SNI matching.

### Expected match examples

| Listener Hostname | TLS Mode| Route Type | Route hostname | Attached? | Intersected Hostname | SNI | SNI Match? | Host header | Host header match? | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
|`www.example.com` | None | HTTPRoute | `www.example.com` | ✅ | `www.example.com` | | | `www.example.com` | ✅ ||
|`*.example.com` | None | HTTPRoute | `www.example.com` | ✅ | `www.example.com` | | | `www.example.com` | ✅ ||
|`*.example.com` | None | HTTPRoute | `*.com` | ✅ | `*.example.com` | | | `www.example.com` | ✅ ||
|`*.example.com` | None | HTTPRoute | `*.com` | ✅ | `*.example.com` | | | `foo.bar.example.com` | ✅ | Wildcard matches for Host header will match one _or more_ DNS labels.|
|`*.example.com` | None | HTTPRoute | `www.example.com` | ✅ | `www.example.com` | | | `example.com` | ❌ ||
|`*.example.com` | None | HTTPRoute | `www.example.com` | ✅ | `www.example.com` | | | `foo.example.com` | ❌ ||
| `*.example.com` | Terminated | HTTPRoute | `www.example.com` | ✅ | `www.example.com` |`www.example.com` | ✅ | `www.example.com` | ✅ ||
| `*.example.com` | Terminated | HTTPRoute | `foo.bar.example.com` | ✅ | `foo.bar.example.com` |`foo.bar.example.com` | ✅ | `foo.bar.example.com` | ✅ | SSL Certificate _must_ match **intersected hostname**, not Listener hostname, or else SNI matching will fail, because `*.example.com` on a certificate does _not_ match `foo.bar.example.com` as an SNI. |
| `*.example.com` | Terminated | HTTPRoute | `*.example.com` | ✅ | `*.example.com` |`foo.bar.example.com` | ❌ | || `*.example.com` on a certificate does _not_ match `foo.bar.example.com` as an SNI. |
| `*.example.com` | Terminated | HTTPRoute | `foo.example.com` | ✅ | `foo.example.com` |`foo.example.com` | ✅ | `foo.example.com` | ✅ ||
| `www.example.com` | Passthrough | TLSRoute | `www.example.com` | ✅ | `www.example.com` | `www.example.com`  | ✅ | |||
| `*.example.com` | Passthrough | TLSRoute | `www.example.com` | ✅ | `www.example.com` | `www.example.com`  | ✅ | |||
| `*.example.com` | Passthrough | TLSRoute | `www.example.com` | ✅ | `www.example.com` | `foo.example.com`  |  ❌  | |||
| `*.example.com` | Passthrough | TLSRoute | `foo.bar.example.com` | ✅ | `foo.bar.example.com` | `www.example.com`  | ❌ | ||The SNI must match the intersected hostname.|
| `*.example.com` | Passthrough | TLSRoute | unset | ✅ | `*.example.com` | `www.example.com`  | ✅ | |||
| `*.example.com` | Passthrough | TLSRoute | unset | ✅ | `*.example.com` | `foo.bar.example.com`  | ❌ | ||SNI matches against wildcard names in certificates can _only_ match a single DNS label. (This assumes that the certificate name matches the intersected hostname, which is not required.)|
| `www.example.com` | Terminated | TLSRoute | `www.example.com` | ✅ | `www.example.com` | `www.example.com`  | ✅ | ||The examples for TLSRoute in Terminated mode are the same as the examples for TLSRoute in Passthrough mode and have been elided.|


GRPCRoute behaves the same as HTTPRoute.


## Programmatic use of the `hostname` field 

All of this detail on hostname use in Gateway API has some effects on Gateway API integrations that wish to use hostname fields in programmatic ways (such as provisioning certificates for TLS, or DNS records for Gateways.)

Because of the combination of all the above rules, there is one absolute invariant:

**Traffic of any protocol that supports hostnames MUST be able to be accepted for the _intersected hostname_.**

That is, using either the `hostname` from the Gateway, ListenerSet, or Route in isolation is not guaranteed to be correct. Integrations may be right in _most_ cases by doing that, but it is not guaranteed. To be sure of correctness, Integrations must consider the **intersected hostname**, the result of the hostname intersection process, and how it interacts with the way the integration is using that hostname.

A negative corollary to that rule is:

**If no _intersected hostname_ can be determined, then Integrations MUST ignore that Listener and/or those attached Routes**.

The intersected hostname is the canonical representation of required hostnames for any particular Listener and all its attached Routes, so if no intersected hostname can be determined, it's not correct to do anything with that Listener. There's simply not enough information to ensure that an integration will do the correct thing.

These rules apply whether the Listener in question is inside a Gateway _or_ a ListenerSet object.

Some common examples, with recommendations, are below.

### Automatic Provisioning of DNS records

The main rule in this case is straightforward:

**All _intersected hostnames_ represented in all Listeners on the Gateway MUST resolve to all addresses in that Gateway's `status.addresses`**.

Exactly how this requirement is met is up to the integration.

Some example configs, along with _example_ ways they could be handled:

* Listener `hostname`: `*.example.com`
* HTTPRoute `hostname`s, on separate HTTPRoutes: `foo.example.com`, `bar.example.com`, `baz.quux.example.com`.
* Gateway Addresses: `192.168.0.1`, `192.168.0.2`.

In this case, the imperative result is that queries for any of the hostnames on the HTTPRoutes resolve to either `192.168.0.1`, `192.168.0.2`, or more likely, swap between both.

This could be achieved with setups like:

* Individual A records for `foo.example.com`, `bar.example.com`, and `baz.quux.example.com`, each pointing to both addresses `192.168.0.1` and `192.168.0.2`.
* An A record for `gateway-name.example.com`, pointing to both `192.168.0.1` and `192.168.0.2`, then CNAMEs for `foo.example.com`, `bar.example.com`, and `baz.quux.example.com`
* A wildcard A `*.example.com` record pointing to both `192.168.0.1` and `192.168.0.2`, if the authoritative DNS server supports that. Note that in this case, traffic to any hostname that is _not_ `foo.example.com`, `bar.example.com`, or `baz.quux.example.com` is expected to be denied by the Gateway API implementation actually serving the Gateway.
* Anything else that results in specific hostname requests resolving to the correct addresses.

Note a couple of things:

* `quux.example.com` is _not_ included, and, similarly to the wildcard case, will be rejected by the underlying Gateway API implementation even if it _does_ resolve to the underlying addresses. Gateway API says that implementations SHOULD NOT create intermediate records like this if it can be avoided.


### Automatic Provisioning of TLS Certificates

Automatic provisioning of TLS Certificates is a little more complex than DNS provisioning, because of the subtle difference between wildcard match definitions. This really only affects wildcard certificate generation.

The main rule for this is:

**Every intersected hostname on the Listener must be represented in a generated certificate used on that Listener.**

In the simplest cases, this means that every intersected hostname that rolls up to a Listener must be listed in the CN or SAN fields of a generated certificate that is to be attached to that Listener.

For the simple example:


* Listener `hostname`: `*.example.com`
* HTTPRoute `hostname`s, on separate HTTPRoutes: `foo.example.com`, `bar.example.com`, `baz.quux.example.com`.
* Gateway Addresses: `192.168.0.1`, `192.168.0.2`.

Any generated certificate present on the Listener MUST have the hostnames `foo.example.com`, `bar.example.com`, and `baz.quux.example.com` represented in the generated certificate, in either the CN or SAN fields.

Similarly to the DNS provisioning case, `quux.example.com` is not represented and SHOULD NOT be included.

#### Handling Wildcard certificates

Contrary to what some summaries of standards like OWASP indicate, wildcard certificates _can_ be used in a reasonably safe way. But, programmatically generating wildcard certificates without administrator intervention is very rarely a good idea, so the position of Gateway API is this:

**Integrations MUST NOT use the `hostname` field to programmatically generate wildcard certificates**.

To put this another way, if the intersected hostname includes a wildcard character, then TLS Certificate integrations MUST ignore it.

Gateway API is working on a guide on how the API's design intends for wildcard certificates to be managed.
