//go:build experimental
// +build experimental

/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	xgatewayv1alpha1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
)

func TestBackendTrafficPolicyConfig(t *testing.T) {
	tests := []struct {
		name               string
		wantErrors         []string
		sessionPersistence xgatewayv1alpha1.SessionPersistence
		retryConstraint    xgatewayv1alpha1.RetryConstraint
	}{
		{
			name: "valid BackendTrafficPolicyConfig no retryConstraint budgetPercent",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Interval: toDuration("10s"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("1s"),
				}),
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTrafficPolicyConfig no retryConstraint budgetInterval",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent: ptrTo(20),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("1s"),
				}),
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTrafficPolicyConfig no retryConstraint minRetryRate",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("10s"),
				}),
			},
			wantErrors: []string{},
		},
		{
			name: "invalid BackendTrafficPolicyConfig budgetInterval too long",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("2h"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("10s"),
				}),
			},
			wantErrors: []string{"interval cannot be greater than one hour or less than one second"},
		},
		{
			name: "invalid BackendTrafficPolicyConfig budgetInterval too short",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("1ms"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("10s"),
				}),
			},
			wantErrors: []string{"interval cannot be greater than one hour or less than one second"},
		},
		{
			name: "invalid BackendTrafficPolicyConfig minRetryRate interval",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("10s"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("2h"),
				}),
			},
			wantErrors: []string{"interval cannot be greater than one hour"},
		},
		{
			name: "valid BackendTrafficPolicyConfig no cookie lifetimeType",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("10s"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("1s"),
				}),
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTrafficPolicyConfig session cookie",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName: ptrTo("foo"),
				Type:        ptrTo(gatewayv1.CookieBasedSessionPersistence),
				CookieConfig: &gatewayv1.CookieConfig{
					LifetimeType: ptrTo(gatewayv1.SessionCookieLifetimeType),
				},
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("10s"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("1s"),
				}),
			},
			wantErrors: []string{},
		},
		{
			name: "invalid BackendTrafficPolicyConfig permanent cookie",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName: ptrTo("foo"),
				Type:        ptrTo(gatewayv1.CookieBasedSessionPersistence),
				CookieConfig: &gatewayv1.CookieConfig{
					LifetimeType: ptrTo(gatewayv1.PermanentCookieLifetimeType),
				},
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("10s"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("1s"),
				}),
			},
			wantErrors: []string{"AbsoluteTimeout must be specified when cookie lifetimeType is Permanent"},
		},
		{
			name: "valid BackendTrafficPolicyConfig permanent cookie",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName:     ptrTo("foo"),
				AbsoluteTimeout: toDuration("1h"),
				Type:            ptrTo(gatewayv1.CookieBasedSessionPersistence),
				CookieConfig: &gatewayv1.CookieConfig{
					LifetimeType: ptrTo(gatewayv1.PermanentCookieLifetimeType),
				},
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("10s"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("1s"),
				}),
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTrafficPolicyConfig header-based session persistence",
			sessionPersistence: xgatewayv1alpha1.SessionPersistence{
				SessionName: ptrTo("foo"),
				Type:        ptrTo(gatewayv1.HeaderBasedSessionPersistence),
			},
			retryConstraint: xgatewayv1alpha1.RetryConstraint{
				Budget: ptrTo(xgatewayv1alpha1.BudgetDetails{
					Percent:  ptrTo(20),
					Interval: toDuration("10s"),
				}),
				MinRetryRate: ptrTo(xgatewayv1alpha1.RequestRate{
					Count:    ptrTo(10),
					Interval: toDuration("1s"),
				}),
			},
			wantErrors: []string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trafficPolicy := &xgatewayv1alpha1.XBackendTrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: xgatewayv1alpha1.BackendTrafficPolicySpec{
					TargetRefs: []xgatewayv1alpha1.LocalPolicyTargetReference{{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					}},
					RetryConstraint:    &tc.retryConstraint,
					SessionPersistence: &tc.sessionPersistence,
				},
			}
			validateBackendTrafficPolicy(t, trafficPolicy, tc.wantErrors)
		})
	}
}

func validateBackendTrafficPolicy(t *testing.T, trafficPolicy *xgatewayv1alpha1.XBackendTrafficPolicy, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, trafficPolicy)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating XBackendTrafficPolicy %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", trafficPolicy.Namespace, trafficPolicy.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating XBackendTrafficPolicy %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", trafficPolicy.Namespace, trafficPolicy.Name), err, missingErrorStrings)
	}
}
