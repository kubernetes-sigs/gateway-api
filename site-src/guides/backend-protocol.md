# Backend Protocol

Not all Gateway implementations support automatic protocol selection. Even in some cases protocols are disabled without an explicit opt-in (eg. websockets with Contour & NGINX). 

When a Route's backend references a Kubernetes Service application developers can specify the protocol using `ServicePort` [`appProtocol`][appProtocol] field.

For example the following `frontend` Kubernetes Service is indicating the port `8080` supports HTTP/2 Prior Knowledge.


```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend
spec:
  selector:
    app: frontend
  ports:
  - protocol: TCP
    appProtocol: kubernetes.io/h2c
    port: 8080
    targetPort: 8080
```

Currently, Gateway conformance is testing support for:

- `kubernetes.io/h2c` - HTTP/2 Prior Knowledge
- `kubernetes.io/ws` - WebSocket over HTTP

[appProtocol]: https://kubernetes.io/docs/concepts/services-networking/service/#application-protocol
