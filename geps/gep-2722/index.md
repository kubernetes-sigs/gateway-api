# GEP-2722: Goals and UX for gwctl

* Issue: [#2722](https://github.com/kubernetes-sigs/gateway-api/issues/2722)
* Status: Memorandum

## TLDR

TLDR: This GEP proposes `gwctl`, a new command line tool designed to streamline
  the experience of working with Gateway API resources. It offers a familiar
  kubectl-like interface for viewing resources while providing more detailed and
  informative output that is specifically focused on the Gateway API. For
  advanced filtering and other in-depth features, `gwctl` can be effectively
  used alongside `kubectl`.

## Motivations

* Limited kubectl customizability for CRDs: 
    * kubectl's customization capabilities for CRDs (through
      `additionalPrinterColumns`) is constrained, limiting the ability to create
      optimal views for Gateway API resources. 
* Complex policy attachment management: 
    * As described in [GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/),
      policies present a valuable mechanism for expanding the capabilities of
      Gateway API resources. However, discoverability poses a challenge due to
      the absence of a clear connection between resources and their associated
      policies. There have been growing questions around suitability of policies
      as a means to provide extensions.
* Challenging multi-resource model navigation:
    * Comprehending the relationships between multiple Gateway API resources can
      be challenging within kubectl. 

## Goals

* Greater control over output formatting and presentation:
    * Offer greater control over output formatting and presentation, enhancing
      visibility and understanding of Gateway API resources.
* Improved policy discoverability, increasing adoption and usability:
    * Make policies easily discoverable, aiding in the adoption and fostering
      broader acceptance of policies as an extension mechanism.
* Simplified multi-resource model navigation:
    * Facilitate navigation of the multi-resource model by making connections
      between Gateway API resources explicit, aiding in configuration,
      troubleshooting, and issue identification.
* Proactive error detection and reporting:
    * Leverage native understanding of resource relationships to proactively
      detect and report on potential configuration errors, further simplifying
      issue identification and resolution. This would complement the ability of
      users to readily pinpoint configuration problems themselves.
* Provide incentive for policy implementations that are consistent across cloud
  providers: 
    * Encourage the adoption of consistent policy implementations across
      different Gateway API providers, promoting interoperability and
      predictability.

## Commands Specification

### Milestone 1

#### Supported Commands:

* **get**: Retrieves information about specified resources without including
  additional information from related resources.
* **describe**: Provides detailed information about specified resources,
  including augmented information from related resources.

#### Supported Resources:

* gatewayclass
* gateways
* httproutes
* namespaces
* backends (not a native k8s resource)
* policycrds (not a native k8s resource)
* policies (not a native k8s resource)

#### Filtering Options:

* `-n` **Namespace**: Filters resources by namespace. Applicable to all
  resources except cluster-scoped resources.
* `-l` **Labels**: Filters resources by labels. Applicable to all resources.
* `-A` **All Namespaces**: Fetches resources across all namespaces (redundant
  for cluster-scoped resources).
* `-t` **Target Resource**: Filters policies based on the target resource type
  they apply to. Applicable only to the policies resource.
    * Syntax: `-t <key1>=<value1>,<key2>=<value2>,...`
    * Supported keys:
        * kind: Resource kind (e.g., "httproute", "gateway")
        * namespace: Resource namespace
        * name: Resource name
        * group: Resource API group

#### Output Formats:

* **describe**: Fixed format, not customizable. Shows comprehensive resource
  information with details from related resources.

