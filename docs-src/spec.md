<p>Packages:</p>
<ul>
<li>
<a href="#networking.x-k8s.io%2fv1alpha1">networking.x-k8s.io/v1alpha1</a>
</li>
</ul>
<h2 id="networking.x-k8s.io/v1alpha1">networking.x-k8s.io/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains API Schema definitions for the networking.x-k8s.io
API group.</p>
</p>
Resource Types:
<ul><li>
<a href="#networking.x-k8s.io/v1alpha1.Gateway">Gateway</a>
</li><li>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClass">GatewayClass</a>
</li><li>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRoute">HTTPRoute</a>
</li><li>
<a href="#networking.x-k8s.io/v1alpha1.TCPRoute">TCPRoute</a>
</li><li>
<a href="#networking.x-k8s.io/v1alpha1.TLSRoute">TLSRoute</a>
</li><li>
<a href="#networking.x-k8s.io/v1alpha1.UDPRoute">UDPRoute</a>
</li></ul>
<h3 id="networking.x-k8s.io/v1alpha1.Gateway">Gateway
</h3>
<p>
<p>Gateway represents an instantiation of a service-traffic handling
infrastructure by binding Listeners to a set of IP addresses.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
networking.x-k8s.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Gateway</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewaySpec">
GatewaySpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>gatewayClassName</code></br>
<em>
string
</em>
</td>
<td>
<p>GatewayClassName used for this Gateway. This is the name of a
GatewayClass resource.</p>
</td>
</tr>
<tr>
<td>
<code>listeners</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">
[]Listener
</a>
</em>
</td>
<td>
<p>Listeners associated with this Gateway. Listeners define
logical endpoints that are bound on this Gateway&rsquo;s addresses.
At least one Listener MUST be specified.</p>
<p>Each Listener in this array must have a unique Port field,
however a GatewayClass may collapse compatible Listener
definitions into a single implementation-defined acceptor
configuration even if their Port fields would otherwise conflict.</p>
<p>Listeners are compatible if all of the following conditions are true:</p>
<ol>
<li>all their Protocol fields are &ldquo;HTTP&rdquo;, or all their Protocol fields are &ldquo;HTTPS&rdquo; or TLS&rdquo;</li>
<li>their Hostname fields are specified with a match type other than &ldquo;Any&rdquo;</li>
<li>their Hostname fields are not an exact match for any other Listener</li>
</ol>
<p>As a special case, each group of compatible listeners
may contain exactly one Listener with a match type of &ldquo;Any&rdquo;.</p>
<p>If the GatewayClass collapses compatible Listeners, the
hostname provided in the incoming client request MUST be
matched to a Listener to find the correct set of Routes.
The incoming hostname MUST be matched using the Hostname
field for each Listener in order of most to least specific.
That is, &ldquo;Exact&rdquo; matches must be processed before &ldquo;Domain&rdquo;
matches, which must be processed before &ldquo;Any&rdquo; matches.</p>
<p>If this field specifies multiple Listeners that have the same
Port value but are not compatible, the GatewayClass must raise
a &ldquo;PortConflict&rdquo; condition on the Gateway.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>addresses</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayAddress">
[]GatewayAddress
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Addresses requested for this gateway. This is optional and
behavior can depend on the GatewayClass. If a value is set
in the spec and the requested address is invalid, the
GatewayClass MUST indicate this in the associated entry in
GatewayStatus.Addresses.</p>
<p>If no Addresses are specified, the GatewayClass may
schedule the Gateway in an implementation-defined manner,
assigning an appropriate set of Addresses.</p>
<p>The GatewayClass MUST bind all Listeners to every
GatewayAddress that it assigns to the Gateway.</p>
<p>Support: Core</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayStatus">
GatewayStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayClass">GatewayClass
</h3>
<p>
<p>GatewayClass describes a class of Gateways available to the user
for creating Gateway resources.</p>
<p>GatewayClass is a Cluster level resource.</p>
<p>Support: Core.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
networking.x-k8s.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>GatewayClass</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassSpec">
GatewayClassSpec
</a>
</em>
</td>
<td>
<p>Spec for this GatewayClass.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>controller</code></br>
<em>
string
</em>
</td>
<td>
<p>Controller is a domain/path string that indicates the
controller that is managing Gateways of this class.</p>
<p>Example: &ldquo;acme.io/gateway-controller&rdquo;.</p>
<p>This field is not mutable and cannot be empty.</p>
<p>The format of this field is DOMAIN &ldquo;/&rdquo; PATH, where DOMAIN
and PATH are valid Kubernetes names
(<a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names">https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names</a>).</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>allowedGatewayNamespaces</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowedGatewayNamespaces is a selector of namespaces that Gateways of
this class can be created in. Implementations must not support Gateways
when they are created in namespaces not specified by this field.</p>
<p>Gateways that appear in namespaces not specified by this field must
continue to be supported if they have already been provisioned. This must
be indicated by the Gateway&rsquo;s presence in the ProvisionedGateways list in
the status for this GatewayClass. If the status on a Gateway indicates
that it has been provisioned but the Gateway does not appear in the
ProvisionedGateways list on GatewayClass it must not be supported.</p>
<p>When this field is unspecified (default) or an empty selector, Gateways
in any namespace will be able to use this GatewayClass.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>parametersRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ParametersRef is a controller-specific resource containing the
configuration parameters corresponding to this class. This is optional if
the controller does not require any additional configuration.</p>
<p>Parameters resources are implementation specific custom resources. These
resources must be cluster-scoped.</p>
<p>If the referent cannot be found, the GatewayClass&rsquo;s &ldquo;InvalidParameters&rdquo;
status condition will be true.</p>
<p>Support: Custom</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassStatus">
GatewayClassStatus
</a>
</em>
</td>
<td>
<p>Status of the GatewayClass.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRoute">HTTPRoute
</h3>
<p>
<p>HTTPRoute is the Schema for the HTTPRoute resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
networking.x-k8s.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>HTTPRoute</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteSpec">
HTTPRouteSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>hosts</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteHost">
[]HTTPRouteHost
</a>
</em>
</td>
<td>
<p>Hosts is a list of Host definitions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteStatus">
HTTPRouteStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TCPRoute">TCPRoute
</h3>
<p>
<p>TCPRoute is the Schema for the TCPRoute resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
networking.x-k8s.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>TCPRoute</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteSpec">
TCPRouteSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteRule">
[]TCPRouteRule
</a>
</em>
</td>
<td>
<p>Rules are a list of TCP matchers and actions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteStatus">
TCPRouteStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TLSRoute">TLSRoute
</h3>
<p>
<p>TLSRoute is the Schema for the TLSRoute resource.
TLSRoute is similar to TCPRoute but can be configured to match against
TLS-specific metadata.
This allows more flexibility in matching streams for in a given TLS listener.</p>
<p>If you need to forward traffic to a single target for a TLS listener, you
could chose to use a TCPRoute with a TLS listener.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
networking.x-k8s.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>TLSRoute</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteSpec">
TLSRouteSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteRule">
[]TLSRouteRule
</a>
</em>
</td>
<td>
<p>Rules are a list of TLS matchers and actions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteStatus">
TLSRouteStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.UDPRoute">UDPRoute
</h3>
<p>
<p>UDPRoute is the Schema for the UDPRoute resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
networking.x-k8s.io/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>UDPRoute</code></td>
</tr>
<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteSpec">
UDPRouteSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteRule">
[]UDPRouteRule
</a>
</em>
</td>
<td>
<p>Rules are a list of UDP matchers and actions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteStatus">
UDPRouteStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.AddressType">AddressType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayAddress">GatewayAddress</a>)
</p>
<p>
<p>AddressType defines how a network address is represented as a text string.
Valid AddressType values are:</p>
<ul>
<li>&ldquo;IPAddress&rdquo;</li>
<li>&ldquo;NamedAddress&rdquo;</li>
</ul>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayAddress">GatewayAddress
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewaySpec">GatewaySpec</a>, 
<a href="#networking.x-k8s.io/v1alpha1.GatewayStatus">GatewayStatus</a>)
</p>
<p>
<p>GatewayAddress describes an address that can be bound to a Gateway.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.AddressType">
AddressType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type of the Address. This is either &ldquo;IPAddress&rdquo; or &ldquo;NamedAddress&rdquo;.</p>
<p>Support: Extended</p>
</td>
</tr>
<tr>
<td>
<code>value</code></br>
<em>
string
</em>
</td>
<td>
<p>Value. Examples: &ldquo;1.2.3.4&rdquo;, &ldquo;128::1&rdquo;, &ldquo;my-ip-address&rdquo;. Validity of the
values will depend on <code>Type</code> and support by the controller.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayAllowType">GatewayAllowType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">RouteGateways</a>)
</p>
<p>
<p>GatewayAllowType specifies which Gateways should be allowed to use a Route.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayClassConditionType">GatewayClassConditionType
(<code>string</code> alias)</p></h3>
<p>
<p>GatewayClassConditionType is the type of status conditions. This
type should be used with the GatewayClassStatus.Conditions field.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayClassSpec">GatewayClassSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClass">GatewayClass</a>)
</p>
<p>
<p>GatewayClassSpec reflects the configuration of a class of Gateways.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>controller</code></br>
<em>
string
</em>
</td>
<td>
<p>Controller is a domain/path string that indicates the
controller that is managing Gateways of this class.</p>
<p>Example: &ldquo;acme.io/gateway-controller&rdquo;.</p>
<p>This field is not mutable and cannot be empty.</p>
<p>The format of this field is DOMAIN &ldquo;/&rdquo; PATH, where DOMAIN
and PATH are valid Kubernetes names
(<a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names">https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names</a>).</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>allowedGatewayNamespaces</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowedGatewayNamespaces is a selector of namespaces that Gateways of
this class can be created in. Implementations must not support Gateways
when they are created in namespaces not specified by this field.</p>
<p>Gateways that appear in namespaces not specified by this field must
continue to be supported if they have already been provisioned. This must
be indicated by the Gateway&rsquo;s presence in the ProvisionedGateways list in
the status for this GatewayClass. If the status on a Gateway indicates
that it has been provisioned but the Gateway does not appear in the
ProvisionedGateways list on GatewayClass it must not be supported.</p>
<p>When this field is unspecified (default) or an empty selector, Gateways
in any namespace will be able to use this GatewayClass.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>parametersRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ParametersRef is a controller-specific resource containing the
configuration parameters corresponding to this class. This is optional if
the controller does not require any additional configuration.</p>
<p>Parameters resources are implementation specific custom resources. These
resources must be cluster-scoped.</p>
<p>If the referent cannot be found, the GatewayClass&rsquo;s &ldquo;InvalidParameters&rdquo;
status condition will be true.</p>
<p>Support: Custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayClassStatus">GatewayClassStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClass">GatewayClass</a>)
</p>
<p>
<p>GatewayClassStatus is the current status for the GatewayClass.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions is the current status from the controller for
this GatewayClass.</p>
</td>
</tr>
<tr>
<td>
<code>provisionedGateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayReference">
[]GatewayReference
</a>
</em>
</td>
<td>
<p>ProvisionedGateways is a list of Gateways that have been provisioned
using this class. Implementations must add any Gateways of this class to
this list once they have been provisioned and remove Gateways as soon as
they are deleted or deprovisioned.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayConditionReason">GatewayConditionReason
(<code>string</code> alias)</p></h3>
<p>
<p>GatewayConditionReason defines the set of reasons that explain
why a particular Gateway condition type has been raised.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayConditionType">GatewayConditionType
(<code>string</code> alias)</p></h3>
<p>
<p>GatewayConditionType is a type of condition associated with a
Gateway. This type should be used with the GatewayStatus.Conditions
field.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayReference">GatewayReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassStatus">GatewayClassStatus</a>, 
<a href="#networking.x-k8s.io/v1alpha1.RouteGatewayStatus">RouteGatewayStatus</a>, 
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">RouteGateways</a>)
</p>
<p>
<p>GatewayReference identifies a Gateway in a specified namespace.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the referent.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<p>Namespace is the namespace of the referent.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewaySpec">GatewaySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Gateway">Gateway</a>)
</p>
<p>
<p>GatewaySpec defines the desired state of Gateway.</p>
<p>Not all possible combinations of options specified in the Spec are
valid. Some invalid configurations can be caught synchronously via a
webhook, but there are many cases that will require asynchronous
signaling via the GatewayStatus block.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>gatewayClassName</code></br>
<em>
string
</em>
</td>
<td>
<p>GatewayClassName used for this Gateway. This is the name of a
GatewayClass resource.</p>
</td>
</tr>
<tr>
<td>
<code>listeners</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">
[]Listener
</a>
</em>
</td>
<td>
<p>Listeners associated with this Gateway. Listeners define
logical endpoints that are bound on this Gateway&rsquo;s addresses.
At least one Listener MUST be specified.</p>
<p>Each Listener in this array must have a unique Port field,
however a GatewayClass may collapse compatible Listener
definitions into a single implementation-defined acceptor
configuration even if their Port fields would otherwise conflict.</p>
<p>Listeners are compatible if all of the following conditions are true:</p>
<ol>
<li>all their Protocol fields are &ldquo;HTTP&rdquo;, or all their Protocol fields are &ldquo;HTTPS&rdquo; or TLS&rdquo;</li>
<li>their Hostname fields are specified with a match type other than &ldquo;Any&rdquo;</li>
<li>their Hostname fields are not an exact match for any other Listener</li>
</ol>
<p>As a special case, each group of compatible listeners
may contain exactly one Listener with a match type of &ldquo;Any&rdquo;.</p>
<p>If the GatewayClass collapses compatible Listeners, the
hostname provided in the incoming client request MUST be
matched to a Listener to find the correct set of Routes.
The incoming hostname MUST be matched using the Hostname
field for each Listener in order of most to least specific.
That is, &ldquo;Exact&rdquo; matches must be processed before &ldquo;Domain&rdquo;
matches, which must be processed before &ldquo;Any&rdquo; matches.</p>
<p>If this field specifies multiple Listeners that have the same
Port value but are not compatible, the GatewayClass must raise
a &ldquo;PortConflict&rdquo; condition on the Gateway.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>addresses</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayAddress">
[]GatewayAddress
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Addresses requested for this gateway. This is optional and
behavior can depend on the GatewayClass. If a value is set
in the spec and the requested address is invalid, the
GatewayClass MUST indicate this in the associated entry in
GatewayStatus.Addresses.</p>
<p>If no Addresses are specified, the GatewayClass may
schedule the Gateway in an implementation-defined manner,
assigning an appropriate set of Addresses.</p>
<p>The GatewayClass MUST bind all Listeners to every
GatewayAddress that it assigns to the Gateway.</p>
<p>Support: Core</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayStatus">GatewayStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Gateway">Gateway</a>)
</p>
<p>
<p>GatewayStatus defines the observed state of Gateway.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>addresses</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayAddress">
[]GatewayAddress
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Addresses lists the IP addresses that have actually been
bound to the Gateway. These addresses may differ from the
addresses in the Spec, e.g. if the Gateway automatically
assigns an address from a reserved pool.</p>
<p>These addresses should all be of type &ldquo;IPAddress&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<p>Conditions describe the current conditions of the Gateway.</p>
<p>Implementations should prefer to express Gateway conditions
using the <code>GatewayConditionType</code> and <code>GatewayConditionReason</code>
constants so that operators and tools can converge on a common
vocabulary to describe Gateway state.</p>
<p>Known condition types are:</p>
<ul>
<li>&ldquo;Scheduled&rdquo;</li>
<li>&ldquo;Ready&rdquo;</li>
</ul>
</td>
</tr>
<tr>
<td>
<code>listeners</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ListenerStatus">
[]ListenerStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Listeners provide status for each unique listener port defined in the Spec.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayTLSConfig">GatewayTLSConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">Listener</a>)
</p>
<p>
<p>GatewayTLSConfig describes a TLS configuration.</p>
<p>References
- nginx: <a href="https://nginx.org/en/docs/http/configuring_https_servers.html">https://nginx.org/en/docs/http/configuring_https_servers.html</a>
- envoy: <a href="https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/auth/cert.proto">https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/auth/cert.proto</a>
- haproxy: <a href="https://www.haproxy.com/documentation/aloha/9-5/traffic-management/lb-layer7/tls/">https://www.haproxy.com/documentation/aloha/9-5/traffic-management/lb-layer7/tls/</a>
- gcp: <a href="https://cloud.google.com/load-balancing/docs/use-ssl-policies#creating_an_ssl_policy_with_a_custom_profile">https://cloud.google.com/load-balancing/docs/use-ssl-policies#creating_an_ssl_policy_with_a_custom_profile</a>
- aws: <a href="https://docs.aws.amazon.com/elasticloadbalancing/latest/application/create-https-listener.html#describe-ssl-policies">https://docs.aws.amazon.com/elasticloadbalancing/latest/application/create-https-listener.html#describe-ssl-policies</a>
- azure: <a href="https://docs.microsoft.com/en-us/azure/app-service/configure-ssl-bindings#enforce-tls-1112">https://docs.microsoft.com/en-us/azure/app-service/configure-ssl-bindings#enforce-tls-1112</a></p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>mode</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TLSModeType">
TLSModeType
</a>
</em>
</td>
<td>
<p>Mode defines the TLS behavior for the TLS session initiated by the client.
There are two possible modes:
- Terminate: The TLS session between the downstream client
and the Gateway is terminated at the Gateway.
- Passthrough: The TLS session is NOT terminated by the Gateway. This
implies that the Gateway can&rsquo;t decipher the TLS stream except for
the ClientHello message of the TLS protocol.
CertificateRef field is ignored in this mode.</p>
</td>
</tr>
<tr>
<td>
<code>certificateRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.SecretsDefaultLocalObjectReference">
SecretsDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CertificateRef is the reference to Kubernetes object that
contain a TLS certificate and private key.
This certificate MUST be used for TLS handshakes for the domain
this GatewayTLSConfig is associated with.
If an entry in this list omits or specifies the empty
string for both the group and the resource, the resource defaults to &ldquo;secrets&rdquo;.
An implementation may support other resources (for example, resource
&ldquo;mycertificates&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Support: Core (Kubernetes Secrets)
Support: Implementation-specific (Other resource types)</p>
</td>
</tr>
<tr>
<td>
<code>options</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Options are a list of key/value pairs to give extended options
to the provider.</p>
<p>There variation among providers as to how ciphersuites are
expressed. If there is a common subset for expressing ciphers
then it will make sense to loft that as a core API
construct.</p>
<p>Support: Implementation-specific.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GenericForwardToTarget">GenericForwardToTarget
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteAction">TCPRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteAction">TLSRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteAction">UDPRouteAction</a>)
</p>
<p>
<p>GenericForwardToTarget identifies a target object within a known namespace.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>targetRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ServicesDefaultLocalObjectReference">
ServicesDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<p>TargetRef is an object reference to forward matched requests to.
The resource may be &ldquo;services&rdquo; (omit or use the empty string for the
group), or an implementation may support other resources (for
example, resource &ldquo;myroutetargets&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;services&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: Core (Kubernetes Services)
Support: Implementation-specific (Other resource types)</p>
</td>
</tr>
<tr>
<td>
<code>targetPort</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TargetPort">
TargetPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetPort specifies the destination port number to use for the TargetRef.
If unspecified and TargetRef is a Service object consisting of a single
port definition, that port will be used. If unspecified and TargetRef is
a Service object consisting of multiple port definitions, an error is
surfaced in status.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>weight</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TargetWeight">
TargetWeight
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Weight specifies the proportion of traffic forwarded to a targetRef, computed
as weight/(sum of all weights in targetRefs). Weight is not a percentage and
the sum of weights does not need to equal 100. The following example (in yaml)
sends 70% of traffic to service &ldquo;my-trafficsplit-sv1&rdquo; and 30% of the traffic
to service &ldquo;my-trafficsplit-sv2&rdquo;:</p>
<p>forwardTo:
- targetRef:
name: my-trafficsplit-sv1
weight: 70
- targetRef:
name: my-trafficsplit-sv2
weight: 30</p>
<p>If only one targetRef is specified, 100% of the traffic is forwarded to the
targetRef. If unspecified, weight defaults to 1.</p>
<p>Support: Core (HTTPRoute)
Support: Extended (TCPRoute)</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPForwardToTarget">HTTPForwardToTarget
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardingTarget">HTTPForwardingTarget</a>)
</p>
<p>
<p>HTTPForwardToTarget identifies a target object within a known namespace.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>targetRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ServicesDefaultLocalObjectReference">
ServicesDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<p>TargetRef is an object reference to forward matched requests to.
The resource may be &ldquo;services&rdquo; (omit or use the empty string for the
group), or an implementation may support other resources (for
example, resource &ldquo;myroutetargets&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;services&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: Core (Kubernetes Services)
Support: Implementation-specific (Other resource types)</p>
</td>
</tr>
<tr>
<td>
<code>targetPort</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TargetPort">
TargetPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetPort specifies the destination port number to use for the TargetRef.
If unspecified and TargetRef is a Service object consisting of a single
port definition, that port will be used. If unspecified and TargetRef is
a Service object consisting of multiple port definitions, an error is
surfaced in status.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>weight</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TargetWeight">
TargetWeight
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Weight specifies the proportion of traffic forwarded to a targetRef, computed
as weight/(sum of all weights in targetRefs). Weight is not a percentage and
the sum of weights does not need to equal 100. The following example (in yaml)
sends 70% of traffic to service &ldquo;my-trafficsplit-sv1&rdquo; and 30% of the traffic
to service &ldquo;my-trafficsplit-sv2&rdquo;:</p>
<p>forwardTo:
- targetRef:
name: my-trafficsplit-sv1
weight: 70
- targetRef:
name: my-trafficsplit-sv2
weight: 30</p>
<p>If only one targetRef is specified, 100% of the traffic is forwarded to the
targetRef. If unspecified, weight defaults to 1.</p>
<p>Support: Core (HTTPRoute)
Support: Extended (TCPRoute)</p>
</td>
</tr>
<tr>
<td>
<code>filters</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">
[]HTTPRouteFilter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Filters defined at this-level should be executed if and only if
the request is being forwarded to the target defined here.</p>
<p>Conformance: For any implementation, filtering support, including core
filters, is NOT guaranteed at this-level.
Use Filters in HTTPRouteRule for portable filters across implementations.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPForwardingTarget">HTTPForwardingTarget
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteRule">HTTPRouteRule</a>)
</p>
<p>
<p>HTTPForwardingTarget is the target to send the request to for a given a match.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>to</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardToTarget">
[]HTTPForwardToTarget
</a>
</em>
</td>
<td>
<p>To references referenced object(s) where the request should be sent. The
resource may be &ldquo;services&rdquo; (omit or use the empty string for the
group), or an implementation may support other resources (for
example, resource &ldquo;myroutetargets&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;services&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: core</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;action&rdquo; behavior.  The resource may be &ldquo;configmaps&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myrouteactions&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPHeaderMatch">HTTPHeaderMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteMatch">HTTPRouteMatch</a>)
</p>
<p>
<p>HTTPHeaderMatch describes how to select a HTTP route by matching HTTP request headers.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HeaderMatchType">
HeaderMatchType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>HeaderMatchType specifies how to match a HTTP request
header against the Values map.</p>
<p>Support: core (Exact)
Support: custom (ImplementationSpecific)</p>
<p>Default: &ldquo;Exact&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>values</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Values is a map of HTTP Headers to be matched.
It MUST contain at least one entry.</p>
<p>The HTTP header field name to match is the map key, and the
value of the HTTP header is the map value. HTTP header field
names MUST be matched case-insensitively.</p>
<p>Multiple match values are ANDed together, meaning, a request
must match all the specified headers to select the route.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPPathMatch">HTTPPathMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteMatch">HTTPRouteMatch</a>)
</p>
<p>
<p>HTTPPathMatch describes how to select a HTTP route by matching the HTTP request path.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.PathMatchType">
PathMatchType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Type specifies how to match against the path Value.</p>
<p>Support: core (Exact, Prefix)
Support: custom (RegularExpression, ImplementationSpecific)</p>
<p>Since RegularExpression PathType has custom conformance, implementations
can support POSIX, PCRE or any other dialects of regular expressions.
Please read the implementation&rsquo;s documentation to determine the supported
dialect.</p>
<p>Default: &ldquo;Prefix&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>value</code></br>
<em>
string
</em>
</td>
<td>
<p>Value of the HTTP path to match against.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRequestHeaderFilter">HTTPRequestHeaderFilter
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">HTTPRouteFilter</a>)
</p>
<p>
<p>HTTPRequestHeaderFilter defines configuration for the
RequestHeader filter.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>add</code></br>
<em>
map[string]string
</em>
</td>
<td>
<p>Add adds the given header (name, value) to the request
before the action.</p>
<p>Input:
GET /foo HTTP/1.1</p>
<p>Config:
add: {&ldquo;my-header&rdquo;: &ldquo;foo&rdquo;}</p>
<p>Output:
GET /foo HTTP/1.1
my-header: foo</p>
<p>Support: extended?</p>
</td>
</tr>
<tr>
<td>
<code>remove</code></br>
<em>
[]string
</em>
</td>
<td>
<p>Remove the given header(s) from the HTTP request before the
action. The value of RemoveHeader is a list of HTTP header
names. Note that the header names are case-insensitive
[RFC-2616 4.2].</p>
<p>Input:
GET /foo HTTP/1.1
My-Header1: ABC
My-Header2: DEF
My-Header2: GHI</p>
<p>Config:
remove: [&ldquo;my-header1&rdquo;, &ldquo;my-header3&rdquo;]</p>
<p>Output:
GET /foo HTTP/1.1
My-Header2: DEF</p>
<p>Support: extended?</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRequestMirrorFilter">HTTPRequestMirrorFilter
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">HTTPRouteFilter</a>)
</p>
<p>
<p>HTTPRequestMirrorFilter defines configuration for the
RequestMirror filter.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>targetRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ServicesDefaultLocalObjectReference">
ServicesDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<p>TargetRef is an object reference to forward matched requests to.
The resource may be &ldquo;services&rdquo; (omit or use the empty string for the
group), or an implementation may support other resources (for
example, resource &ldquo;myroutetargets&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;services&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: Core (Kubernetes Services)
Support: Implementation-specific (Other resource types)</p>
</td>
</tr>
<tr>
<td>
<code>targetPort</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TargetPort">
TargetPort
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TargetPort specifies the destination port number to use for the TargetRef.
If unspecified and TargetRef is a Service object consisting of a single
port definition, that port will be used. If unspecified and TargetRef is
a Service object consisting of multiple port definitions, an error is
surfaced in status.</p>
<p>Support: Core</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteFilter">HTTPRouteFilter
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardToTarget">HTTPForwardToTarget</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteRule">HTTPRouteRule</a>)
</p>
<p>
<p>HTTPRouteFilter defines additional processing steps that must be completed
during the request or response lifecycle.
HTTPRouteFilters are meant as an extension point to express additional
processing that may be done in Gateway implementations. Some examples include
request or response modification, implementing authentication strategies,
rate-limiting, and traffic shaping.
API guarantee/conformance is defined based on the type of the filter.
TODO(hbagdi): re-render CRDs once controller-tools supports union tags:
- <a href="https://github.com/kubernetes-sigs/controller-tools/pull/298">https://github.com/kubernetes-sigs/controller-tools/pull/298</a>
- <a href="https://github.com/kubernetes-sigs/controller-tools/issues/461">https://github.com/kubernetes-sigs/controller-tools/issues/461</a></p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>type</code></br>
<em>
string
</em>
</td>
<td>
<p>Type identifies the filter to execute.
Types are classified into three conformance-levels (similar to
other locations in this API):
- Core and extended: These filter types and their corresponding configuration
is defined in this package. All implementations must implement
the core filters. Implementers are encouraged to support extended filters.
Definitions for filter-specific configuration for these
filters is defined in this package.
- Custom: These filters are defined and supported by specific vendors.
In the future, filters showing convergence in behavior across multiple
implementations will be considered for inclusion in extended or core
conformance rings. Filter-specific configuration for such filters
is specified using the ExtensionRef field. <code>Type</code> should be set to
&ldquo;ImplementationSpecific&rdquo; for custom filters.</p>
<p>Implementers are encouraged to define custom implementation
types to extend the core API with implementation-specific behavior.</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;filter&rdquo; behavior.  The resource may be &ldquo;configmap&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myroutefilters&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.
ExtensionRef MUST NOT be used for core and extended filters.</p>
</td>
</tr>
<tr>
<td>
<code>requestHeader</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRequestHeaderFilter">
HTTPRequestHeaderFilter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
<tr>
<td>
<code>requestMirror</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRequestMirrorFilter">
HTTPRequestMirrorFilter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteHost">HTTPRouteHost
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteSpec">HTTPRouteSpec</a>)
</p>
<p>
<p>HTTPRouteHost is the configuration for a given set of hosts.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>hostnames</code></br>
<em>
[]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Hostnames defines a set of hostname that should match against
the HTTP Host header to select a HTTPRoute to process the request.
Hostname is the fully qualified domain name of a network host,
as defined by RFC 3986. Note the following deviations from the
&ldquo;host&rdquo; part of the URI as defined in the RFC:</p>
<ol>
<li>IPs are not allowed.</li>
<li>The <code>:</code> delimiter is not respected because ports are not allowed.</li>
</ol>
<p>Incoming requests are matched against the hostnames before the
HTTPRoute rules. If no hostname is specified, traffic is routed
based on the HTTPRouteRules.</p>
<p>Hostname can be &ldquo;precise&rdquo; which is a domain name without the terminating
dot of a network host (e.g. &ldquo;foo.example.com&rdquo;) or &ldquo;wildcard&rdquo;, which is
a domain name prefixed with a single wildcard label (e.g. &ldquo;<em>.example.com&rdquo;).
The wildcard character &lsquo;</em>&rsquo; must appear by itself as the first DNS
label and matches only a single label.
You cannot have a wildcard label by itself (e.g. Host == &ldquo;*&rdquo;).
Requests will be matched against the Host field in the following order:
1. If Host is precise, the request matches this rule if
the http host header is equal to Host.
2. If Host is a wildcard, then the request matches this rule if
the http host header is to equal to the suffix
(removing the first label) of the wildcard rule.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteRule">
[]HTTPRouteRule
</a>
</em>
</td>
<td>
<p>Rules are a list of HTTP matchers, filters and actions.</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;host&rdquo; block. The resource may be &ldquo;configmaps&rdquo;  or an implementation-defined
resource (for example, resource &ldquo;myroutehosts&rdquo; in group &ldquo;networking.acme.io&rdquo;).</p>
<p>If the referent cannot be found,
the GatewayClass&rsquo;s &ldquo;InvalidParameters&rdquo; status condition
will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteMatch">HTTPRouteMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteRule">HTTPRouteRule</a>)
</p>
<p>
<p>HTTPRouteMatch defines the predicate used to match requests to a given
action. Multiple match types are ANDed together, i.e. the match will
evaluate to true only if all conditions are satisfied.</p>
<p>For example, the match below will match a HTTP request only if its path
starts with <code>/foo</code> AND it contains the <code>version: &quot;1&quot;</code> header:</p>
<pre><code>match:
path:
value: &quot;/foo&quot;
headers:
values:
version: &quot;1&quot;
</code></pre>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>path</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPPathMatch">
HTTPPathMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Path specifies a HTTP request path matcher. If this field is not
specified, a default prefix match on the &ldquo;/&rdquo; path is provided.</p>
</td>
</tr>
<tr>
<td>
<code>headers</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPHeaderMatch">
HTTPHeaderMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Headers specifies a HTTP request header matcher.</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;match&rdquo; behavior.  The resource may be &ldquo;configmap&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myroutematchers&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteRule">HTTPRouteRule
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteHost">HTTPRouteHost</a>)
</p>
<p>
<p>HTTPRouteRule defines semantics for matching an incoming HTTP request against
a set of matching rules and executing an action (and optionally filters) on
the request.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>matches</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteMatch">
[]HTTPRouteMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Matches define conditions used for matching the rule against
incoming HTTP requests.
Each match is independent, i.e. this rule will be matched
if <strong>any</strong> one of the matches is satisfied.</p>
<p>For example, take the following matches configuration:</p>
<pre><code>matches:
- path:
value: &quot;/foo&quot;
headers:
values:
version: &quot;2&quot;
- path:
value: &quot;/v2/foo&quot;
</code></pre>
<p>For a request to match against this rule, a request should satisfy
EITHER of the two conditions:</p>
<ul>
<li>path prefixed with <code>/foo</code> AND contains the header <code>version: &quot;2&quot;</code></li>
<li>path prefix of <code>/v2/foo</code></li>
</ul>
<p>See the documentation for HTTPRouteMatch on how to specify multiple
match conditions that should be ANDed together.</p>
<p>If no matches are specified, the default is a prefix
path match on &ldquo;/&rdquo;, which has the effect of matching every
HTTP request.</p>
</td>
</tr>
<tr>
<td>
<code>filters</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">
[]HTTPRouteFilter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Filters define the filters that are applied to requests that match
this rule.</p>
<p>The effects of ordering of multiple behaviors are currently undefined.
This can change in the future based on feedback during the alpha stage.</p>
<p>Conformance-levels at this level are defined based on the type of filter:
- ALL core filters MUST be supported by all implementations.
- Implementers are encouraged to support extended filters.
- Implementation-specific custom filters have no API guarantees across implementations.
Specifying a core filter multiple times has undefined or custom conformance.</p>
<p>Support: core</p>
</td>
</tr>
<tr>
<td>
<code>forward</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardingTarget">
HTTPForwardingTarget
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Forward defines the upstream target(s) where the request should be sent.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteSpec">HTTPRouteSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRoute">HTTPRoute</a>)
</p>
<p>
<p>HTTPRouteSpec defines the desired state of HTTPRoute</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>hosts</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteHost">
[]HTTPRouteHost
</a>
</em>
</td>
<td>
<p>Hosts is a list of Host definitions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteStatus">HTTPRouteStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRoute">HTTPRoute</a>)
</p>
<p>
<p>HTTPRouteStatus defines the observed state of HTTPRoute.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>RouteStatus</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteStatus">
RouteStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>RouteStatus</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HeaderMatchType">HeaderMatchType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPHeaderMatch">HTTPHeaderMatch</a>)
</p>
<p>
<p>HeaderMatchType specifies the semantics of how HTTP headers should be compared.
Valid HeaderMatchType values are:</p>
<ul>
<li>&ldquo;Exact&rdquo;</li>
<li>&ldquo;ImplementationSpecific&rdquo;</li>
</ul>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.HostnameMatch">HostnameMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">Listener</a>)
</p>
<p>
<p>HostnameMatch specifies how a Listener should match the incoming
hostname from a client request. Depending on the incoming protocol,
the match must apply to names provided by the client at both the
TLS and the HTTP protocol layers.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>match</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HostnameMatchType">
HostnameMatchType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Match specifies how the hostname provided by the client should be
matched against the given value.</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Name contains the name to match against. This value must
be a fully qualified host or domain name conforming to the
preferred name syntax defined in
<a href="https://tools.ietf.org/html/rfc1034#section-3.5">RFC 1034</a></p>
<p>In addition to any RFC rules, this field MUST NOT contain</p>
<ol>
<li>IP address literals</li>
<li>Colon-delimited port numbers</li>
<li>Percent-encoded octets</li>
</ol>
<p>This field is required for the &ldquo;Domain&rdquo; and &ldquo;Exact&rdquo; match types.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HostnameMatchType">HostnameMatchType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HostnameMatch">HostnameMatch</a>)
</p>
<p>
<p>HostnameMatchType specifies the types of matches that are valid
for hostnames.
Valid match types are:</p>
<ul>
<li>&ldquo;Domain&rdquo;</li>
<li>&ldquo;Exact&rdquo;</li>
<li>&ldquo;Any&rdquo;</li>
</ul>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.Listener">Listener
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewaySpec">GatewaySpec</a>)
</p>
<p>
<p>Listener embodies the concept of a logical endpoint where a Gateway can
accept network connections. Each listener in a Gateway must have a unique
combination of Hostname, Port, and Protocol. This will be enforced by a
validating webhook.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>hostname</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HostnameMatch">
HostnameMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Hostname specifies to match the virtual hostname for
protocol types that define this concept.</p>
<p>Incoming requests that include a hostname are matched
according to the given HostnameMatchType to select
the Routes from this Listener.</p>
<p>If a match type other than &ldquo;Any&rdquo; is supplied, it MUST
be compatible with the specified Protocol field.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
int32
</em>
</td>
<td>
<p>Port is the network port. Multiple listeners may use the
same port, subject to the Listener compatibility rules.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>protocol</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ProtocolType">
ProtocolType
</a>
</em>
</td>
<td>
<p>Protocol specifies the network protocol this listener
expects to receive. The GatewayClass MUST validate that
match type specified in the Hostname field is appropriate
for the protocol.</p>
<ul>
<li>For the &ldquo;TLS&rdquo; protocol, the Hostname match MUST be
applied to the <a href="https://tools.ietf.org/html/rfc6066#section-3">SNI</a>
server name offered by the client.</li>
<li>For the &ldquo;HTTP&rdquo; protocol, the Hostname match MUST be
applied to the host portion of the
<a href="https://tools.ietf.org/html/rfc7230#section-5.5">effective request URI</a>
or the <a href="https://tools.ietf.org/html/rfc7540#section-8.1.2.3">:authority pseudo-header</a></li>
<li>For the &ldquo;HTTPS&rdquo; protocol, the Hostname match MUST be
applied at both the TLS and HTTP protocol layers.</li>
</ul>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>tls</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayTLSConfig">
GatewayTLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS is the TLS configuration for the Listener. This field
is required if the Protocol field is &ldquo;HTTPS&rdquo; or &ldquo;TLS&rdquo; and
ignored otherwise.</p>
<p>The association of SNIs to Certificate defined in GatewayTLSConfig is
defined based on the Hostname field for this listener:
- &ldquo;Domain&rdquo;: Certificate should be used for the domain and its
first-level subdomains.
- &ldquo;Exact&rdquo;: Certificate should be used for the domain only.
- &ldquo;Any&rdquo;: Certificate in GatewayTLSConfig is the default certificate to use.</p>
<p>The GatewayClass MUST use the longest matching SNI out of all
available certificates for any TLS handshake.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>routes</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteBindingSelector">
RouteBindingSelector
</a>
</em>
</td>
<td>
<p>Routes specifies a schema for associating routes with the
Listener using selectors. A Route is a resource capable of
servicing a request and allows a cluster operator to expose
a cluster resource (i.e. Service) by externally-reachable
URL, load-balance traffic and terminate SSL/TLS.  Typically,
a route is a &ldquo;HTTPRoute&rdquo; or &ldquo;TCPRoute&rdquo; in group
&ldquo;networking.x-k8s.io&rdquo;, however, an implementation may support
other types of resources.</p>
<p>The Routes selector MUST select a set of objects that
are compatible with the application protocol specified in
the Protocol field.</p>
<p>Although a client request may technically match multiple route rules,
only one rule may ultimately receive the request. Matching precedence
MUST be determined in order of the following criteria:</p>
<ul>
<li>The most specific match. For example, the most specific HTTPRoute match
is determined by the longest matching combination of hostname and path.</li>
<li>The oldest Route based on creation timestamp. For example, a Route with
a creation timestamp of &ldquo;2020-09-08 01:02:03&rdquo; is given precedence over
a Route with a creation timestamp of &ldquo;2020-09-08 01:02:04&rdquo;.</li>
<li>If everything else is equivalent, the Route appearing first in
alphabetical order (namespace/name) should be given precedence. For
example, foo/bar is given precedence over foo/baz.</li>
</ul>
<p>All valid portions of a Route selected by this field should be supported.
Invalid portions of a Route can be ignored (sometimes that will mean the
full Route). If a portion of a Route transitions from valid to invalid,
support for that portion of the Route should be dropped to ensure
consistency. For example, even if a filter specified by a Route is
invalid, the rest of the Route should still be supported.</p>
<p>Support: Core</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.ListenerConditionType">ListenerConditionType
(<code>string</code> alias)</p></h3>
<p>
<p>ListenerConditionType is a type of condition associated with the
listener. This type should be used with the ListenerStatus.Conditions
field.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.ListenerStatus">ListenerStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayStatus">GatewayStatus</a>)
</p>
<p>
<p>ListenerStatus is the status associated with a Listener port.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>port</code></br>
<em>
string
</em>
</td>
<td>
<p>Port is the unique Listener port value for which this message
is reporting the status. If more than one Gateway Listener
shares the same port value, this message reports the combined
status of all such Listeners.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<p>Conditions describe the current condition of this listener.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.LocalObjectReference">LocalObjectReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassSpec">GatewayClassSpec</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardingTarget">HTTPForwardingTarget</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">HTTPRouteFilter</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteHost">HTTPRouteHost</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteMatch">HTTPRouteMatch</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteAction">TCPRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteMatch">TCPRouteMatch</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteAction">TLSRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteMatch">TLSRouteMatch</a>, 
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteAction">UDPRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteMatch">UDPRouteMatch</a>)
</p>
<p>
<p>RouteMatchExtensionObjectReference identifies a route-match extension object
within a known namespace.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>group</code></br>
<em>
string
</em>
</td>
<td>
<p>Group is the API group name of the referent.</p>
</td>
</tr>
<tr>
<td>
<code>resource</code></br>
<em>
string
</em>
</td>
<td>
<p>Resource is the API resource name of the referent.</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the referent.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.PathMatchType">PathMatchType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPPathMatch">HTTPPathMatch</a>)
</p>
<p>
<p>PathMatchType specifies the semantics of how HTTP paths should be compared.
Valid PathMatchType values are:</p>
<ul>
<li>&ldquo;Exact&rdquo;</li>
<li>&ldquo;Prefix&rdquo;</li>
<li>&ldquo;RegularExpression&rdquo;</li>
<li>&ldquo;ImplementationSpecific&rdquo;</li>
</ul>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.ProtocolType">ProtocolType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">Listener</a>)
</p>
<p>
<p>ProtocolType defines the application protocol accepted by a
Listener. Implementations are not required to accept all the
defined protocols. If an implementation does not support a
specified protocol, it should raise a &ldquo;ConditionUnsupportedProtocol&rdquo;
condition for the affected Listener.</p>
<p>Valid ProtocolType values are:</p>
<ul>
<li>&ldquo;HTTP&rdquo;</li>
<li>&ldquo;HTTPS&rdquo;</li>
<li>&ldquo;TLS&rdquo;</li>
<li>&ldquo;TCP&rdquo;</li>
<li>&ldquo;UDP&rdquo;</li>
</ul>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.RouteBindingSelector">RouteBindingSelector
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">Listener</a>)
</p>
<p>
<p>RouteBindingSelector defines a schema for associating routes with the Gateway.
If NamespaceSelector and RouteSelector are defined, only routes matching both
selectors are associated with the Gateway.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>routeNamespaces</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteNamespaces">
RouteNamespaces
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RouteNamespaces indicates in which namespaces Routes should be selected
for this Gateway. This is restricted to the namespace of this Gateway by
default.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>routeSelector</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>RouteSelector specifies a set of route labels used for selecting
routes to associate with the Gateway. If RouteSelector is defined,
only routes matching the RouteSelector are associated with the Gateway.
An empty RouteSelector matches all routes.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>group</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Group is the group of the route resource to select. Omitting the value or specifying
the empty string indicates the networking.x-k8s.io API group.
For example, use the following to select an HTTPRoute:</p>
<p>routes:
resource: httproutes</p>
<p>Otherwise, if an alternative API group is desired, specify the desired
group:</p>
<p>routes:
group: acme.io
resource: fooroutes</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>resource</code></br>
<em>
string
</em>
</td>
<td>
<p>Resource is the API resource name of the route resource to select.</p>
<p>Resource MUST correspond to route resources that are compatible with the
application protocol specified in the Listener&rsquo;s Protocol field.</p>
<p>If an implementation does not support or recognize this
resource type, it SHOULD raise a &ldquo;ConditionInvalidRoutes&rdquo;
condition for the affected Listener.</p>
<p>Support: Core</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.RouteConditionType">RouteConditionType
(<code>string</code> alias)</p></h3>
<p>
<p>RouteConditionType is a type of condition for a route.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.RouteGatewayStatus">RouteGatewayStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.RouteStatus">RouteStatus</a>)
</p>
<p>
<p>RouteGatewayStatus describes the status of a route with respect to an
associated Gateway.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>gatewayRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayReference">
GatewayReference
</a>
</em>
</td>
<td>
<p>GatewayRef is a reference to a Gateway object that is associated with
the route.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<p>Conditions describes the status of the route with respect to the
Gateway.  For example, the &ldquo;Admitted&rdquo; condition indicates whether the
route has been admitted or rejected by the Gateway, and why.  Note
that the route&rsquo;s availability is also subject to the Gateway&rsquo;s own
status conditions and listener status.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.RouteGateways">RouteGateways
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteSpec">HTTPRouteSpec</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteSpec">TCPRouteSpec</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteSpec">TLSRouteSpec</a>, 
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteSpec">UDPRouteSpec</a>)
</p>
<p>
<p>RouteGateways defines which Gateways will be able to use a route. If this
field results in preventing the selection of a Route by a Gateway, an
&ldquo;Admitted&rdquo; condition with a status of false must be set for the Gateway on
that Route.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>allow</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayAllowType">
GatewayAllowType
</a>
</em>
</td>
<td>
<p>Allow indicates which Gateways will be allowed to use this route.
Possible values are:
* All: Gateways in any namespace can use this route.
* FromList: Only Gateways specified in GatewayRefs may use this route.
* SameNamespace: Only Gateways in the same namespace may use this route.</p>
</td>
</tr>
<tr>
<td>
<code>gatewayRefs</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayReference">
[]GatewayReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>GatewayRefs must be specified when Allow is set to &ldquo;FromList&rdquo;. In that
case, only Gateways referenced in this list will be allowed to use this
route. This field is ignored for other values of &ldquo;Allow&rdquo;.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.RouteNamespaces">RouteNamespaces
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.RouteBindingSelector">RouteBindingSelector</a>)
</p>
<p>
<p>RouteNamespaces indicate which namespaces Routes should be selected from.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>namespaceSelector</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>NamespaceSelector is a selector of namespaces that Routes should be
selected from. This is a standard Kubernetes LabelSelector, a label query
over a set of resources. The result of matchLabels and matchExpressions
are ANDed. Controllers must not support Routes in namespaces outside this
selector.</p>
<p>An empty selector (default) indicates that Routes in any namespace can be
selected.</p>
<p>The OnlySameNamespace field takes precedence over this field. This
selector will only take effect when OnlySameNamespace is false.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>onlySameNamespace</code></br>
<em>
bool
</em>
</td>
<td>
<p>OnlySameNamespace is a boolean used to indicate if Route references are
limited to the same Namespace as the Gateway. When true, only Routes
within the same Namespace as the Gateway should be selected.</p>
<p>This field takes precedence over the NamespaceSelector field. That
selector should only take effect when this field is set to false.</p>
<p>Support: Core</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.RouteStatus">RouteStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteStatus">HTTPRouteStatus</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteStatus">TCPRouteStatus</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteStatus">TLSRouteStatus</a>, 
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteStatus">UDPRouteStatus</a>)
</p>
<p>
<p>RouteStatus defines the observed state that is required across
all route types.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGatewayStatus">
[]RouteGatewayStatus
</a>
</em>
</td>
<td>
<p>Gateways is a list of the Gateways that are associated with the
route, and the status of the route with respect to each of these
Gateways. When a Gateway selects this route, the controller that
manages the Gateway should add an entry to this list when the
controller first sees the route and should update the entry as
appropriate when the route is modified.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.SecretsDefaultLocalObjectReference">SecretsDefaultLocalObjectReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayTLSConfig">GatewayTLSConfig</a>)
</p>
<p>
<p>SecretsDefaultLocalObjectReference identifies an API object within a
known namespace that defaults group to core and resource to secrets
if unspecified.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>group</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Group is the group of the referent.  Omitting the value or specifying
the empty string indicates the core API group.  For example, use the
following to specify a secrets resource:</p>
<p>fooRef:
resource: secrets
name: mysecret</p>
<p>Otherwise, if the core API group is not desired, specify the desired
group:</p>
<p>fooRef:
group: acme.io
resource: foos
name: myfoo</p>
</td>
</tr>
<tr>
<td>
<code>resource</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resource is the API resource name of the referent. Omitting the value
or specifying the empty string indicates the secrets resource. For
example, use the following to specify a secrets resource:</p>
<p>fooRef:
name: mysecret</p>
<p>Otherwise, if the secrets resource is not desired, specify the desired
group:</p>
<p>fooRef:
group: acme.io
resource: foos
name: myfoo</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the referent.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.ServicesDefaultLocalObjectReference">ServicesDefaultLocalObjectReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GenericForwardToTarget">GenericForwardToTarget</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardToTarget">HTTPForwardToTarget</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRequestMirrorFilter">HTTPRequestMirrorFilter</a>)
</p>
<p>
<p>ServicesDefaultLocalObjectReference identifies an API object within a
known namespace that defaults group to core and resource to services
if unspecified.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>group</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Group is the group of the referent.  Omitting the value or specifying
the empty string indicates the core API group.  For example, use the
following to specify a service:</p>
<p>fooRef:
resource: services
name: myservice</p>
<p>Otherwise, if the core API group is not desired, specify the desired
group:</p>
<p>fooRef:
group: acme.io
resource: foos
name: myfoo</p>
</td>
</tr>
<tr>
<td>
<code>resource</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resource is the API resource name of the referent. Omitting the value
or specifying the empty string indicates the services resource. For example,
use the following to specify a services resource:</p>
<p>fooRef:
name: myservice</p>
<p>Otherwise, if the services resource is not desired, specify the desired
group:</p>
<p>fooRef:
group: acme.io
resource: foos
name: myfoo</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the referent.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TCPRouteAction">TCPRouteAction
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteRule">TCPRouteRule</a>)
</p>
<p>
<p>TCPRouteAction is the action for a given rule.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>forwardTo</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GenericForwardToTarget">
[]GenericForwardToTarget
</a>
</em>
</td>
<td>
<p>ForwardTo sends requests to the referenced object(s). The
resource may be &ldquo;services&rdquo; (omit or use the empty string for the
group), or an implementation may support other resources (for
example, resource &ldquo;myroutetargets&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;services&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: core</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;action&rdquo; behavior.  The resource may be &ldquo;configmaps&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myrouteactions&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the TCPRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TCPRouteMatch">TCPRouteMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteRule">TCPRouteRule</a>)
</p>
<p>
<p>TCPRouteMatch defines the predicate used to match connections to a
given action.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;match&rdquo; behavior.  The resource may be &ldquo;configmap&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myroutematchers&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the TCPRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TCPRouteRule">TCPRouteRule
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteSpec">TCPRouteSpec</a>)
</p>
<p>
<p>TCPRouteRule is the configuration for a given rule.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>matches</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteMatch">
[]TCPRouteMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Matches define conditions used for matching the rule against
incoming TCP connections.
Each match is independent, i.e. this rule will be matched
if <strong>any</strong> one of the matches is satisfied.</p>
</td>
</tr>
<tr>
<td>
<code>action</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteAction">
TCPRouteAction
</a>
</em>
</td>
<td>
<p>Action defines what happens to the connection.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TCPRouteSpec">TCPRouteSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRoute">TCPRoute</a>)
</p>
<p>
<p>TCPRouteSpec defines the desired state of TCPRoute</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteRule">
[]TCPRouteRule
</a>
</em>
</td>
<td>
<p>Rules are a list of TCP matchers and actions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TCPRouteStatus">TCPRouteStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRoute">TCPRoute</a>)
</p>
<p>
<p>TCPRouteStatus defines the observed state of TCPRoute</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>RouteStatus</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteStatus">
RouteStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>RouteStatus</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TLSModeType">TLSModeType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayTLSConfig">GatewayTLSConfig</a>)
</p>
<p>
<p>TLSModeType type defines behavior of gateway with TLS protocol.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.TLSRouteAction">TLSRouteAction
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteRule">TLSRouteRule</a>)
</p>
<p>
<p>TLSRouteAction is the action for a given rule.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>forwardTo</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GenericForwardToTarget">
[]GenericForwardToTarget
</a>
</em>
</td>
<td>
<p>ForwardTo sends requests to the referenced object(s). The
resource may be &ldquo;services&rdquo; (omit or use the empty string for the
group), or an implementation may support other resources (for
example, resource &ldquo;myroutetargets&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;services&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: core</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;action&rdquo; behavior.  The resource may be &ldquo;configmaps&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myrouteactions&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the TLSRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TLSRouteMatch">TLSRouteMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteRule">TLSRouteRule</a>)
</p>
<p>
<p>TLSRouteMatch defines the predicate used to match connections to a
given action.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>snis</code></br>
<em>
[]string
</em>
</td>
<td>
<p>SNIs defines a set of SNI names that should match against the
SNI attribute of TLS CLientHello message in TLS handshake.</p>
<p>SNI can be &ldquo;precise&rdquo; which is a domain name without the terminating
dot of a network host (e.g. &ldquo;foo.example.com&rdquo;) or &ldquo;wildcard&rdquo;, which is
a domain name prefixed with a single wildcard label (e.g. &ldquo;<em>.example.com&rdquo;).
The wildcard character &lsquo;</em>&rsquo; must appear by itself as the first DNS
label and matches only a single label.
You cannot have a wildcard label by itself (e.g. Host == &ldquo;*&rdquo;).
Requests will be matched against the Host field in the following order:</p>
<ol>
<li>If SNI is precise, the request matches this rule if
the SNI in ClientHello is equal to one of the defined SNIs.</li>
<li>If SNI is a wildcard, then the request matches this rule if
the SNI is to equal to the suffix
(removing the first label) of the wildcard rule.</li>
</ol>
<p>Support: core</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;match&rdquo; behavior.  The resource may be &ldquo;configmap&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myroutematchers&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the TLSRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TLSRouteRule">TLSRouteRule
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteSpec">TLSRouteSpec</a>)
</p>
<p>
<p>TLSRouteRule is the configuration for a given rule.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>matches</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteMatch">
[]TLSRouteMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Matches define conditions used for matching the rule against
incoming TLS handshake.
Each match is independent, i.e. this rule will be matched
if <strong>any</strong> one of the matches is satisfied.</p>
</td>
</tr>
<tr>
<td>
<code>action</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteAction">
TLSRouteAction
</a>
</em>
</td>
<td>
<p>Action defines what happens to the connection.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TLSRouteSpec">TLSRouteSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRoute">TLSRoute</a>)
</p>
<p>
<p>TLSRouteSpec defines the desired state of TLSRoute</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRouteRule">
[]TLSRouteRule
</a>
</em>
</td>
<td>
<p>Rules are a list of TLS matchers and actions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TLSRouteStatus">TLSRouteStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TLSRoute">TLSRoute</a>)
</p>
<p>
<p>TLSRouteStatus defines the observed state of TLSRoute</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>RouteStatus</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteStatus">
RouteStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>RouteStatus</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TargetPort">TargetPort
(<code>int32</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GenericForwardToTarget">GenericForwardToTarget</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardToTarget">HTTPForwardToTarget</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRequestMirrorFilter">HTTPRequestMirrorFilter</a>)
</p>
<p>
<p>TargetPort specifies the destination port number to use for a TargetRef.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.TargetWeight">TargetWeight
(<code>int32</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GenericForwardToTarget">GenericForwardToTarget</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPForwardToTarget">HTTPForwardToTarget</a>)
</p>
<p>
<p>TargetWeight specifies weight used for making a forwarding decision
to a TargetRef.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.UDPRouteAction">UDPRouteAction
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteRule">UDPRouteRule</a>)
</p>
<p>
<p>UDPRouteAction is the action for a given rule.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>forwardTo</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GenericForwardToTarget">
GenericForwardToTarget
</a>
</em>
</td>
<td>
<p>ForwardTo sends requests to the referenced object.  The
resource may be &ldquo;services&rdquo; (omit or use the empty string for the
group), or an implementation may support other resources (for
example, resource &ldquo;myroutetargets&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;services&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the UDPRoute will be true.</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;action&rdquo; behavior.  The resource may be &ldquo;configmaps&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myrouteactions&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the UDPRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.UDPRouteMatch">UDPRouteMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteRule">UDPRouteRule</a>)
</p>
<p>
<p>UDPRouteMatch defines the predicate used to match packets to a
given action.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.LocalObjectReference">
LocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;match&rdquo; behavior.  The resource may be &ldquo;configmap&rdquo; (use the empty
string for the group) or an implementation-defined resource (for
example, resource &ldquo;myroutematchers&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the UDPRoute will be true.</p>
<p>Support: custom</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.UDPRouteRule">UDPRouteRule
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteSpec">UDPRouteSpec</a>)
</p>
<p>
<p>UDPRouteRule is the configuration for a given rule.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>matches</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteMatch">
[]UDPRouteMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Matches defines which packets match this rule.</p>
</td>
</tr>
<tr>
<td>
<code>action</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteAction">
UDPRouteAction
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Action defines what happens to the packet.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.UDPRouteSpec">UDPRouteSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRoute">UDPRoute</a>)
</p>
<p>
<p>UDPRouteSpec defines the desired state of UDPRoute.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRouteRule">
[]UDPRouteRule
</a>
</em>
</td>
<td>
<p>Rules are a list of UDP matchers and actions.</p>
</td>
</tr>
<tr>
<td>
<code>gateways</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteGateways">
RouteGateways
</a>
</em>
</td>
<td>
<p>Gateways defines which Gateways can use this Route.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.UDPRouteStatus">UDPRouteStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.UDPRoute">UDPRoute</a>)
</p>
<p>
<p>UDPRouteStatus defines the observed state of UDPRoute.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>RouteStatus</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteStatus">
RouteStatus
</a>
</em>
</td>
<td>
<p>
(Members of <code>RouteStatus</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>.
</em></p>
