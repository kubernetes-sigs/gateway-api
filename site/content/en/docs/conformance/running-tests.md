---
title: "Running the suite"
weight: 100
---

### Running Tests

There are two main contrasting sets of conformance tests:

* Gateway related tests (can also be thought of as ingress tests)
* Service Mesh related tests

By default, the suite infers supported features from GatewayClass and XMesh
status. To select features manually for development, use the
`-supported-features` flag. When this flag is set, only the listed features are
enabled.

We'll cover each use case separately, but it's also possible to combine these
if your implementation implements both. There are also options which pertain
to the entire test suite regardless of which tests you're running.

#### Gateway Tests

By default `Gateway` oriented conformance tests will expect a GatewayClass
named `gateway-conformance` to be installed in the cluster, and tests will be
run against that. Most often, you'll use a different class, which can be
specified with the `-gateway-class` flag along with the corresponding test
command. Check your instance for the `gateway-class` name to use. You must
also enable `Gateway` support and test support for any `*Routes` your
implementation supports.

The following runs all the tests relevant to `Gateway`, `HTTPRoute`, and
`ReferenceGrant`:

```shell
go test ./conformance -run TestConformance -args \
    --gateway-class=my-gateway-class \
    --supported-features=Gateway,HTTPRoute,ReferenceGrant
```

Other useful flags may be found in [conformance flags][cflags].

[cflags]: https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/utils/flags/flags.go

#### Mesh Tests

Mesh tests can be run by simply enabling the `Mesh` feature:

```shell
go test ./conformance -run TestConformance -args --supported-features=Mesh
```

If your mesh also includes ingress support with an API such as `HTTPRoute`, you
can run the relevant tests in the same test run by enabling the `Gateway`
feature and any relevant API features, e.g:

```shell
go test ./conformance -run TestConformance -args --supported-features=Mesh,Gateway,HTTPRoute
```

#### Namespace Labels and Annotations

If labels are needed on namespaces used for testing, you can use the
`-namespace-labels` flag to pass one or more `name=value` labels to set on the
test namespaces. Likewise, `-namespace-annotations` can be used to specify
annotations to be applied to the test namespaces. For mesh testing, this flag
can be used if an implementation requires labels on namespaces that host mesh
workloads, for example, to enable sidecar injection.

As an example, when testing Linkerd, you might run

```shell
go test ./conformance -run TestConformance -args \
   ...
   --namespace-annotations=linkerd.io/inject=enabled
```

so that the test namespaces are correctly injected into the mesh.

#### Excluding Features

The `-exempt-features` flag removes features from a manual selection. This can
be useful when focusing on a smaller set of tests during development. For
example, the following excludes `ReferenceGrant` tests:

```shell
go test ./conformance -run TestConformance -args \
    --supported-features=Gateway,HTTPRoute,ReferenceGrant \
    --exempt-features=ReferenceGrant
```

#### Suite Level Options

When running tests of any kind you may not want the test suite to cleanup the
test resources when it completes (i.e. so that you can inspect the cluster
state in the event of a failure). You can skip cleanup by setting the `--cleanup-base-resources=false`
flag.

It may be helpful (particularly when working on implementing a specific
feature) to run a very specific test by name. This can be done by setting the
`--run-test` flag.

#### Network Policies

In clusters that use [Container Network Interface (CNI) plugins][network_plugins]
which enforce network policies some conformance tests might require custom
[`NetworkPolicy`][netpol] resources to be added to the cluster in order to allow
the traffic to reach the required destinations.

Users are expected to add those required network policies so that their
implementations pass those tests.

[network_plugins]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[netpol]: https://kubernetes.io/docs/concepts/services-networking/network-policies/

### Compiling Conformance tests

{{% alert color="primary" %}}
This step is optional, and the binary can be used to execute conformance tests
on an environment different from the developer environment

{{% /alert %}}

In case there's a need to generate a conformance test binary to be executed on a 
different environment, a binary `conformance.test` can be generated with:

```bash
make conformance-bin
```

Specific GOOS and GOARCH flags can be passed to generate the binary for different
architectures:

```bash
make GOOS=darwin GOARCH=arm64 conformance-bin
```

Then the conformance binary can be called as `./conformance.test <flags>`