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
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/require"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func matchValidationSchema(kind string, maxItems int64) (*apiext.JSONSchemaProps, *apiext.JSONSchemaProps, *apiext.JSONSchemaProps) {
	object := "path"
	description := `<gateway:validation:levelUpValidation="path.value">`
	validations := apiext.ValidationRules{{
		Rule: `self.type in ['Exact', 'PathPrefix'] ? self.value.matches('^/[a-z]+$') : true`, Message: "invalid path",
	}}
	if kind == "GRPCRoute" {
		object = "method"
		description = `<gateway:validation:levelUpValidation="method.service">` + "\n" +
			`<gateway:validation:levelUpValidation="method.method">`
		validations = apiext.ValidationRules{
			{Rule: `(!has(self.type) || self.type == 'Exact') && has(self.service) ? self.service.matches('^[a-z]+$') : true`, Message: "invalid service"},
			{Rule: `(!has(self.type) || self.type == 'Exact') && has(self.method) ? self.method.matches('^[a-z]+$') : true`, Message: "invalid method"},
		}
	}
	leaf := apiext.JSONSchemaProps{Type: "object", XValidations: validations}
	match := &apiext.JSONSchemaProps{Type: "object", Properties: map[string]apiext.JSONSchemaProps{object: leaf}}
	rule := &apiext.JSONSchemaProps{Type: "object", Properties: map[string]apiext.JSONSchemaProps{
		"matches": {Type: "array", MaxItems: &maxItems, Description: description, Items: &apiext.JSONSchemaPropsOrArray{Schema: match}},
	}}
	root := &apiext.JSONSchemaProps{Properties: map[string]apiext.JSONSchemaProps{
		"spec": {Properties: map[string]apiext.JSONSchemaProps{
			"rules": {Type: "array", Items: &apiext.JSONSchemaPropsOrArray{Schema: rule}},
		}},
	}}
	return root, rule, match
}

// levelUpSchema runs the two generator steps the way main does, collecting the tags
// before anything strips them and applying them afterwards.
func levelUpSchema(root *apiext.JSONSchemaProps) error {
	targets, err := collectLevelUpTargets(root)
	if err != nil {
		return err
	}
	return applyLevelUpTargets(root, targets)
}

func setMatchesDescription(rule *apiext.JSONSchemaProps, description string) {
	matches := rule.Properties["matches"]
	matches.Description = description
	rule.Properties["matches"] = matches
}

func TestLevelUpValidation(t *testing.T) {
	for _, kind := range []string{"HTTPRoute", "GRPCRoute"} {
		t.Run(kind, func(t *testing.T) {
			root, rule, match := matchValidationSchema(kind, 17)
			original := root.DeepCopy()
			object, count := "path", 2
			if kind == "GRPCRoute" {
				object, count = "method", 4
			}
			leaf := match.Properties[object]
			other := apiext.ValidationRule{Rule: "true", Message: "unrelated leaf validation"}
			leaf.XValidations = append(leaf.XValidations, other)
			match.Properties[object] = leaf
			rule.XValidations = apiext.ValidationRules{other}
			require.NoError(t, levelUpSchema(root))
			require.Equal(t, apiext.ValidationRules{other}, match.Properties[object].XValidations)
			require.Len(t, rule.XValidations, count+1)
			require.Equal(t, other, rule.XValidations[0])
			require.Contains(t, rule.XValidations[1].Rule, "[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15].all")
			require.Contains(t, rule.XValidations[2].Rule, "[16].all")
			require.Contains(t, rule.XValidations[1].Message, object+" in matches 0-15")
			require.Contains(t, rule.XValidations[2].Message, object+" in matches 16")
			require.NoError(t, levelUpSchema(original))
			require.Equal(t, rule.XValidations[1:], original.Properties["spec"].Properties["rules"].Items.Schema.XValidations)
		})
	}

	t.Run("untagged schema is a no-op", func(t *testing.T) {
		root, rule, _ := matchValidationSchema("HTTPRoute", 17)
		setMatchesDescription(rule, "")
		original := root.DeepCopy()
		require.NoError(t, levelUpSchema(root))
		require.Equal(t, original, root)
	})
}

