package internal

import (
	"fmt"
	"strings"

	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

func (qce *QueryCollectionExecutor) buildLogQuery(predicate *CollectionRequest) (string, bool, error) {
	if predicate == nil {
		return fmt.Sprintf("{%s}", qce.buildDefaultLabelFilter()), true, nil
	}

	var sb strings.Builder
	if err := qce.evalLabelExpressions(&sb, predicate); err != nil {
		return "", false, err
	}
	if qce.Model.Decolorize != nil && *qce.Model.Decolorize {
		sb.WriteString("| decolorize ")
	}
	for _, pipeline := range qce.Model.Pipelines {
		str, err := pipeline.Render()
		if err != nil {
			return "", false, err
		}
		if pipeline.GetType() == metadata.PipelineTypeLabelFilter {
			sb.WriteString("| ")
		}
		sb.WriteString(str)
		sb.WriteRune(' ')
	}

	if predicate.LogLine != nil {
		lineFilter, err := qce.evalLineFilter(predicate.LogLine)
		if err != nil {
			return "", false, err
		}
		if lineFilter != "" {
			sb.WriteString(lineFilter)
		}
	}

	if len(qce.Model.Pipelines) > 0 {
		fieldCondition, err := qce.evalFieldExpression(predicate.FieldExpression)
		if err != nil {
			return "", false, err
		}

		if fieldCondition != "" {
			sb.WriteString("| ")
			sb.WriteString(fieldCondition)
		}
	}

	return sb.String(), true, nil
}

func (qce *QueryCollectionExecutor) evalLabelExpressions(sb *strings.Builder, predicate *CollectionRequest) error {
	labelExpressions := mergeLabelExpressions(nil, predicate.LabelExpressions)
	for key, label := range qce.Model.Labels {
		if label.Filter == nil {
			continue
		}

		_, ok := labelExpressions[key]
		if !ok || (label.Filter.Static != nil && *label.Filter.Static) {
			labelExpressions[key] = LabelExpression{
				Name: key,
				Expressions: []schema.ExpressionBinaryComparisonOperator{
					*schema.NewExpressionBinaryComparisonOperator(*schema.NewComparisonTargetColumn(key, nil, nil),
						label.Filter.Operator,
						schema.NewComparisonValueScalar(label.Filter.Value),
					),
				},
			}
		}
	}

	sb.WriteRune('{')
	if len(labelExpressions) > 0 {
		keys := utils.GetSortedKeys(labelExpressions)
		for i, key := range keys {
			expr := labelExpressions[key]
			condition, ok, err := (&LabelExpressionBuilder{
				Labels:          qce.Model.Labels,
				LabelExpression: expr,
			}).Evaluate(qce.Variables)
			if err != nil || !ok {
				return err
			}
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(condition)
		}
	} else {
		sb.WriteString(qce.buildDefaultLabelFilter())
	}
	sb.WriteString("} ")

	return nil
}

func (qce *QueryCollectionExecutor) evalLineFilter(expr *schema.ExpressionBinaryComparisonOperator) (string, error) {
	value, err := getComparisonValueString(expr.Value, qce.Variables)
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", nil
	}

	pipeline := metadata.NewPipelineLineFilter(expr.Operator, *value)

	return pipeline.Render()
}

func (qce *QueryCollectionExecutor) evalFieldExpression(expression schema.ExpressionEncoder) (string, error) {
	if expression == nil {
		return "", nil
	}

	switch expr := expression.(type) {
	case *schema.ExpressionAnd:
		if len(expr.Expressions) == 0 {
			return "", nil
		}
		conditions := []string{}
		for _, e := range expr.Expressions {
			cond, err := qce.evalFieldExpression(e.Interface())
			if err != nil {
				return "", err
			}
			if cond != "" {
				conditions = append(conditions, cond)
			}
		}

		return "(" + strings.Join(conditions, " and ") + ")", nil
	case *schema.ExpressionOr:
		if len(expr.Expressions) == 0 {
			return "", nil
		}
		conditions := []string{}
		for _, e := range expr.Expressions {
			cond, err := qce.evalFieldExpression(e.Interface())
			if err != nil {
				return "", err
			}
			if cond != "" {
				conditions = append(conditions, cond)
			}
		}

		return "(" + strings.Join(conditions, " or ") + ")", nil
	case *schema.ExpressionBinaryComparisonOperator:
		return qce.evalFieldExpressionBinaryComparisonOperator(expr)
	default:
		return "", fmt.Errorf("unsupported expression: %+v", expression)
	}
}

// evaluate the expression binary comparison operator of a field
func (qce *QueryCollectionExecutor) evalFieldExpressionBinaryComparisonOperator(expr *schema.ExpressionBinaryComparisonOperator) (string, error) {
	value, err := getComparisonValue(expr.Value, qce.Variables)
	if err != nil {
		return "", err
	}

	pipeline := metadata.NewPipelineLabelFilter(expr.Column.Name, expr.Operator, value)

	return pipeline.Render()
}
