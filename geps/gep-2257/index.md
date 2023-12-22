# GEP-2257: Gateway API Duration Format

* Issue: [#2257](https://github.com/kubernetes-sigs/gateway-api/issues/2257)
* Status: Experimental

## TL;DR

As we extend the Gateway API to have more functionality, we need a standard
way to represent duration values. The first instance is the [HTTPRoute
Timeouts GEP][GEP-1742]; doubtless others will arise.

[GEP-1742]:/geps/gep-1742

## Gateway API Duration Format

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
interpreted as described in [RFC 8174].

[RFC 8174]:https://datatracker.ietf.org/doc/html/rfc8174

A _Gateway API Duration_ or _GEP-2257 Duration_ is a string value that:

- MUST match the regular expression `^([0-9]{1,5}(h|m|s|ms)){1,4}$` and
- MUST be interpreted as specified by [Golang's `time.ParseDuration`][gotime].

Since both of these conditions MUST be true, the effect is that GEP-2257
Durations are a subset of what `time.ParseDuration` supports:

- A GEP-2257 Duration MUST be one to four _components_, each of which consists
  of an integer _value_ followed immediately by a _unit_. For example, `1h`,
  `5m`, `1h5m`, and `1h30m30s500ms` are all valid GEP-2257 Durations.

- For each component, the value MUST be one to five decimal digits. Floating
  point is not allowed. Leading zeroes do not mean octal; the value MUST
  always be interpreted as a decimal integer. For example, `1h`, `60m`, `01h`,
  and `00060m` are all equivalent GEP-2257 Durations.

- For each component, the unit MUST be one of `h` (hour), `m` (minute), `s`
  (second), or `ms` (millisecond). No units larger than hours or smaller than
  milliseconds are supported.

- The total duration expressed by a GEP-2257 Duration string is the sum of
  each of its components. For example, `1h30m` would be a 90-minute duration,
  and `1s500ms` would be a 1.5-second duration.

- There is no requirement that all units must be used. A GEP-2257 Duration of
  `1h500ms` is supported (although probably not terribly useful).

- Units MAY be repeated, although users SHOULD NOT rely on this support since
  this GEP is Experimental, and future revisions may remove support for
  repeated units. If units are repeated, the total duration remains the sum of
  all components: a GEP-2257 duration of `1h2h20m10m` is a duration of 3 hours
  30 minutes.

- Since the value and the unit are both required within a component, `0` is
  not a valid GEP-2257 duration string (though `0s` is). Likewise the empty
  string is not a valid GEP-2257 duration.

- Users SHOULD represent the zero duration as `0s`, although they MAY use any
  of `0h`, `0m`, etc. Implementations formatting a GEP-2257 Duration for
  output MUST render the zero duration as `0s`.

- The “standard” form of a GEP-2257 Duration uses descending, nonrepeating
  units, using the largest unit possible for each component (so `1h` rather
  than `60m` or `30m1800s`, and `1h30m` rather than either `90m` or `30m1h`).
  Users SHOULD use this standard form when writing GEP-2257 Durations.
  Implementations formatting GEP-2257 Durations MUST render them using this
  standard form.

    - **Note**: Implementations of Kubernetes APIs MUST NOT modify user input.
      For example, implementations MUST NOT normalize `30m1800s` to `1h` in a
      CRD `spec`. This "standard form" requirement is limited to instances
      where an implementation needs to format a GEP-2257 Duration value for
      output.

A GEP-2257 Duration parser can be easily implemented by doing a regex-match
check before calling a parser equivalent to Go's `time.ParseDuration`. Such
parsers are readily available in (at least) [Go itself][gotime], [Rust's
`kube_core` crate from `kube-rs`][kube-core], and [Python's
`durationpy`][durationpy] package. We expect that these three languages cover
the vast majority of the Kubernetes ecosystem.

[gotime]:https://pkg.go.dev/time#ParseDuration
[kube-core]:https://docs.rs/kube-core/latest/kube_core/duration/struct.Duration.html
[durationpy]:https://github.com/icholy/durationpy

## Alternatives

We considered three main alternatives:

- Raw Golang `time.ParseDuration` format. This is very widely used in the Go
  ecosystem -- however, it is a very open-ended specification and, in
  particular, its support for floating-point values and negative durations
  makes it difficult to validate.

- Golang `strfmt.ParseDuration` as used in the APIServer's OpenAPI validation
  code. It turns out that `strfmt.ParseDuration` is a superset of
  `time.ParseDuration`, so all the problems in validation are still present.
  Additionally, `strfmt.ParseDuration` supports day and week units, requiring
  discussion of leap seconds.

- ISO8601/RFC3339 durations. These are considerably less user-friendly than
  our proposal: `PT0.5S` is simply not as immediately clear as "500ms".

There is (a lot) more discussion in [PR 2155].

[PR 2155]:https://github.com/kubernetes-sigs/gateway-api/pull/2155

## Graduation Criteria

To graduate GEP-2257 to Standard channel, we need to meet the following
criteria:

- Publish a set of test vectors for the GEP-2257 duration format.

- Have Go, Rust, and Python implementations of the parser, with a test suite
  covering all the test vectors.

- Have a custom CEL validator for GEP-2257 Duration fields.

- Have support for GEP-2257 Durations in standard Kubernetes libraries.
