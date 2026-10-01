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
	"strconv"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/proto"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// A leaf validation is charged for rules.maxItems * matches.maxItems executions.
// Moving it to the rule and checking literal batches of match indexes divides
// that cost among separate expressions. It does not reduce the total CRD cost.
// Keep the batch size stable across generator dependency upgrades; CRD install
// tests on supported Kubernetes versions verify both cost budgets.
const matchValidationBatchSize = 16

// batchRouteMatchValidations lowers only the three conditional regex checks
// whose cost exceeds the per-expression budget at 64 route rules. The original
// kubebuilder annotations remain the source of the predicate and regex. This
// deliberately is not a general CEL optimizer: unexpected shapes fail generation
// so a source edit cannot silently drop validation or change its scope.
func batchRouteMatchValidations(kind string, root *apiext.JSONSchemaProps) error {
	var object string
	var fields []string
	switch kind {
	case "HTTPRoute":
		object, fields = "path", []string{"value"}
	case "GRPCRoute":
		object, fields = "method", []string{"service", "method"}
	default:
		return nil
	}
	spec := root.Properties["spec"]
	rules := spec.Properties["rules"]
	if rules.Items == nil || rules.Items.Schema == nil {
		return fmt.Errorf("spec.rules: expected an array of route rules")
	}
	rule := rules.Items.Schema
	matches := rule.Properties["matches"]
	if matches.Items == nil || matches.Items.Schema == nil || matches.MaxItems == nil || *matches.MaxItems <= 0 {
		return fmt.Errorf("spec.rules[].matches: expected an array with positive maxItems")
	}
	leaf, ok := matches.Items.Schema.Properties[object]
	if !ok || leaf.Type != "object" || leaf.Nullable {
		return fmt.Errorf("spec.rules[].matches[].%s: expected a non-nullable object", object)
	}
	env, err := cel.NewEnv()
	if err != nil {
		return err
	}
	// Build the replacement before changing the schema, including all gRPC checks.
	remaining := append(apiext.ValidationRules(nil), leaf.XValidations...)
	var generated apiext.ValidationRules
	for _, field := range fields {
		index := -1
		var expression *exprpb.Expr
		for i, validation := range remaining {
			ast, issues := env.Parse(validation.Rule)
			if issues.Err() != nil {
				return fmt.Errorf("%s: parsing validation: %w", object, issues.Err())
			}
			parsed, err := cel.AstToParsedExpr(ast)
			if err != nil {
				return err
			}
			if !isConditionalRegex(parsed.GetExpr(), field) {
				continue
			}
			if index != -1 {
				return fmt.Errorf("%s.%s: ambiguous conditional regex validations", object, field)
			}
			index, expression = i, parsed.GetExpr()
		}
		if index == -1 {
			return fmt.Errorf("%s.%s: expected one validation of the form condition ? self.%s.matches(literal) : true", object, field, field)
		}
		validation := remaining[index]
		if validation.MessageExpression != "" || validation.FieldPath != "" || validation.OptionalOldSelf != nil {
			return fmt.Errorf("%s.%s: batching does not support messageExpression, fieldPath, or optionalOldSelf", object, field)
		}
		replacement, issues := env.Parse("self.matches[i]." + object)
		if issues.Err() != nil {
			return issues.Err()
		}
		parsedReplacement, err := cel.AstToParsedExpr(replacement)
		if err != nil {
			return err
		}
		expression, err = rebaseValidationSelf(expression, parsedReplacement.GetExpr())
		if err != nil {
			return fmt.Errorf("%s.%s: %w", object, field, err)
		}
		predicate, err := cel.AstToString(cel.ParsedExprToAst(&exprpb.ParsedExpr{Expr: expression}))
		if err != nil {
			return fmt.Errorf("%s.%s: rendering validation: %w", object, field, err)
		}
		for start := int64(0); start < *matches.MaxItems; start += matchValidationBatchSize {
			var indexes []string
			for i := start; i < min(start+matchValidationBatchSize, *matches.MaxItems); i++ {
				indexes = append(indexes, strconv.FormatInt(i, 10))
			}
			batch := validation
			batch.Rule = fmt.Sprintf("!has(self.matches) || [%s].all(i, self.matches.size() <= i || !has(self.matches[i].%s) || (%s))", strings.Join(indexes, ","), object, predicate)
			batch.Message = "matches[]." + object + ": " + validation.Message
			generated = append(generated, batch)
		}
		remaining = append(remaining[:index], remaining[index+1:]...)
	}
	leaf.XValidations = remaining
	matches.Items.Schema.Properties[object] = leaf
	rule.XValidations = append(rule.XValidations, generated...)
	return nil
}

// Match structure, not the regex text, message, or annotation position. Predicate
// and regex edits automatically flow to the generated rules.
func isConditionalRegex(e *exprpb.Expr, field string) bool {
	conditional := e.GetCallExpr()
	if conditional == nil || conditional.GetFunction() != operators.Conditional || len(conditional.GetArgs()) != 3 || !conditional.GetArgs()[2].GetConstExpr().GetBoolValue() {
		return false
	}
	call := conditional.GetArgs()[1].GetCallExpr()
	if call == nil || call.GetFunction() != "matches" || len(call.GetArgs()) != 1 {
		return false
	}
	if _, ok := call.GetArgs()[0].GetConstExpr().GetConstantKind().(*exprpb.Constant_StringValue); !ok {
		return false
	}
	selection := call.GetTarget().GetSelectExpr()
	return selection != nil && !selection.GetTestOnly() && selection.GetField() == field && selection.GetOperand().GetIdentExpr().GetName() == "self"
}

// Rewrite identifiers in the syntax tree, leaving strings and selected field
// names untouched. Comprehensions and non-self variables (including oldSelf)
// are rejected: moving them would require scope and transition-rule handling.
// has() is a test-only select in the AST and is preserved by the unparser.
func rebaseValidationSelf(e, replacement *exprpb.Expr) (*exprpb.Expr, error) {
	switch node := e.GetExprKind().(type) {
	case *exprpb.Expr_ConstExpr:
	case *exprpb.Expr_IdentExpr:
		if node.IdentExpr.GetName() != "self" {
			return nil, fmt.Errorf("unsupported variable %q in batched validation", node.IdentExpr.GetName())
		}
		return proto.Clone(replacement).(*exprpb.Expr), nil
	case *exprpb.Expr_SelectExpr:
		operand, err := rebaseValidationSelf(node.SelectExpr.GetOperand(), replacement)
		if err != nil {
			return nil, err
		}
		node.SelectExpr.Operand = operand
	case *exprpb.Expr_CallExpr:
		if node.CallExpr.GetTarget() != nil {
			target, err := rebaseValidationSelf(node.CallExpr.GetTarget(), replacement)
			if err != nil {
				return nil, err
			}
			node.CallExpr.Target = target
		}
		for i, arg := range node.CallExpr.GetArgs() {
			rebased, err := rebaseValidationSelf(arg, replacement)
			if err != nil {
				return nil, err
			}
			node.CallExpr.Args[i] = rebased
		}
	case *exprpb.Expr_ListExpr:
		if len(node.ListExpr.GetOptionalIndices()) != 0 {
			return nil, fmt.Errorf("optional list elements are not supported in batched validation")
		}
		for i, element := range node.ListExpr.GetElements() {
			rebased, err := rebaseValidationSelf(element, replacement)
			if err != nil {
				return nil, err
			}
			node.ListExpr.Elements[i] = rebased
		}
	default:
		return nil, fmt.Errorf("unsupported expression %T in batched validation", e.GetExprKind())
	}
	return e, nil
}
