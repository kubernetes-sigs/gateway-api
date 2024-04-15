The following tables are populated from the conformance reports uploaded by project implementations. They are separated into the extended features that each project supports listed in their reports.

| Features                                  | Kong               | cilium             | envoyproxy         | istio              | kumahq             | nginxinc           | projectcontour     | solo.io            |
|:------------------------------------------|:-------------------|:-------------------|:-------------------|:-------------------|:-------------------|:-------------------|:-------------------|:-------------------|
| HTTPRouteBackendRequestHeaderModification | :x:                | :x:                | :x:                | :x:                | :x:                | :x:                | :x:                | :x:                |
| HTTPRouteQueryParamMatching               | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| HTTPRouteMethodMatching                   | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| HTTPRouteResponseHeaderModification       | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                | :white_check_mark: | :white_check_mark: |
| HTTPRoutePortRedirect                     | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| HTTPRouteSchemeRedirect                   | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: |
| HTTPRoutePathRedirect                     | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                | :white_check_mark: | :white_check_mark: |
| HTTPRouteHostRewrite                      | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                |
| HTTPRoutePathRewrite                      | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                |
| HTTPRouteRequestMirror                    | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                | :white_check_mark: | :x:                |
| HTTPRouteRequestMultipleMirrors           | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                | :x:                | :white_check_mark: | :x:                |
| HTTPRouteRequestTimeout                   | :x:                | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                | :x:                | :white_check_mark: | :x:                |
| HTTPRouteBackendTimeout                   | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                | :x:                | :white_check_mark: | :x:                |
| HTTPRouteParentRefPort                    | :x:                | :white_check_mark: | :x:                | :x:                | :x:                | :x:                | :x:                | :x:                |