* **get**:
    * One-line format (default): Displays basic resource information in a single
      line.
    * YAML format (-o yaml): Presents resource information in the YAML data
      format.
    * JSON format (-o json): Presents resource information in the JSON data
      format.
    * Wide format (-o wide): Includes additional columns beyond those displayed in
      the one-line format.
  
  <details>
    <summary>Output columns while using get</summary>
    <table>
        <tbody>
            <tr>
                <th>Resource</th>
                <th>Output Columns</th>
                <th>Description</th>
                <th>Visibility (Defaults to always unless specified otherwise)</th>
            </tr>
            <tr>
                <td rowspan="6">gatewayclass</td>
                <td>NAME</td>
                <td>Name of the GatewayClass</td>
                <td></td>
            </tr>
            <tr>
                <td>CONTROLLER</td>
                <td>Controller managing the GatewayClass</td>
                <td></td>
            </tr>
            <tr>
                <td>ACCEPTED</td>
                <td>Whether the GatewayClass is accepted by the controller</td>
                <td></td>
            </tr>
            <tr>
                <td>AGE</td>
                <td>Age of the GatewayClass</td>
                <td></td>
            </tr>
            <tr>
                <td>GATEWAYS</td>
                <td>Count of Gateways using this GatewayClass</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td>DESCRIPTION</td>
                <td>Description from the GatewayClass</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td rowspan="8">gateway</td>
                <td>NAME</td>
                <td>Name of the Gateway</td>
                <td></td>
            </tr>
            <tr>
                <td>CLASS</td>
                <td>Class of the Gateway</td>
                <td></td>
            </tr>
            <tr>
                <td>ADDRESSES</td>
                <td>Addresses of the Gateway (displayed using <addresses> + n more)</addresses>
                </td>
                <td></td>
            </tr>
            <tr>
                <td>PORTS</td>
                <td>Ports exposed by the Gateway</td>
                <td></td>
            </tr>
            <tr>
                <td>PROGRAMMED</td>
                <td>Whether the Gateway is programmed</td>
                <td></td>
            </tr>
            <tr>
                <td>AGE</td>
                <td>Age of the Gateway</td>
                <td></td>
            </tr>
            <tr>
                <td>POLICIES</td>
                <td>Count of policies affecting this Gateway</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td>HTTPROUTES</td>
                <td>Count of HTTPRoutes that are attached to this Gateway</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td rowspan="6">httproute</td>
                <td>NAMESPACE</td>
                <td>Namespace of the HTTPRoute</td>
                <td></td>
            </tr>
            <tr>
                <td>NAME</td>
                <td>Name of the HTTPRoute</td>
                <td></td>
            </tr>
            <tr>
                <td>HOSTNAMES</td>
                <td>Hostnames associated with the HTTPRoute</td>
                <td></td>
            </tr>
            <tr>
                <td>PARENT REFS</td>
                <td>Count of parent references of the HTTPRoute (e.g., Gateways)</td>
                <td></td>
            </tr>
            <tr>
                <td>AGE</td>
                <td>Age of the HTTPRoute</td>
                <td></td>
            </tr>
            <tr>
                <td>POLICIES</td>
                <td>Count of policies affecting this HTTPRoute</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td rowspan="4">namespace</td>
                <td>NAME</td>
                <td>Name of the namespace</td>
                <td></td>
            </tr>
            <tr>
                <td>STATUS</td>
                <td>Status of the namespace</td>
                <td></td>
            </tr>
            <tr>
                <td>AGE</td>
                <td>Age of the namespace</td>
                <td></td>
            </tr>
            <tr>
                <td>POLICIES</td>
                <td>Count of policies affecting this Namespace</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td rowspan="5">backend</td>
                <td>NAME</td>
                <td>Name of the backend</td>
                <td></td>
            </tr>
            <tr>
                <td>TYPE</td>
                <td>Type of the backend (currently only supports Services)</td>
                <td></td>
            </tr>
            <tr>
                <td>REFERRED BY ROUTES</td>
                <td>HTTPRoutes that refer to the backend (displayed using <names> + n more)</names>
                </td>
                <td></td>
            </tr>
            <tr>
                <td>AGE</td>
                <td>Age of the backend</td>
                <td></td>
            </tr>
            <tr>
                <td>POLICIES</td>
                <td>Count of policies affecting this Backend</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td rowspan="5">policycrd</td>
                <td>NAME</td>
                <td>Name of the Policy CRD in the form &lt;kind.group&gt;</td>
                <td></td>
            </tr>
            <tr>
                <td>POLICY TYPE</td>
                <td>Type of policy defined by the CRD (Inherited or Direct)</td>
                <td></td>
            </tr>
            <tr>
                <td>SCOPE</td>
                <td>Scope of the policy (Namespaced or Cluster)</td>
                <td></td>
            </tr>
            <tr>
                <td>POLICIES COUNT</td>
                <td>Count of policy resources of this particular type.</td>
                <td>-o wide</td>
            </tr>
            <tr>
                <td>AGE</td>
                <td>Age of the Policy CRD</td>
                <td></td>
            </tr>
            <tr>
                <td rowspan="6">policy</td>
                <td>NAME</td>
                <td>Name of the policy</td>
                <td></td>
            </tr>
            <tr>
                <td>KIND</td>
                <td>The kind of policy in the form &lt;kind.group&gt;
                </td>
                <td></td>
            </tr>
            <tr>
                <td>TARGET NAME</td>
                <td>Name of the resource the policy applies to</td>
                <td></td>
            </tr>
            <tr>
                <td>TARGET KIND</td>
                <td>The kind of target resource in the form</td>
                <td></td>
            </tr>
            <tr>
                <td>POLICY TYPE</td>
                <td>Type of policy (Inherited or Direct)</td>
                <td></td>
            </tr>
            <tr>
                <td>AGE</td>
                <td>Age of the Policy CRD</td>
                <td></td>
            </tr>
        </tbody>
    </table>
  </details>

