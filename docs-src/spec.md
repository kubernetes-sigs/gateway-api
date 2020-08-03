<p>Packages:</p>
<ul>
<li>
<a href="#networking.x-k8s.io%2fv1alpha1">networking.x-k8s.io/v1alpha1</a>
</li>
</ul>
<h2 id="networking.x-k8s.io/v1alpha1">networking.x-k8s.io/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains API Schema definitions for the networking v1alpha1 API group</p>
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
<code>class</code></br>
<em>
string
</em>
</td>
<td>
<p>Class used for this Gateway. This is the name of a GatewayClass resource.</p>
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
definitions into single implementation-defined acceptor
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
host name provided in the incoming client request MUST be
matched to a Listener to find the correct set of Routes.
The incoming host name MUST be matched using the Hostname
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
GatewayStatus.Listeners.</p>
<p>If no ListenerAddresses are specified, the GatewayClass may
schedule the Gateway in an implementation-defined manner,
assigning an appropriate set of ListenerAddresses.</p>
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
controller that managing Gateways of this class.</p>
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
<code>allowedGatewayNamespaceSelector</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowedGatewayNamespaceSelector is a selector of namespaces that Gateways
can use this GatewayClass from. This is a standard Kubernetes
LabelSelector, a label query over a set of resources. The result of
matchLabels and matchExpressions are ANDed. Controllers must not support
Gateways in namespaces outside this selector.</p>
<p>An empty selector (default) indicates that Gateways can use this
GatewayClass from any namespace.</p>
<p>When a Gateway attempts to use this class from a namespace that is not
allowed by this selector, the controller implementing the GatewayClass
may add a new &ldquo;ForbiddenNamespaceForClass&rdquo; condition to the Gateway
status. Adding this condition is considered optional since not all
controllers will have access to all namespaces.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>allowedRouteNamespaces</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteNamespaces">
RouteNamespaces
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowedRouteNamespaces indicates in which namespaces Routes can be
selected for Gateways of this class. This is restricted to the namespace
of the Gateway by default.</p>
<p>When any Routes are selected by a Gateway in a namespace that is not
allowed by this selector, the controller implementing the GatewayClass
may add a new &ldquo;ForbiddenRoutesForClass&rdquo; condition to the Gateway status.
Adding this condition is considered optional since not all controllers
will have access to all namespaces.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>parametersRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ParametersRef is a controller specific resource containing
the configuration parameters corresponding to this
class. This is optional if the controller does not require
any additional configuration.</p>
<p>Valid resources for reference are up to the Controller. Examples
include &ldquo;configmaps&rdquo; (omit or specify the empty string for the group
to indicate the core API group) or a custom resource (CRD).  Omitting
or specifying the empty string for both the resource and group
indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent cannot
be found, the GatewayClass&rsquo;s &ldquo;InvalidParameters&rdquo; status condition
will be true.</p>
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
<p>HTTPRoute is the Schema for the httproutes API</p>
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
<code>default</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteHost">
HTTPRouteHost
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default is the default host to use. Default.Hostnames must
be an empty list.</p>
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
<p>TCPRoute is the Schema for the tcproutes API</p>
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
<h3 id="networking.x-k8s.io/v1alpha1.AddressType">AddressType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayAddress">GatewayAddress</a>)
</p>
<p>
<p>AddressType defines how a network address is represented as a text string.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">ConfigMapsDefaultLocalObjectReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassSpec">GatewayClassSpec</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteAction">HTTPRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">HTTPRouteFilter</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteHost">HTTPRouteHost</a>, 
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteMatch">HTTPRouteMatch</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteAction">TCPRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteMatch">TCPRouteMatch</a>)
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
<em>(Optional)</em>
<p>Group is the group of the referent.  Omitting the value or specifying
the empty string indicates the core API group.  For example, use the
following to specify a configmaps:</p>
<p>fooRef:
resource: configmaps
name: myconfigmap</p>
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
or specifying the empty string indicates the configmaps resource. For
example, use the following to specify a configmaps resource:</p>
<p>fooRef:
name: myconfigmap</p>
<p>Otherwise, if the configmaps resource is not desired, specify the desired
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
<h3 id="networking.x-k8s.io/v1alpha1.ForwardToTarget">ForwardToTarget
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteAction">HTTPRouteAction</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteAction">TCPRouteAction</a>)
</p>
<p>
<p>ForwardToTarget identifies a target object within a known namespace.</p>
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
<p>Support: Core (httproute)
Support: Extended (tcproute)</p>
</td>
</tr>
</tbody>
</table>
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
<h3 id="networking.x-k8s.io/v1alpha1.GatewayClassCondition">GatewayClassCondition
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassStatus">GatewayClassStatus</a>)
</p>
<p>
<p>GatewayClassCondition contains the details for the current
condition of this GatewayClass.</p>
<p>Support: Core, unless otherwise specified.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassConditionType">
GatewayClassConditionType
</a>
</em>
</td>
<td>
<p>Type of this condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of this condition.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason is a machine consumable string for the last
transition. It should be a one-word, CamelCase
string. Reason will be defined by the controller.</p>
<p>Support: Custom; values will be controller-specific.
This field must not be empty.</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<p>Message is a human readable reason for last transition.
This field may be empty.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastTransitionTime is the time of the last change to this condition.
This should be when the underlying condition changed.
If that is not known, then using the time when the API field changed is acceptable.</p>
</td>
</tr>
<tr>
<td>
<code>observedGeneration</code></br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>If set, this represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.condition[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayClassConditionType">GatewayClassConditionType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassCondition">GatewayClassCondition</a>)
</p>
<p>
<p>GatewayClassConditionType is the type of status conditions.</p>
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
controller that managing Gateways of this class.</p>
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
<code>allowedGatewayNamespaceSelector</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowedGatewayNamespaceSelector is a selector of namespaces that Gateways
can use this GatewayClass from. This is a standard Kubernetes
LabelSelector, a label query over a set of resources. The result of
matchLabels and matchExpressions are ANDed. Controllers must not support
Gateways in namespaces outside this selector.</p>
<p>An empty selector (default) indicates that Gateways can use this
GatewayClass from any namespace.</p>
<p>When a Gateway attempts to use this class from a namespace that is not
allowed by this selector, the controller implementing the GatewayClass
may add a new &ldquo;ForbiddenNamespaceForClass&rdquo; condition to the Gateway
status. Adding this condition is considered optional since not all
controllers will have access to all namespaces.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>allowedRouteNamespaces</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.RouteNamespaces">
RouteNamespaces
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AllowedRouteNamespaces indicates in which namespaces Routes can be
selected for Gateways of this class. This is restricted to the namespace
of the Gateway by default.</p>
<p>When any Routes are selected by a Gateway in a namespace that is not
allowed by this selector, the controller implementing the GatewayClass
may add a new &ldquo;ForbiddenRoutesForClass&rdquo; condition to the Gateway status.
Adding this condition is considered optional since not all controllers
will have access to all namespaces.</p>
<p>Support: Core</p>
</td>
</tr>
<tr>
<td>
<code>parametersRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ParametersRef is a controller specific resource containing
the configuration parameters corresponding to this
class. This is optional if the controller does not require
any additional configuration.</p>
<p>Valid resources for reference are up to the Controller. Examples
include &ldquo;configmaps&rdquo; (omit or specify the empty string for the group
to indicate the core API group) or a custom resource (CRD).  Omitting
or specifying the empty string for both the resource and group
indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent cannot
be found, the GatewayClass&rsquo;s &ldquo;InvalidParameters&rdquo; status condition
will be true.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassCondition">
[]GatewayClassCondition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions is the current status from the controller for
this GatewayClass.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayCondition">GatewayCondition
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayStatus">GatewayStatus</a>)
</p>
<p>
<p>GatewayCondition is an error status for a given route.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.GatewayConditionType">
GatewayConditionType
</a>
</em>
</td>
<td>
<p>Type indicates the type of condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status describes the current state of this condition. Can be &ldquo;True&rdquo;,
&ldquo;False&rdquo;, or &ldquo;Unknown&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<p>Message is a human-understandable message describing the condition.
This field may be empty.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason indicates why the condition is in this state.
This field must not be empty.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastTransitionTime indicates the last time this condition changed.
This should be when the underlying condition changed.
If that is not known, then using the time when the API field changed is acceptable.</p>
</td>
</tr>
<tr>
<td>
<code>observedGeneration</code></br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>If set, this represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.condition[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayConditionType">GatewayConditionType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayCondition">GatewayCondition</a>)
</p>
<p>
<p>GatewayConditionType is a type of condition associated with a Gateway.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.GatewayObjectReference">GatewayObjectReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteStatus">HTTPRouteStatus</a>, 
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteStatus">TCPRouteStatus</a>)
</p>
<p>
<p>GatewayObjectReference identifies a Gateway object.</p>
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
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Namespace is the namespace of the referent.</p>
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
<h3 id="networking.x-k8s.io/v1alpha1.GatewaySpec">GatewaySpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Gateway">Gateway</a>)
</p>
<p>
<p>GatewaySpec defines the desired state of Gateway.</p>
<p>The Spec is split into two major pieces: listeners describing
client-facing properties and routes that describe application-level
routing.</p>
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
<code>class</code></br>
<em>
string
</em>
</td>
<td>
<p>Class used for this Gateway. This is the name of a GatewayClass resource.</p>
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
definitions into single implementation-defined acceptor
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
host name provided in the incoming client request MUST be
matched to a Listener to find the correct set of Routes.
The incoming host name MUST be matched using the Hostname
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
GatewayStatus.Listeners.</p>
<p>If no ListenerAddresses are specified, the GatewayClass may
schedule the Gateway in an implementation-defined manner,
assigning an appropriate set of ListenerAddresses.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.GatewayCondition">
[]GatewayCondition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions describe the current conditions of the Gateway.</p>
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
<p>Listeners provides status for each unique listener port defined in the Spec.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HTTPHeaderFilter">HTTPHeaderFilter
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">HTTPRouteFilter</a>)
</p>
<p>
<p>HTTPHeaderFilter defines the filter behavior for a request match.</p>
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
<p>Remove the given header(s) on the HTTP request before the
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
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteAction">HTTPRouteAction
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteRule">HTTPRouteRule</a>)
</p>
<p>
<p>HTTPRouteAction is the action taken given a match.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.ForwardToTarget">
[]ForwardToTarget
</a>
</em>
</td>
<td>
<p>ForwardTo sends requests to the referenced object(s).  The
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
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
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
<h3 id="networking.x-k8s.io/v1alpha1.HTTPRouteFilter">HTTPRouteFilter
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteRule">HTTPRouteRule</a>)
</p>
<p>
<p>HTTPRouteFilter defines a filter-like action to be applied to
requests.</p>
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
<code>headers</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPHeaderFilter">
HTTPHeaderFilter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Headers related filters.</p>
<p>Support: extended</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
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
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
<p>Support: custom</p>
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
the HTTP Host header to select a HTTPRoute to process a the request.
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
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ExtensionRef is an optional, implementation-specific extension to the
&ldquo;host&rdquo; block.  The resource may be &ldquo;configmaps&rdquo; (omit or specify the
empty string for the group) or an implementation-defined resource
(for example, resource &ldquo;myroutehosts&rdquo; in group &ldquo;networking.acme.io&rdquo;).
Omitting or specifying the empty string for both the resource and
group indicates that the resource is &ldquo;configmaps&rdquo;.  If the referent
cannot be found, the &ldquo;InvalidRoutes&rdquo; status condition on any Gateway
that includes the HTTPRoute will be true.</p>
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
<p>HTTPRouteMatch defines the predicate used to match requests to a
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
<code>pathMatchType</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.PathMatchType">
PathMatchType
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PathType is defines the semantics of the <code>Path</code> matcher.</p>
<p>Support: core (Exact, Prefix)
Support: extended (RegularExpression)
Support: custom (ImplementationSpecific)</p>
<p>Default: &ldquo;Exact&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<p>Path is the value of the HTTP path as interpreted via
PathType.</p>
<p>Default: &ldquo;/&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>headerMatchType</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>HeaderMatchType defines the semantics of the <code>Header</code> matcher.</p>
<p>Support: core (Exact)
Support: custom (ImplementationSpecific)</p>
<p>Default: &ldquo;Exact&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>headers</code></br>
<em>
map[string]string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Headers are the HTTP Headers to match as interpreted via
HeaderMatchType. Multiple headers are ANDed together, meaning, a request
must contain all the headers specified in order to select this route.</p>
</td>
</tr>
<tr>
<td>
<code>extensionRef</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
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
<p>HTTPRouteRule is the configuration for a given path.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteMatch">
HTTPRouteMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Match defines which requests match this path.</p>
</td>
</tr>
<tr>
<td>
<code>filter</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteFilter">
HTTPRouteFilter
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Filter defines what filters are applied to the request.</p>
</td>
</tr>
<tr>
<td>
<code>action</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteAction">
HTTPRouteAction
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Action defines what happens to the request.</p>
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
<code>default</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteHost">
HTTPRouteHost
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Default is the default host to use. Default.Hostnames must
be an empty list.</p>
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
<code>gatewayRefs</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayObjectReference">
[]GatewayObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.HostnameMatch">HostnameMatch
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">Listener</a>)
</p>
<p>
<p>HostnameMatch specifies how a Listener should match the incoming
host name from a client request. Depending on the incoming protocol,
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
<p>Match specifies how the host name provided by the client should be
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
for host names.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.Listener">Listener
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewaySpec">GatewaySpec</a>)
</p>
<p>
<p>Listener embodies the concept of a logical endpoint where a
Gateway can accept network connections.</p>
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
<p>Hostname specifies to match the virtual host name for
protocol types that define this concept.</p>
<p>Incoming requests that include a host name are matched
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
<a href="#networking.x-k8s.io/v1alpha1.TLSConfig">
TLSConfig
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>TLS is the TLS configuration for the Listener. This field
is required if the Protocol field is &ldquo;HTTPS&rdquo; or &ldquo;TLS&rdquo; and
ignored otherwise.</p>
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
<p>Support: Core</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.ListenerCondition">ListenerCondition
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.ListenerStatus">ListenerStatus</a>)
</p>
<p>
<p>ListenerCondition is an error status for a given listener.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.ListenerConditionType">
ListenerConditionType
</a>
</em>
</td>
<td>
<p>Type indicates the type of condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status describes the current state of this condition. Can be &ldquo;True&rdquo;,
&ldquo;False&rdquo;, or &ldquo;Unknown&rdquo;.</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<p>Message is a human-understandable message describing the condition.
This field may be empty.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason indicates why the condition is in this state.
This field must not be empty.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastTransitionTime indicates the last time this condition changed.
This should be when the underlying condition changed.
If that is not known, then using the time when the API field changed is acceptable.</p>
</td>
</tr>
<tr>
<td>
<code>observedGeneration</code></br>
<em>
int64
</em>
</td>
<td>
<em>(Optional)</em>
<p>If set, this represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.condition[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.ListenerConditionType">ListenerConditionType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.ListenerCondition">ListenerCondition</a>)
</p>
<p>
<p>ListenerConditionType is a type of condition associated with the listener.</p>
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
<a href="#networking.x-k8s.io/v1alpha1.ListenerCondition">
[]ListenerCondition
</a>
</em>
</td>
<td>
<p>Conditions describe the current condition of this listener.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.PathMatchType">PathMatchType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.HTTPRouteMatch">HTTPRouteMatch</a>)
</p>
<p>
<p>PathMatchType specifies the semantics of how HTTP paths should be compared.</p>
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
<h3 id="networking.x-k8s.io/v1alpha1.RouteNamespaces">RouteNamespaces
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayClassSpec">GatewayClassSpec</a>, 
<a href="#networking.x-k8s.io/v1alpha1.RouteBindingSelector">RouteBindingSelector</a>)
</p>
<p>
<p>RouteNamespaces is used by Gateway and GatewayClass to indicate which
namespaces Routes should be selected from.</p>
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
<em>(Optional)</em>
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
<h3 id="networking.x-k8s.io/v1alpha1.SecretsDefaultLocalObjectReference">SecretsDefaultLocalObjectReference
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.TLSConfig">TLSConfig</a>)
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
<a href="#networking.x-k8s.io/v1alpha1.ForwardToTarget">ForwardToTarget</a>)
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
<a href="#networking.x-k8s.io/v1alpha1.ForwardToTarget">
[]ForwardToTarget
</a>
</em>
</td>
<td>
<p>ForwardTo sends requests to the referenced object(s).  The
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
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
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
<a href="#networking.x-k8s.io/v1alpha1.ConfigMapsDefaultLocalObjectReference">
ConfigMapsDefaultLocalObjectReference
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
<code>match</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.TCPRouteMatch">
TCPRouteMatch
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Match defines which connections match this rule.</p>
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
<em>(Optional)</em>
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
<code>gatewayRefs</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.GatewayObjectReference">
[]GatewayObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.x-k8s.io/v1alpha1.TLSConfig">TLSConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.Listener">Listener</a>)
</p>
<p>
<p>TLSConfig describes a TLS configuration.</p>
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
<code>certificateRefs</code></br>
<em>
<a href="#networking.x-k8s.io/v1alpha1.SecretsDefaultLocalObjectReference">
[]SecretsDefaultLocalObjectReference
</a>
</em>
</td>
<td>
<p>CertificateRefs is a list of references to Kubernetes objects that each
contain an identity certificate.  The host name in a TLS SNI client hello
message is used for certificate matching and route host name selection.
The SNI server_name must match a route host name for the Gateway to route
the TLS request.  If an entry in this list omits or specifies the empty
string for both the group and the resource, the resource defaults to &ldquo;secrets&rdquo;.
An implementation may support other resources (for example, resource
&ldquo;mycertificates&rdquo; in group &ldquo;networking.acme.io&rdquo;).</p>
<p>Support: Core (Kubernetes Secrets)
Support: Implementation-specific (Other resource types)</p>
</td>
</tr>
<tr>
<td>
<code>minimumVersion</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>MinimumVersion of TLS allowed. It is recommended to use one of
the TLS<em>* constants above. Note: MinimumVersion is not strongly
typed to allow implementation-specific versions to be used without
requiring updates to the API types. String must be of the form
&ldquo;<protocol><major></em><minor>&rdquo;.</p>
<p>Support: Core for TLS1_{1,2,3}. Implementation-specific for all other
values.</p>
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
<h3 id="networking.x-k8s.io/v1alpha1.TargetPort">TargetPort
(<code>int32</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.ForwardToTarget">ForwardToTarget</a>)
</p>
<p>
<p>TargetPort specifies the destination port number to use for a TargetRef.</p>
</p>
<h3 id="networking.x-k8s.io/v1alpha1.TargetWeight">TargetWeight
(<code>int32</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#networking.x-k8s.io/v1alpha1.ForwardToTarget">ForwardToTarget</a>)
</p>
<p>
<p>TargetWeight specifies weight used for making a forwarding decision
to a TargetRef.</p>
</p>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>.
</em></p>
