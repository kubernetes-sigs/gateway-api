# Gateway API Glossary

### Consumer Route

A Route bound to a workload's Service by a consumer of a given workload,
refining the specific consumer's use of the workload.

### Gateway Controller

A _gateway controller_ is software that runs on a cluster to manage the
infrastructure associated with routing traffic from outside the cluster to
inside the cluster (north/south traffic), analogous to the earlier _ingress
controller_ concept.

### East/West traffic

Traffic from workload to workload within a cluster.

### North/South traffic

Traffic from outside a cluster to inside a cluster (and vice versa).

### Producer Route

A Route bound to a workload's Service by the creator of a given workload,
defining what is acceptable use of the workload. Producer routes must always
be in the same Namespace as their workload's Service.

### Service mesh

A _service mesh_ is software that runs on a cluster to manage infrastructure
providing security, reliability, and observability for communications between
workloads (east/west traffic).

### Service backend

The part of a Kubernetes Service resource that is a set of endpoints
associated with Pods and their IPs. Some east/west traffic happens by having
workloads direct requests to specific endpoints within a Service backend.

### Service frontend

The part of a Kubernetes Service resource that allocates a DNS record and a
cluster IP. East/west traffic often - but not always - works by having
workloads direct requests to a Service frontend.

### Service routing

_Service routing_ is sending requests to a specific Service to the service
frontend, allowing the underlying network infrastructure (usually `kube-proxy`
or a [service mesh](#service-mesh)) to choose the specific endpoint to which
the request is routed.

### Endpoint routing

_Endpoint routing_ is sending requests to a specific Service directly to one
of the endpoints of the Service backend, bypassing routing decisions which
might be made by the underlying network infrastructure. This is commonly
necessary for advanced routing cases like sticky sessions, where the gateway
will need to guarantee that every request for a specific session goes to the
same endpoint.

### Workload

An instance of computation that provides a function within a cluster. A
workload usually comprises a Service, some Pods, and often a Deployment.