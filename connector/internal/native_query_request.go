package internal

import (
	"fmt"

	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
)

// NativeQueryRequest the structured native request which is evaluated from the raw expression
type NativeQueryRequest struct {
	*BaseQueryRequest
	Expression      schema.Expression
	HasValueBoolExp bool
}

// EvalNativeQueryRequest evaluates the requested collection data of the query request
func EvalNativeQueryRequest(request *schema.QueryRequest, arguments map[string]any, variables map[string]any, runtime *metadata.RuntimeSettings) (*NativeQueryRequest, error) {
	result := &NativeQueryRequest{
		BaseQueryRequest: &BaseQueryRequest{
			Variables: variables,
			runtime:   runtime,
		},
	}
	if len(request.Query.Predicate) > 0 {
		newExpr, err := result.evalQueryPredicate(request.Query.Predicate)
		if err != nil {
			return nil, err
		}
		if newExpr != nil {
			result.Expression = newExpr.Encode()
		}
	}

	orderBy, err := evalCollectionOrderBy(request.Query.OrderBy)
	if err != nil {
		return nil, err
	}
	result.OrderBy = orderBy

	return result, nil
}

func (pr *NativeQueryRequest) evalQueryPredicate(expression schema.Expression) (schema.ExpressionEncoder, error) {
	switch expr := expression.Interface().(type) {
	case *schema.ExpressionAnd:
		exprs, err := pr.evalQueryExpressions(expr.Expressions)
		if err != nil {
			return nil, err
		}

		return schema.NewExpressionAnd(exprs...), nil
	case *schema.ExpressionOr:
		exprs, err := pr.evalQueryExpressions(expr.Expressions)
		if err != nil {
			return nil, err
		}

		return schema.NewExpressionOr(exprs...), nil
	case *schema.ExpressionNot, *schema.ExpressionUnaryComparisonOperator:
		return expr, nil
	case *schema.ExpressionBinaryComparisonOperator:
		return pr.evalBinaryComparisonOperator(expr)
	default:
		return nil, fmt.Errorf("unsupported expression: %+v", expression)
	}
}

func (pr *NativeQueryRequest) evalBinaryComparisonOperator(expr *schema.ExpressionBinaryComparisonOperator) (*schema.ExpressionBinaryComparisonOperator, error) {
	if expr.Column.Type != schema.ComparisonTargetTypeColumn {
		return nil, fmt.Errorf("%s: unsupported comparison target `%s`", expr.Column.Name, expr.Column.Type)
	}

	switch expr.Column.Name {
	case metadata.TimestampKey:
		if err := pr.evalTimestampFromBinaryComparisonOperator(expr, pr.Variables, pr.runtime.UnixTimeUnit); err != nil {
			return nil, err
		}
	case metadata.LogLineKey, metadata.MetricValueKey:
		pr.HasValueBoolExp = true

		return expr, nil
	default:
		return expr, nil
	}

	return nil, nil
}

func (pr *NativeQueryRequest) evalQueryExpressions(expressions []schema.Expression) ([]schema.ExpressionEncoder, error) {
	results := []schema.ExpressionEncoder{}
	for _, nestedExpr := range expressions {
		evalExpr, err := pr.evalQueryPredicate(nestedExpr)
		if err != nil {
			return nil, err
		}
		if evalExpr != nil {
			results = append(results, evalExpr)
		}
	}

	return results, nil
}
