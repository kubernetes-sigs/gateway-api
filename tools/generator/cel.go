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
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// How many list indexes each leveled-up expression covers. Keep it stable across
// generator dependency upgrades; CRD install tests on supported Kubernetes versions
// verify both cost budgets.
const matchValidationBatchSize = 16

// levelUpTagPrefix introduces <gateway:validation:levelUpValidation="<object>.<field>">
// on a list field. The tagged validation is hoisted off the list items and onto the
// object that owns the list, which is how a check whose cost exceeds the
// per-expression budget stays installable. Tags are collected before the channel
// tweaks strip them from descriptions, and applied afterwards.
const levelUpTagPrefix = `<gateway:validation:levelUpValidation=`

// levelUpTagPrefixes returns the tag forms that apply to a channel: the unscoped
// one above and <gateway:<channel>:validation:levelUpValidation=...>, which mirrors
// the other channel-scoped validation markers.
func levelUpTagPrefixes(channel string) []string {
	return []string{levelUpTagPrefix, fmt.Sprintf("<gateway:%s:validation:levelUpValidation=", channel)}
}

// levelUpTarget is one tagged list and the leaf validation to hoist out of it.
type levelUpTarget struct {
	path []string // property names from the CRD root to the tagged list
	leaf []string // object and field names relative to the list item schema
}

// collectLevelUpTargets reads every levelUpValidation tag in the schema that applies
// to the channel. Properties are visited in name order so generation does not depend
// on map iteration order; tags on one property keep the order they were written in,
// unscoped ones first.
func collectLevelUpTargets(root *apiext.JSONSchemaProps, channel string) ([]levelUpTarget, error) {
	var targets []levelUpTarget
	var walk func(node *apiext.JSONSchemaProps, path []string) error
	walk = func(node *apiext.JSONSchemaProps, path []string) error {
		for _, prefix := range levelUpTagPrefixes(channel) {
			if !strings.Contains(node.Description, prefix) {
				continue
			}
			tags, err := experimentalValidationMarkers(node.Description, prefix)
			if err != nil {
				return err
			}
			for _, tag := range tags {
				value, err := strconv.Unquote(tag)
				if err != nil {
					return fmt.Errorf("%s: invalid levelUpValidation tag %s: %w", strings.Join(path, "."), tag, err)
				}
				leaf := strings.Split(value, ".")
				if len(leaf) != 2 {
					return fmt.Errorf("%s: expected a levelUpValidation path of the form <object>.<field>, got %q", strings.Join(path, "."), value)
				}
				targets = append(targets, levelUpTarget{path: slices.Clone(path), leaf: leaf})
			}
		}
		for _, name := range slices.Sorted(maps.Keys(node.Properties)) {
			child := node.Properties[name]
			if err := walk(&child, append(slices.Clone(path), name)); err != nil {
				return err
			}
		}
		if node.Items != nil && node.Items.Schema != nil {
			return walk(node.Items.Schema, path)
		}
		return nil
	}
	if err := walk(root, nil); err != nil {
		return nil, err
	}
	return targets, nil
}

// applyLevelUpTargets hoists every collected validation. A tagged property that the
// channel pruned is an error rather than a silent skip.
func applyLevelUpTargets(root *apiext.JSONSchemaProps, targets []levelUpTarget) error {
	for _, target := range targets {
		if err := applyLevelUpTarget(root, target.path, target.leaf); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(target.path, "."), err)
		}
	}
	return nil
}

// applyLevelUpTarget walks to the tagged list. Properties holds values rather than
// pointers, so every step writes the child back after recursing into it.
func applyLevelUpTarget(node *apiext.JSONSchemaProps, path, leaf []string) error {
	if node.Items != nil && node.Items.Schema != nil {
		return applyLevelUpTarget(node.Items.Schema, path, leaf)
	}
	if _, ok := node.Properties[path[0]]; !ok {
		return fmt.Errorf("property %q not found", path[0])
	}
	if len(path) == 1 {
		return levelUpValidation(node, path[0], leaf)
	}
	child := node.Properties[path[0]]
	err := applyLevelUpTarget(&child, path[1:], leaf)
	node.Properties[path[0]] = child
	return err
}

