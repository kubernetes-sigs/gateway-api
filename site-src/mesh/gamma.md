# The GAMMA Initiative (Gateway API for Service Mesh)

Gateway API was originally designed to manage traffic from clients outside the
cluster to services inside the cluster -- the _ingress_ or
[_north/south_][north/south traffic] case. Over time, though, interest from
[service mesh] users prompted the creation of the GAMMA (**G**ateway **A**PI for
**M**esh **M**anagement and **A**dministration) initiative in 2022 to define how
Gateway API could also be used for inter-service or [_east/west_
traffic][east/west traffic] within the same cluster.

The GAMMA initiative is a dedicated workstream within the Gateway API
subproject, rather than being a separate subproject. GAMMA’s goal is to define
how Gateway API can be used to configure a service mesh, with the intention of
making minimal changes to Gateway API and always preserving the [role-oriented]
nature of Gateway API. Additionally, we strive to advocate for consistency
between implementations of Gateway API by service mesh projects, regardless of
their technology stack or proxy.

## Deliverables

The work of the GAMMA initiative will be captured in [Gateway Enhancement
Proposals][geps] that extend or refine the Gateway API specification to cover
mesh and mesh-adjacent use cases. To date, these have been relatively small
changes (albeit sometimes with relatively large impacts!) and we expect that to
continue. Governance of the Gateway API specification remains solely with the
maintainers of the Gateway API subproject.

The ideal final outcome of the GAMMA initiative is that service mesh use cases
become a first-party concern of Gateway API, at which point there will be no
further need for a separate initiative.

## Contributing

We welcome contributors of all levels! There are [many ways to
contribute][contributor-ladder] to Gateway API and GAMMA, both technical and
non-technical.

The simplest way to get started is to attend one of GAMMA's regular meetings,
which happen every two weeks on Tuesdays for 1 hour (as found on the
[sig-network calendar]), alternating between 3pm PT and 8am PT slots to try to
be as time-zone inclusive as possible. GAMMA meetings will be moderated by the
[GAMMA leads] with notes taken by a volunteer. Community members should feel
free to attend both GAMMA and Gateway API meetings, but are by no means
obligated to do so.

[contributor-ladder]:/contributing/contributor-ladder
[east/west traffic]:/concepts/glossary#eastwest-traffic
[GAMMA
    leads]:https://github.com/kubernetes-sigs/gateway-api/blob/main/OWNERS_ALIASES#L23
[geps]:/geps/overview
[north/south traffic]:/concepts/glossary#northsouth-traffic
[service mesh]:/concepts/glossary#service-mesh
[sig-network calendar]:/contributing/community/#meetings
[role-oriented]:/concepts/roles-and-personas