#### Additional Notes:

* The behavior of `get` and `describe` commands is similar to kubectl, with
  `get` focusing on concise resource information and `describe` providing
  comprehensive details.
* `backends` represent resources that can be attached as backends to HTTPRoutes
  (currently limited to k8s services).
* `policycrds` and `policies` are not native k8s resources but represent subsets
  of CRDs and custom resource objects related to policies. (`policycrds` are
  identified by
  [PolicyLabelKey](https://github.com/kubernetes-sigs/gateway-api/blob/5658635bce70f3a52b6936ad5a99249a4f5116ad/apis/v1alpha2/policy_types.go#L31C2-L31C16))

#### Examples of commands that should be supported:

* `gwctl get gateways -n foo` (Lists basic information about Gateways in the
  "foo" namespace)
* `gwctl get httproutes -l version=v1,app=myapp` (Lists basic information about
  HTTPRoutes with the labels "version=v1" and "app=myapp")
* `gwctl get gateways -n foo -o yaml` (Shows detailed Gateway information in
  YAML format within the "foo" namespace)
* `gwctl get httproutes -l version=v1,app=myapp -o json` (Shows detailed
  HTTPRoute information in JSON format with specified labels)
* `gwctl describe gateways my-gateway` (Provides comprehensive details about the
  "my-gateway" Gateway, including information from related resources)
* `gwctl describe policies my-policy` (Shows detailed information about the
  "my-policy" policy, encompassing data from relevant any related resources when
  applicable)
* `gwctl get policies -t kind=httproute` (Lists basic information about policies that apply to HTTPRoutes)

#### Distribution
To ensure gwctl is widely accessible and easy to adopt, the following
distribution mechanisms will be provided:

* **Prebuilt Binaries:** Prebuilt binaries for various platforms (Linux, macOS,
  Windows) will be made available for download. Tooling like
  [GoReleaser](https://goreleaser.com/) can be used to streamline some of the
  build processes. Binaries will be offered in two variants:
    * `gwctl` for standalone use.
    * `kubectl-gw` for use as a kubectl plugin (`kubectl gw`).
* **Kubectl Plugin Integration:** gwctl will be integrated with
  [Krew](https://github.com/kubernetes-sigs/krew), the kubectl plugin manager.
  This should immensely help with improving discoverability of the plugin,
  allowing easier installation, and handling automatic updates for the user.
* **Versioning:** gwctl versions will be aligned with Gateway API releases (for
  the time when gwctl is developed within the same repository as Gateway API)
    * As gwctl matures, the need for maintaining it within the primary Gateway API
      repository will be reassessed. Factors such as a potential divergence in
      release cadence, independent contributor growth or the desire to reduce the
      triage workload for Gateway API maintainers could motivate a move to a
      separate repository.

### Future Milestones
* Each output of `describe` will include an extra `Analysis` field. This field
  will display any errors or other analysis information associated with the
  resource.
* Investigate the feasibility and any advantages of using Graphviz or webview
  for visualizing data and presenting information in a visually appealing
  manner.
* Evaluate how gwctl can be extended to support [Mesh use
  cases](https://gateway-api.sigs.k8s.io/concepts/gamma/#how-the-gateway-api-works-for-service-mesh)

## References

* [GEP-713: Metaresources and Policy
  Attachment](https://gateway-api.sigs.k8s.io/geps/gep-713)

## Sample outputs

The example outputs provided below serve as a guideline for the implementation,
outlining the range of values that may be presented:

* `gwctl get gatewayclass -o wide`
  ```
  NAME                            CONTROLLER                      ACCEPTED  AGE   DESCRIPTION             Gateways
  bar-com-internal-gateway-class  bar.baz/internal-gateway-class  True      100d  Internal Load Balancer  10
  foo-com-external-gateway-class  foo.com/external-gateway-class  True      365d  External Load Balancer  25
  ```

* `gwctl get gateway -o wide`
  ```
  NAME               CLASS                    ADDRESSES      PORTS     PROGRAMMED  AGE  POLICIES  HTTPROUTES
  demo-gateway-2     external-class           10.0.0.1       80        True        20d  10        5
  abc-gateway-12345  internal-class           192.168.100.5  443,8080  False       5d   2         1
  random-gateway     regional-internal-class  10.11.12.13    8443      Unknown     3s   3         5
  ```

* `gwctl get httproute -o wide`
  ```
  NAMESPACE  NAME                 HOSTNAMES                          PARENT REFS               AGE  POLICIES
  default    foo-httproute-1      example.com,example2.com + 1 more  ns2/demo-gateway-2        5m   2
  default    qmn-httproute-100    example.com                        demo-gateway-1            5m   1
  ns1        bar-route-21         foo.com,bar.com + 5 more           default/demo-gateway-200  5m   3
  ns2        bax-httproute-18777  None                               ns1/demo-gateway-345      5m   4
  ```

* `gwctl get namespace -o wide`
  ```
  NAME         STATUS  AGE  POLICIES
  default      Active  46d  3
  kube-system  Active  46d  5
  ```

* `gwctl get backend -o wide`
  ```
  NAME         TYPE     REFERRED BY ROUTES                         AGE  POLICIES
  foo-svc      Service  foo-httproute-1,abc-httproute-33 + 4 more  45m  5
  bar-baz-svc  Service  bar-httproute                              11d  1
  ```

* `gwctl get policycrds -o wide`
  ```
  NAME                               POLICY TYPE  SCOPE       POLICIES COUNT  AGE
  healthcheckpolicies.foo.com        Direct       Namespaced  1               5d
  retryonpolicies.foo.com            Direct       Namespaced  2               4d
  timeoutpolicies.bar.com            Inherited    Cluster     1               10m
  tlsminimumversionpolicies.baz.com  Direct       Namespaced  3               45s
  ```

* `gwctl get policies -o wide`
  ```
  NAME                                 KIND                                TARGET NAME                     TARGET KIND   POLICY TYPE  AGE
  demo-timeout-policy-on-gatewayclass  TimeoutPolicy.foo.com               foo-com-external-gateway-class  GatewayClass  Inherited    10d
  demo-timeout-policy-on-namespace     TimeoutPolicy.foo.com               default                         Namespace     Inherited    10d
  demo-health-check-1                  HealthCheckPolicy.bar.com           demo-gateway-1                  Gateway       Direct       10d
  demo-retry-policy-1                  RetryOnPolicy.baz.com               demo-gateway-1                  Gateway       Direct       10d
  demo-retry-policy-2                  RetryOnPolicy.baz.com               demo-httproute-2                HTTPRoute     Direct       10d
  demo-tls-min-version-policy-1        TLSMinimumVersionPolicy.foobar.com  demo-gateway-3                  Gateway       Direct       10d
  demo-tls-min-version-policy-2        TLSMinimumVersionPolicy.foobar.com  demo-gateway-4                  Gateway       Direct       10d
  ```

* `gwctl describe gateway demo-gateway`
  ```
  Name: demo-gateway
  Namespace: default
  Labels: <none>
  Annotations:
    annotation.foo: value1
    annotation.bar.baz: abcdefghijkl
  API Version: gateway.networking.k8s.io/v1beta1
  Kind: Gateway
  Metadata:
    creationTimestamp: "2023-12-01T18:29:41Z"
    finalizers:
    - gateway.finalizer.networking.io
    generation: 4
    resourceVersion: "310164667"
    uid: ed046878-f659-4908-b80f-b88c9617ba8a
  Spec:
    gatewayClassName: l7-global-external-managed
    listeners:
    - allowedRoutes:
        namespaces:
          from: Same
      name: http
      port: 80
      protocol: HTTP
  Status:
    addresses:
    - type: IPAddress
      value: 10.0.0.1
    conditions:
    - lastTransitionTime: "2023-12-01T18:49:25Z"
      message: ""
      observedGeneration: 3
      reason: Programmed
      status: "True"
      type: Programmed
    listeners:
    - attachedRoutes: 1
      conditions:
      - lastTransitionTime: "2023-12-01T18:49:25Z"
        message: Some message
        observedGeneration: 3
        reason: Ready
        status: "True"
        type: Ready
      name: http
      supportedKinds:
      - group: gateway.networking.k8s.io
        kind: HTTPRoute
  AttachedRoutes:
    Kind        Name                 Namespace
    ----        ----                 ---------
    HTTPRoute   demo-health-check-1  default
    TCPRoute    demo-retry-policy-1  default
  DirectlyAttachedPolicies:
    TYPE                   NAME
    ----                   ----
    TimeoutPolicy.foo.com  demo-timeout-policy-on-gatewayclass
    RetryOnPolicy.baz.com  demo-retry-policy-1
  InheritedPolicies:
    TYPE                   NAME                                 TARGET KIND   TARGET NAME
    ----                   ----                                 -----------   -----------
    TimeoutPolicy.foo.com  demo-timeout-policy-on-gatewayclass  GatewayClass  abc-gatewayclass
  EffectivePolicies:
    HealthCheckPolicy.foo.com:
      sampleParentField:
        sampleField: hello
    RetryOnPolicy.foo.com:
      sampleParentField:
        sampleField: namaste
    TimeoutPolicy.bar.com:
      timeout1: parent
      timeout2: child
      timeout3: parent
      timeout4: child
  Events:
    Type    Reason  Age                    From                   Message
    ----    ------  ----                   ----                   -------
    Normal  SYNC    2m12s (x46 over 138m)  sc-gateway-controller  SYNC on default/demo-gateway was a success
  ```

* `gwctl describe httproute demo-httproute`
  ```
  Name: demo-httproute
  Namespace: default
  Labels: <none>
  Annotations: <none>
  API Version: gateway.networking.k8s.io/v1beta1
  Kind: HTTPRoute
  Metadata:
    creationTimestamp: "2023-11-09T09:45:03Z"
    generation: 1
    resourceVersion: "290416533"
    uid: 716d9e5f-f57a-4e56-81f6-c579d5d17471
  Spec:
    hostnames:
    - example.com
    parentRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: demo-gateway
    rules:
    - backendRefs:
      - group: ""
        kind: Service
        name: demo-svc
        port: 80
        weight: 1
      matches:
      - path:
          type: PathPrefix
          value: /example
  Status:
    parents:
    - conditions:
      - lastTransitionTime: "2023-12-01T18:49:14Z"
        message: ""
        observedGeneration: 1
        reason: ReconciliationSucceeded
        status: "True"
        type: Reconciled
      controllerName: networking.io/gateway
      parentRef:
        group: gateway.networking.k8s.io
        kind: Gateway
        name: demo-gateway
  DirectlyAttachedPolicies:
    TYPE                       NAME
    ----                       ----
    HealthCheckPolicy.foo.com  demo-health-check-1
    RetryOnPolicy.baz.com      demo-retry-policy-1
  InheritedPolicies:
    TYPE                   NAME                                 TARGET KIND   TARGET NAME
    ----                   ----                                 -----------   -----------
    TimeoutPolicy.foo.com  demo-timeout-policy-on-gatewayclass  GatewayClass  abc-gatewayclass
    RetryOnPolicy.baz.com  demo-retry-policy-1                  Gateway       abc-gateway
  EffectivePolicies:
    HealthCheckPolicy.foo.com:
      sampleParentField:
        sampleField: hello
    RetryOnPolicy.foo.com:
      sampleParentField:
        sampleField: namaste
    TimeoutPolicy.bar.com:
      timeout1: parent
      timeout2: child
      timeout3: parent
      timeout4: child
  Events:
    Type    Reason  Age                    From                   Message
    ----    ------  ----                   ----                   -------
    Normal  SYNC    2m12s (x46 over 138m)  sc-gateway-controller  SYNC on default/demo-gateway was a success
  ```

* `gwctl describe gatewayclass foo-com-external-gateway-class`
  ```
  Name: foo-com-external-gateway-class
  Labels: <none>
  Annotations <none>
  API Version gateway.networking.k8s.io/v1beta1
  Kind: GatewayClass
  Metadata:
    creationTimestamp: "2023-06-28T17:33:03Z"
    generation: 1
    resourceVersion: "108322484"
    uid: 80cea521-5416-41c4-b5d1-2ee30f5366a6
  ControllerName: foo.com/external-gateway-class
  Description: Create an external load balancer
  Status:
    conditions:
    - lastTransitionTime: "2023-05-22T17:29:47Z"
      message: ""
      observedGeneration: 1
      reason: Accepted
      status: "True"
      type: Accepted
  DirectlyAttachedPolicies:
    TYPE                   NAME
    ----                   ----
    TimeoutPolicy.bar.com  demo-timeout-policy-on-gatewayclass
  ```