func TestLevelUpValidationErrors(t *testing.T) {
	tests := []struct {
		name, expression string
		modify           func(*apiext.JSONSchemaProps, *apiext.JSONSchemaProps)
	}{
		{name: "missing check", expression: "true"},
		{name: "invalid expression", expression: "self."},
		{name: "changed shape", expression: `self.value.matches('x')`},
		{name: "transition", expression: `self.type == oldSelf.type ? self.value.matches('x') : true`},
		{name: "comprehension", expression: `[1].all(x, x > 0) ? self.value.matches('x') : true`},
		{name: "ambiguous", modify: func(_, match *apiext.JSONSchemaProps) {
			leaf := match.Properties["path"]
			leaf.XValidations = append(leaf.XValidations, leaf.XValidations[0])
			match.Properties["path"] = leaf
		}},
		{name: "message expression", modify: func(_, match *apiext.JSONSchemaProps) {
			leaf := match.Properties["path"]
			leaf.XValidations[0].MessageExpression = "'invalid'"
			match.Properties["path"] = leaf
		}},
		{name: "field path", modify: func(_, match *apiext.JSONSchemaProps) {
			leaf := match.Properties["path"]
			leaf.XValidations[0].FieldPath = ".value"
			match.Properties["path"] = leaf
		}},
		{name: "optional old self", modify: func(_, match *apiext.JSONSchemaProps) {
			leaf := match.Properties["path"]
			leaf.XValidations[0].OptionalOldSelf = new(true)
			match.Properties["path"] = leaf
		}},
		{name: "missing bound", modify: func(rule, _ *apiext.JSONSchemaProps) {
			matches := rule.Properties["matches"]
			matches.MaxItems = nil
			rule.Properties["matches"] = matches
		}},
		{name: "nullable object", modify: func(_, match *apiext.JSONSchemaProps) {
			leaf := match.Properties["path"]
			leaf.Nullable = true
			match.Properties["path"] = leaf
		}},
		{name: "unterminated tag", modify: func(rule, _ *apiext.JSONSchemaProps) {
			setMatchesDescription(rule, `<gateway:validation:levelUpValidation="path.value"`)
		}},
		{name: "unquoted tag value", modify: func(rule, _ *apiext.JSONSchemaProps) {
			setMatchesDescription(rule, `<gateway:validation:levelUpValidation=path.value>`)
		}},
		{name: "leaf path too short", modify: func(rule, _ *apiext.JSONSchemaProps) {
			setMatchesDescription(rule, `<gateway:validation:levelUpValidation="value">`)
		}},
		{name: "leaf path too long", modify: func(rule, _ *apiext.JSONSchemaProps) {
			setMatchesDescription(rule, `<gateway:validation:levelUpValidation="path.value.extra">`)
		}},
		{name: "unknown object", modify: func(rule, _ *apiext.JSONSchemaProps) {
			setMatchesDescription(rule, `<gateway:validation:levelUpValidation="missing.value">`)
		}},
		{name: "unknown field", modify: func(rule, _ *apiext.JSONSchemaProps) {
			setMatchesDescription(rule, `<gateway:validation:levelUpValidation="path.missing">`)
		}},
		{name: "tagged list without items", modify: func(rule, _ *apiext.JSONSchemaProps) {
			matches := rule.Properties["matches"]
			matches.Items = nil
			rule.Properties["matches"] = matches
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, rule, match := matchValidationSchema("HTTPRoute", 64)
			if tc.expression != "" {
				leaf := match.Properties["path"]
				leaf.XValidations[0].Rule = tc.expression
				match.Properties["path"] = leaf
			}
			if tc.modify != nil {
				tc.modify(rule, match)
			}
			original := root.DeepCopy()
			require.Error(t, levelUpSchema(root))
			require.Equal(t, original, root, "failed translation must not modify the schema")
		})
	}
}

func TestCollectLevelUpTargetsIsDeterministic(t *testing.T) {
	root, rule, _ := matchValidationSchema("GRPCRoute", 64)
	rule.Properties["backendRefs"] = apiext.JSONSchemaProps{
		Type:        "array",
		Description: `<gateway:validation:levelUpValidation="filters.type">`,
	}
	first, err := collectLevelUpTargets(root)
	require.NoError(t, err)
	require.Len(t, first, 3)
	for range 20 {
		next, err := collectLevelUpTargets(root)
		require.NoError(t, err)
		require.Equal(t, first, next, "collection must not depend on map iteration order")
	}
}

