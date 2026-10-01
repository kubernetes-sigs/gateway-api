/*
Copyright The Kubernetes Authors.

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
)

// Exercise all batches in the last allowed rule, with a valid total match count.
// This catches both a missing batch and a transformation attached to the wrong
// schema scope. Existing method/path suites cover defaults and match types.
func TestBatchedMatchValidation(t *testing.T) {
	for index := range 64 {
		t.Run(fmt.Sprintf("http/index-%d", index), func(t *testing.T) {
			rules := make([]gatewayv1.HTTPRouteRule, 64)
			for range 64 {
				rules[63].Matches = append(rules[63].Matches, gatewayv1.HTTPRouteMatch{
					Path: &gatewayv1.HTTPPathMatch{Type: new(gatewayv1.PathMatchExact), Value: new("/valid")},
				})
			}
			rules[63].Matches[index].Path.Value = new("/invalid?")
			route := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("batch-http-%d", time.Now().UnixNano()), Namespace: metav1.NamespaceDefault},
				Spec:       gatewayv1.HTTPRouteSpec{Rules: rules},
			}
			validateHTTPRoute(t, route, []string{"spec.rules[63]", "must only contain valid characters"})
		})
		for _, field := range []string{"service", "method"} {
			t.Run(fmt.Sprintf("grpc/%s/index-%d", field, index), func(t *testing.T) {
				rules := grpcRulesWithMethodMatch(64, -1, "", "")
				rules[63].Matches = nil
				for range 64 {
					rules[63].Matches = append(rules[63].Matches, gatewayv1.GRPCRouteMatch{
						Method: &gatewayv1.GRPCMethodMatch{Type: new(gatewayv1.GRPCMethodMatchExact), Service: new("foo"), Method: new("bar")},
					})
				}
				if field == "service" {
					rules[63].Matches[index].Method.Service = new("invalid!")
				} else {
					rules[63].Matches[index].Method.Method = new("invalid!")
				}
				route := &gatewayv1.GRPCRoute{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("batch-grpc-%d", time.Now().UnixNano()), Namespace: metav1.NamespaceDefault},
					Spec:       gatewayv1.GRPCRouteSpec{Rules: rules},
				}
				validateGRPCRoute(t, route, []string{"spec.rules[63]", field + " must only contain valid characters"})
			})
		}
	}
}

func TestRoute129Matches(t *testing.T) {
	httpRules := make([]gatewayv1.HTTPRouteRule, 3)
	grpcRules := make([]gatewayv1.GRPCRouteRule, 3)
	for i, count := range []int{64, 64, 1} {
		for range count {
			httpRules[i].Matches = append(httpRules[i].Matches, gatewayv1.HTTPRouteMatch{
				Path: &gatewayv1.HTTPPathMatch{Type: new(gatewayv1.PathMatchExact), Value: new("/valid")},
			})
			grpcRules[i].Matches = append(grpcRules[i].Matches, gatewayv1.GRPCRouteMatch{
				Method: &gatewayv1.GRPCMethodMatch{Service: new("foo"), Method: new("bar")},
			})
		}
	}
	wantErrors := []string{"total number of matches across all rules in a route must be at most 128"}
	validateHTTPRoute(t, &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("matches-http-%d", time.Now().UnixNano()), Namespace: metav1.NamespaceDefault},
		Spec:       gatewayv1.HTTPRouteSpec{Rules: httpRules},
	}, wantErrors)
	validateGRPCRoute(t, &gatewayv1.GRPCRoute{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("matches-grpc-%d", time.Now().UnixNano()), Namespace: metav1.NamespaceDefault},
		Spec:       gatewayv1.GRPCRouteSpec{Rules: grpcRules},
	}, wantErrors)
}

// grpcRulesWithMethodMatch returns nRules rules with one valid method match each.
// When badIdx >= 0, the first rule instead gets badIdx+1 matches and the match at
// badIdx uses badService/badMethod.
func grpcRulesWithMethodMatch(nRules, badIdx int, badService, badMethod string) []gatewayv1.GRPCRouteRule {
	valid := gatewayv1.GRPCRouteMatch{Method: &gatewayv1.GRPCMethodMatch{
		Type:    new(gatewayv1.GRPCMethodMatchExact),
		Service: new("foo"),
		Method:  new("bar"),
	}}
	var rules []gatewayv1.GRPCRouteRule
	for range nRules {
		rules = append(rules, gatewayv1.GRPCRouteRule{Matches: []gatewayv1.GRPCRouteMatch{valid}})
	}
	if badIdx >= 0 {
		rules[0].Matches = nil
		for i := 0; i <= badIdx; i++ {
			m := valid
			if i == badIdx {
				m = gatewayv1.GRPCRouteMatch{Method: &gatewayv1.GRPCMethodMatch{
					Type:    new(gatewayv1.GRPCMethodMatchExact),
					Service: new(badService),
					Method:  new(badMethod),
				}}
			}
			rules[0].Matches = append(rules[0].Matches, m)
		}
	}
	return rules
}

func validateGRPCRoute(t *testing.T, route *gatewayv1.GRPCRoute, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, route)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating GRPCRoute %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating GRPCRoute %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, missingErrorStrings)
	}
}
