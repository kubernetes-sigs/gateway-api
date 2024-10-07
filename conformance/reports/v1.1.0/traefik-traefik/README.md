# Traefik Proxy

## Table of Contents

| API channel  | Implementation version                                             | Mode    | Report                                            |
|--------------|--------------------------------------------------------------------|---------|---------------------------------------------------|
| experimental | [v3.1](https://github.com/traefik/traefik/releases/tag/v3.1.0-rc2) | default | [link](./experimental-v3.1.0-default-report.yaml) |

## Reproduce

To reproduce the results, clone the Traefik Proxy repository:

```shell
git clone https://github.com/traefik/traefik.git && cd traefik
```

Check out the desired version:

```shell
git checkout vX.Y
```

Run the conformance tests with:

```shell
make test-gateway-api-conformance
```

Check the produced report in the `./integration/conformance-reports` folder.
