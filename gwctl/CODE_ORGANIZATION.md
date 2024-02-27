## Code Organization

* **cmd/main.go:** 
    * Entry point for the application.
    * Uses the widely adopted Cobra: [https://github.com/spf13/cobra](https://github.com/spf13/cobra) library to manage subcommands and flags.

* **pkg/cmd/(describe, get):**
    * Contain the starting points for individual subcommands.
    * Tasks:
        * Parse command-line arguments into internal structures.
        * Gather necessary data.
        * Output the data to the console.

* **Conceptual Separation (Model-View):**
    * **pkg/resourcediscovery:** Handles data model definition and construction.
    * **pkg/printer:** Manages formatting and printing the data model to the command line (currently the sole view available). 

* **pkg/policymanager:**
    * Provides logic to work with Policies, Policy CRDs, and facilitates policy merging based on type and hierarchy.

* **pkg/resourcediscovery:**
    * **ResourceModel (pkg/resourcediscovery/resourcemodel.go):**
        * Employs a graph-like structure.
            ```
                                                                                                          +------------+
                                                                                                          |   Policy   |
                                                                                                          +------+-----+
                                                                                                                 |
                                                                                                                 |
                                                                                                                 |
                +------------+                           +--------------+                               TargetRef is Namespace
                |   Policy   |                      +----> GatewayClass <-------+                                |
                +----+-------+                      |    +--------------+       |                                |
                     |                              |                           |            +---------+         |
                     |                              |                           |  +-------->|Namespace|<--------+
                     |                      Parent GatewayClass                 |  |         +---------+
                     |                              |                           |  |
                     |                              |                           |  |
                     |                              |                           |  |
                     |                        +-----+-----+               +-----+--+--+
            TargetRef is HTTPRoute    +------->  Gateway  <---+           |  Gateway  <--+
                     |                |       +-----------+   |           +-----------+  |
                     |                |                       |                          |
                     |                |                       |                          |
                     |           Attached to             Attached to                  Attached to
                     |                |                       |                          |
                     |                |                       |                          |
                     |           +----+------+             +--+--------+            +----+------+
                     +---------->| HTTPRoute |             | HTTPRoute |            | HTTPRoute |
                                 +----+------+             +----+------+            +-----+-----+
                                      |                         |                         |
                                      |                         |                         |
                                  Routes to                 Routes to                 Routes to
                                      |                         |                         |
                                      |       +---------+       |                    +----v----+
                                      +-------> Backend <-------+                    | Backend |
                                              +---------+                            +---------+
            ```
        * Nodes: Resources from the Gateway API (Gateways, HTTPRoutes, etc.).
        * Edges: Connections between resources (e.g., HTTPRoutes linked to Backends).
    * **Discoverer (pkg/resourcediscovery/discoverer.go):** 
        * Builds the `ResourceModel`.
        * **Example (`gwctl describe httproutes -l key1=value1`):** 
            1. `DiscoverResourcesForHTTPRoute(filter)`:
                * Identifies HTTPRoutes with the `key1=value1` label.
                * Inserts them as source nodes into the `ResourceModel`.
            2. **Graph Traversal:**
                * Uses a BFS-like algorithm that begins from the source HTTPRoute nodes.
                * Finds associated Gateways, GatewayClasses, Namespaces, Policies, and Backends

* **pkg/printer:**
    * Receives the `ResourceModel` as input. 
    * Extracts relevant information, arranges it into a printable format, and sends it to the command line.


### Additional Notes:

* The code partially adheres to the Model-View pattern, with `resourcediscovery` managing data representation and `printer` controlling output.
* The `Discoverer` constructs the `ResourceModel` graph, performing BFS-like graph traversals to locate connected resources.
* Future improvements could add alternative output views (e.g., Graphviz visualizations, interactive web interfaces).