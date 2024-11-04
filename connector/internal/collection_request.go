package internal

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

// CollectionRequest the structured predicate result which is evaluated from the raw expression
type CollectionRequest struct {
	*BaseQueryRequest
	LogLine          *schema.ExpressionBinaryComparisonOperator
	MetricValue      *schema.ExpressionBinaryComparisonOperator
	LabelExpressions map[string]LabelExpression
	FieldExpression  schema.ExpressionEncoder
	Aggregations     []KeyValue
}

// EvalCollectionRequest evaluates the requested collection data of the query request
func EvalCollectionRequest(request *schema.QueryRequest, model metadata.ModelInfo, arguments map[string]any, variables map[string]any, runtime *metadata.RuntimeSettings) (*CollectionRequest, error) {
	result := &CollectionRequest{
		BaseQueryRequest: &BaseQueryRequest{
			Variables: variables,
			runtime:   runtime,
		},
	}
	if len(request.Query.Predicate) > 0 {
		var err error
		result.LabelExpressions, result.FieldExpression, err = result.evalQueryPredicate(request.Query.Predicate, model)
		if err != nil {
			return nil, err
		}
	}

	orderBy, err := evalCollectionOrderBy(request.Query.OrderBy)
	if err != nil {
		return nil, err
	}
	result.OrderBy = orderBy

	if len(arguments) == 0 {
		return result, nil
	}

	if err := result.evalArguments(arguments); err != nil {
		return nil, err
	}

	return result, nil
}

func (pr *CollectionRequest) evalArguments(arguments map[string]any) error {
	if err := pr.evalArgumentAggregations(arguments); err != nil {
		return err
	}

	if rawInterval, ok := arguments[metadata.ArgumentKeyInterval]; ok {
		dur, err := utils.DecodeNullableDuration(rawInterval, utils.WithBaseUnix(pr.runtime.UnixTimeUnit.Duration()))
		if err != nil {
			return fmt.Errorf("%s: %w", metadata.ArgumentKeyInterval, err)
		}
		if dur != nil {
			pr.Interval = &scalar.Duration{Duration: *dur}
		}
	}
	if rawStep, ok := arguments[metadata.ArgumentKeyStep]; ok {
		dur, err := utils.DecodeNullableDuration(rawStep, utils.WithBaseUnix(pr.runtime.UnixTimeUnit.Duration()))
		if err != nil {
			return fmt.Errorf("%s: %w", metadata.ArgumentKeyStep, err)
		}
		if dur != nil {
			pr.Step = &scalar.Duration{Duration: *dur}
		}
	}
	if rawFlat, ok := arguments[metadata.ArgumentKeyFlat]; ok {
		flat, err := utils.DecodeNullableBoolean(rawFlat)
		if err != nil {
			return fmt.Errorf("%s: %w", metadata.ArgumentKeyFlat, err)
		}
		pr.Flat = flat
	}

	return nil
}

func (pr *CollectionRequest) evalArgumentAggregations(arguments map[string]any) error {
	if fn, ok := arguments[metadata.ArgumentKeyAggregations]; ok && !utils.IsNil(fn) {
		fnMap := []map[string]any{}
		if err := mapstructure.Decode(fn, &fnMap); err != nil {
			return err
		}
		for fi, f := range fnMap {
			i := 0
			for k, v := range f {
				if i > 0 {
					return errOnlyOneAggregationAllowed
				}
				if utils.IsNil(v) {
					return fmt.Errorf("aggregations[%d].%s: value must be not null", fi, k)
				}
				i++
				pr.Aggregations = append(pr.Aggregations, KeyValue{
					Key:   k,
					Value: v,
				})
			}

			if i == 0 {
				return errOnlyOneAggregationAllowed
			}
		}
	}

	return nil
}

func (pr *CollectionRequest) evalQueryPredicate(expression schema.Expression, model metadata.ModelInfo) (map[string]LabelExpression, schema.ExpressionEncoder, error) {
	switch expr := expression.Interface().(type) {
	case *schema.ExpressionAnd:
		return pr.evalExpressionAnd(expr, model)
	case *schema.ExpressionOr:
		evalExprs := []schema.ExpressionEncoder{}
		labelExpressions := map[string]LabelExpression{}
		for _, nestedExpr := range expr.Expressions {
			labelExprs, evalExpr, err := pr.evalQueryPredicate(nestedExpr, model)
			if err != nil {
				return nil, nil, err
			}
			if len(labelExprs) > 0 {
				return nil, nil, errUnsupportedOrLabels
			}
			if evalExpr != nil {
				evalExprs = append(evalExprs, evalExpr)
			}
		}

		return labelExpressions, schema.NewExpressionOr(evalExprs...), nil
	case *schema.ExpressionBinaryComparisonOperator:
		return pr.evalExpressionBinaryComparisonOperator(expr, model)
	default:
		return nil, nil, fmt.Errorf("unsupported expression: %+v", expression)
	}
}

func (pr *CollectionRequest) evalExpressionAnd(expr *schema.ExpressionAnd, model metadata.ModelInfo) (map[string]LabelExpression, schema.ExpressionEncoder, error) {
	evalExprs := []schema.ExpressionEncoder{}
	labelExpressions := map[string]LabelExpression{}
	for _, nestedExpr := range expr.Expressions {
		labelExprs, evalExpr, err := pr.evalQueryPredicate(nestedExpr, model)
		if err != nil {
			return nil, nil, err
		}
		if evalExpr != nil {
			evalExprs = append(evalExprs, evalExpr)
		}
		if len(labelExprs) == 0 {
			continue
		}
		if len(labelExpressions) > 0 && len(evalExprs) > 0 {
			return nil, nil, errUnsupportedMixedLabelAndFieldExpressions
		}
		labelExpressions = mergeLabelExpressions(labelExpressions, labelExprs)
	}

	return labelExpressions, schema.NewExpressionAnd(evalExprs...), nil
}

func (pr *CollectionRequest) evalExpressionBinaryComparisonOperator(expr *schema.ExpressionBinaryComparisonOperator, model metadata.ModelInfo) (map[string]LabelExpression, schema.ExpressionEncoder, error) {
	if expr.Column.Type != schema.ComparisonTargetTypeColumn {
		return nil, nil, fmt.Errorf("%s: unsupported comparison target `%s`", expr.Column.Name, expr.Column.Type)
	}

	switch expr.Column.Name {
	case metadata.TimestampKey:
		err := pr.evalTimestampFromBinaryComparisonOperator(expr, pr.Variables, pr.runtime.UnixTimeUnit)

		return nil, nil, err
	case metadata.MetricValueKey:
		if pr.MetricValue != nil {
			return nil, nil, errMetricValueUnsupportedMultipleComparisons
		}
		pr.MetricValue = expr

		return nil, nil, nil
	case metadata.LogLineKey:
		if pr.LogLine != nil {
			return nil, nil, errStreamValueUnsupportedMultipleComparisons
		}
		pr.LogLine = expr

		return nil, nil, nil
	}

	if _, isLabel := model.Labels[expr.Column.Name]; isLabel {
		return map[string]LabelExpression{
			expr.Column.Name: {
				Name:        expr.Column.Name,
				Expressions: []schema.ExpressionBinaryComparisonOperator{*expr},
			},
		}, nil, nil
	}

	return nil, expr, nil
}