// Properties holds values rather than pointers, so a missed write-back is invisible to
// assertions on an inner schema. Read everything back through the root instead.
func TestLevelUpValidationWritesBackThroughRoot(t *testing.T) {
	root, _, _ := matchValidationSchema("HTTPRoute", 17)
	require.NoError(t, levelUpSchema(root))
	rule := root.Properties["spec"].Properties["rules"].Items.Schema
	require.Len(t, rule.XValidations, 2)
	require.Empty(t, rule.Properties["matches"].Items.Schema.Properties["path"].XValidations)
}

// A tagged property removed by the channel tweaks must fail loudly, not silently skip.
func TestLevelUpValidationRejectsPrunedProperty(t *testing.T) {
	root, rule, _ := matchValidationSchema("HTTPRoute", 17)
	targets, err := collectLevelUpTargets(root)
	require.NoError(t, err)
	delete(rule.Properties, "matches")
	require.ErrorContains(t, applyLevelUpTargets(root, targets), `property "matches" not found`)
}

func compileValidation(t *testing.T, expression string) cel.Program {
	t.Helper()
	env, err := cel.NewEnv(cel.Variable("self", cel.DynType))
	require.NoError(t, err)
	ast, issues := env.Compile(expression)
	require.NoError(t, issues.Err(), expression)
	program, err := env.Program(ast)
	require.NoError(t, err)
	return program
}

func evalValidation(t *testing.T, program cel.Program, self any) bool {
	t.Helper()
	result, _, err := program.Eval(map[string]any{"self": self})
	require.NoError(t, err)
	value, ok := result.Value().(bool)
	require.True(t, ok)
	return value
}

func TestBatchedValidationEquivalence(t *testing.T) {
	for _, kind := range []string{"HTTPRoute", "GRPCRoute"} {
		t.Run(kind, func(t *testing.T) {
			root, rule, match := matchValidationSchema(kind, 64)
			object := "path"
			cases := []map[string]any{
				{"type": "Exact", "value": "/valid"},
				{"type": "PathPrefix", "value": "/invalid?"},
				{"type": "RegularExpression", "value": ".*"},
				{"type": "Exact", "value": "/invalid?"},
			}
			if kind == "GRPCRoute" {
				object = "method"
				cases = []map[string]any{
					{},
					{"service": "valid"},
					{"method": "valid"},
					{"type": "Exact", "service": "invalid!"},
					{"method": "invalid!"},
					{"type": "RegularExpression", "service": ".*", "method": ".*"},
				}
			}
			var original, generated []cel.Program
			for _, v := range match.Properties[object].XValidations {
				original = append(original, compileValidation(t, v.Rule))
			}
			require.NoError(t, levelUpSchema(root))
			for _, v := range rule.XValidations {
				generated = append(generated, compileValidation(t, v.Rule))
			}
			for _, absent := range []map[string]any{{}, {"matches": []any{}}, {"matches": []any{map[string]any{}}}} {
				for _, p := range generated {
					require.True(t, evalValidation(t, p, absent))
				}
			}
			// Exercise every index and every list length, not just the first batch.
			for index := range 64 {
				for c, value := range cases {
					t.Run(fmt.Sprintf("index-%d/case-%d", index, c), func(t *testing.T) {
						matches := make([]any, index+1)
						for i := range matches {
							matches[i] = map[string]any{}
						}
						matches[index] = map[string]any{object: value}
						want, got := true, true
						for _, p := range original {
							want = evalValidation(t, p, value) && want
						}
						for _, p := range generated {
							got = evalValidation(t, p, map[string]any{"matches": matches}) && got
						}
						require.Equal(t, want, got)
					})
				}
			}
		})
	}
}

func TestBatchingPreservesLiteralSelfAndHas(t *testing.T) {
	root, rule, match := matchValidationSchema("HTTPRoute", 1)
	leaf := match.Properties["path"]
	leaf.XValidations[0].Rule = `has(self.value) ? self.value.matches('^self[.]value$') : true`
	match.Properties["path"] = leaf
	require.NoError(t, levelUpSchema(root))
	program := compileValidation(t, rule.XValidations[0].Rule)
	for _, value := range []map[string]any{{}, {"value": "self.value"}} {
		require.True(t, evalValidation(t, program, map[string]any{"matches": []any{map[string]any{"path": value}}}))
	}
	require.False(t, evalValidation(t, program, map[string]any{"matches": []any{map[string]any{"path": map[string]any{"value": "self.matches[0].path.value"}}}}))
}