// levelUpValidation moves one conditional regex off the items of parent.<list> and
// onto parent, checking literal batches of list indexes. A leaf validation is charged
// for list.maxItems executions per owning element, so splitting it across expressions
// keeps each one inside the per-expression budget. It does not reduce the total CRD
// cost. The kubebuilder annotation remains the source of the predicate and regex.
// The supported form is: condition ? self.<field>.matches("regex") : true.
// Changes within this form are picked up automatically, provided they use only
// self and no comprehensions (such as all() or map()). For example, replacing
// the conditional with condition && self.<field>.matches("regex"), referring to
// oldSelf, or adding an all() check requires updating this transformation.
// Otherwise, generation fails rather than silently changing validation.
func levelUpValidation(parent *apiext.JSONSchemaProps, list string, leaf []string) error {
	object, field := leaf[0], leaf[1]
	items := parent.Properties[list]
	if items.Items == nil || items.Items.Schema == nil || items.MaxItems == nil || *items.MaxItems <= 0 {
		return fmt.Errorf("%s: expected an array with positive maxItems", list)
	}
	target, ok := items.Items.Schema.Properties[object]
	if !ok || target.Type != "object" || target.Nullable {
		return fmt.Errorf("%s[].%s: expected a non-nullable object", list, object)
	}
	env, err := cel.NewEnv()
	if err != nil {
		return err
	}
	// At the parent level, self must refer to the tagged object of element i.
	replacement, err := parseCELExpression(env, fmt.Sprintf("self.%s[i].%s", list, object))
	if err != nil {
		return err
	}

	// Build the replacement before changing the schema.
	remaining := append(apiext.ValidationRules(nil), target.XValidations...)
	index, expression, err := findConditionalRegex(env, remaining, field)
	if err != nil {
		return fmt.Errorf("%s.%s: %w", object, field, err)
	}
	validation := remaining[index]
	if validation.MessageExpression != "" || validation.FieldPath != "" || validation.OptionalOldSelf != nil {
		return fmt.Errorf("%s.%s: batching does not support messageExpression, fieldPath, or optionalOldSelf", object, field)
	}
	if err := rebaseValidationSelf(expression, replacement); err != nil {
		return fmt.Errorf("%s.%s: %w", object, field, err)
	}
	predicate, err := cel.AstToString(cel.ParsedExprToAst(&exprpb.ParsedExpr{Expr: expression}))
	if err != nil {
		return fmt.Errorf("%s.%s: rendering validation: %w", object, field, err)
	}
	var generated apiext.ValidationRules
	for start := int64(0); start < *items.MaxItems; start += matchValidationBatchSize {
		var indexes []string
		end := min(start+matchValidationBatchSize, *items.MaxItems) - 1
		for i := start; i <= end; i++ {
			indexes = append(indexes, strconv.FormatInt(i, 10))
		}
		// Name the indexes the batch covers, so a failure points at 16 elements
		// rather than at the whole list.
		span := strconv.FormatInt(start, 10)
		if end > start {
			span = fmt.Sprintf("%d-%d", start, end)
		}
		batch := validation
		batch.Rule = fmt.Sprintf("!has(self.%s) || [%s].all(i, self.%s.size() <= i || !has(self.%s[i].%s) || (%s))", list, strings.Join(indexes, ","), list, list, object, predicate)
		batch.Message = fmt.Sprintf("%s in %s %s: %s", object, list, span, validation.Message)
		generated = append(generated, batch)
	}
	target.XValidations = append(remaining[:index], remaining[index+1:]...)
	items.Items.Schema.Properties[object] = target
	parent.XValidations = append(parent.XValidations, generated...)
	return nil
}

