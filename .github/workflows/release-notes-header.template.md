## Gateway API ${TAG} Release Notes

This is the monthly release for the Gateway API experimental channel for ${MONTH_NAME} ${YEAR}. This release includes the latest features and fixes from Gateway API's main branch.

## Using this Release

To install the CRDs for this release, use install ${TAG}-install.yaml:

```
kubectl apply --server-side=true -f https://github.com/kubernetes-sigs/gateway-api/releases/download/${TAG}/${TAG}-install.yaml

helm upgrade --server-side=true --install gateway-api oci://ghcr.io/kubernetes-sigs/gateway-api/crds --version ${TAG}
```

To build using this release in Go, include this release in your go.mod:

```
require sigs.k8s.io/gateway-api ${TAG}
```

and run `go mod tidy`. You'll find that ${TAG} gets replaced by a Go pseudoversion; this is expected.

## Changes Summary

