# GEP-1494: Auth in Gateway API

* Issue: [#1494](https://github.com/kubernetes-sigs/gateway-api/issues/1494)
* Status: Provisional

(See status definitions [here](/geps/overview/#gep-states).)

## TLDR

Allow authentication works as a policy that can  be attached to both the Gateway and HTTPRoute.

## Goals

1). In the default policy, it can add an ExtensionFilter that specifies the auth so that Gateway can be attached to.
2). Individual HTTPRoutes can choose to opt out by specifying an empty extension, or a null extension. 
3). It can also have disabled authentication.


## Non-Goals

1). How authentication is implemented.
2). No authorization is required in this GEP.

## Introduction

The Gateway API aims to establish a centralized authentication framework that impacts both the Gateway and HTTPRoute. This solution is designed to achieve the following objectives:

1). The Gateway API should support multiple authentication methods, including Multi-Platform Authentication, traditional Authentication, and no authentication at all. Multi-Platform Authentication includes Single Sign-On (SSO), OAuth, JWT tokens, and more, while traditional Authentication involves the use of usernames and passwords.

2). The authentication policy attachment mechanism should work at both the Gateway and HTTPRoute levels. The following conditions are considered:

- Inherited Policy attachment: attach the Policy to a Gateway and default auth settings (but still allow them to be overridden by individual HTTPRoutes)
- Direct Policy attachment: attach the Policy to a Gateway and override auth settings (which will prevent them being changed on individual HTTPRoutes)
- Not have a Policy at all and just set the settings on individual HTTPRoutes


## API

(... details, can point to PR with changes)

## Conformance Details

(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)

## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)