func parseCELExpression(env *cel.Env, rule string) (*exprpb.Expr, error) {
	ast, issues := env.Parse(rule)
	if issues.Err() != nil {
		return nil, issues.Err()
	}
	parsed, err := cel.AstToParsedExpr(ast)
	if err != nil {
		return nil, err
	}
	return parsed.GetExpr(), nil
}

// Find exactly one character check for the field; missing or duplicate checks
// indicate that the annotations changed in a way we cannot safely transform.
func findConditionalRegex(env *cel.Env, validations apiext.ValidationRules, field string) (int, *exprpb.Expr, error) {
	index := -1
	var expression *exprpb.Expr
	for i, validation := range validations {
		parsed, err := parseCELExpression(env, validation.Rule)
		if err != nil {
			return -1, nil, fmt.Errorf("parsing validation: %w", err)
		}
		if !isConditionalRegex(parsed, field) {
			continue
		}
		if index != -1 {
			return -1, nil, fmt.Errorf("ambiguous conditional regex validations")
		}
		index, expression = i, parsed
	}
	if index == -1 {
		return -1, nil, fmt.Errorf("expected one validation of the form condition ? self.%s.matches(literal) : true", field)
	}
	return index, expression, nil
}

// Match structure, not the regex text, message, or annotation position. Predicate
// and regex edits automatically flow to the generated rules.
func isConditionalRegex(e *exprpb.Expr, field string) bool {
	conditional := e.GetCallExpr()
	if conditional.GetFunction() != operators.Conditional || len(conditional.GetArgs()) != 3 {
		return false
	}
	if !conditional.GetArgs()[2].GetConstExpr().GetBoolValue() {
		return false
	}
	call := conditional.GetArgs()[1].GetCallExpr()
	if call.GetFunction() != "matches" || len(call.GetArgs()) != 1 {
		return false
	}
	if _, ok := call.GetArgs()[0].GetConstExpr().GetConstantKind().(*exprpb.Constant_StringValue); !ok {
		return false
	}
	selection := call.GetTarget().GetSelectExpr()
	return !selection.GetTestOnly() && selection.GetField() == field &&
		selection.GetOperand().GetIdentExpr().GetName() == "self"
}

// Rewrite identifiers in place in the syntax tree, leaving strings and selected field
// names untouched. Comprehensions and non-self variables (including oldSelf)
// are rejected: moving them would require scope and transition-rule handling.
// has() is a test-only select in the AST and is preserved by the unparser.
func rebaseValidationSelf(e, replacement *exprpb.Expr) error {
	var children []*exprpb.Expr
	switch node := e.GetExprKind().(type) {
	case *exprpb.Expr_ConstExpr:
	case *exprpb.Expr_IdentExpr:
		if node.IdentExpr.GetName() != "self" {
			return fmt.Errorf("unsupported variable %q in batched validation", node.IdentExpr.GetName())
		}
		e.ExprKind = proto.Clone(replacement).(*exprpb.Expr).ExprKind
	case *exprpb.Expr_SelectExpr:
		children = append(children, node.SelectExpr.GetOperand())
	case *exprpb.Expr_CallExpr:
		if node.CallExpr.GetTarget() != nil {
			children = append(children, node.CallExpr.GetTarget())
		}
		children = append(children, node.CallExpr.GetArgs()...)
	case *exprpb.Expr_ListExpr:
		if len(node.ListExpr.GetOptionalIndices()) != 0 {
			return fmt.Errorf("optional list elements are not supported in batched validation")
		}
		children = node.ListExpr.GetElements()
	default:
		return fmt.Errorf("unsupported expression %T in batched validation", e.GetExprKind())
	}
	for _, child := range children {
		if err := rebaseValidationSelf(child, replacement); err != nil {
			return err
		}
	}
	return nil
}
