package internal

import (
	"context"
	"fmt"

	"github.com/hasura/ndc-loki/connector/client"
	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"go.opentelemetry.io/otel/trace"
)

// QueryCollectionExecutor evaluates and executes the query request.
type QueryCollectionExecutor struct {
	Client    *client.Client
	Tracer    trace.Tracer
	Runtime   *metadata.RuntimeSettings
	Request   *schema.QueryRequest
	Model     metadata.ModelInfo
	QueryType metadata.QueryType
	Variables map[string]any
	Arguments map[string]any
}

// Explain explains the query request
func (qce *QueryCollectionExecutor) Explain(ctx context.Context) (*CollectionRequest, string, bool, error) {
	expressions, err := EvalCollectionRequest(qce.Request, qce.Model, qce.Arguments, qce.Variables, qce.Runtime)
	if err != nil {
		return nil, "", false, schema.UnprocessableContentError(err.Error(), map[string]any{
			"collection": qce.Request.Collection,
		})
	}
	if qce.QueryType == metadata.QueryTypeMetric && len(expressions.Aggregations) == 0 {
		return nil, "", false, schema.UnprocessableContentError(errQueryAggregationsRequired.Error(), map[string]any{
			"collection": qce.Request.Collection,
		})
	}

	queryString, ok, err := qce.buildQueryString(expressions)
	if err != nil {
		return nil, "", false, schema.UnprocessableContentError(fmt.Sprintf("failed to evaluate the query: %s", err.Error()), map[string]any{
			"collection": qce.Request.Collection,
		})
	}

	return expressions, queryString, ok, nil
}

// Execute executes the query request
func (qce *QueryCollectionExecutor) Execute(ctx context.Context) (*schema.RowSet, error) {
	ctx, span := qce.Tracer.Start(ctx, "Execute Collection")
	defer span.End()

	expressions, queryString, ok, err := qce.Explain(ctx)
	if err != nil {
		return nil, err
	}

	if !ok {
		// early returns zero rows
		// the evaluated query always returns empty values
		return &schema.RowSet{
			Aggregates: schema.RowSetAggregates{},
			Rows:       []map[string]any{},
		}, nil
	}

	if expressions.Flat == nil {
		expressions.Flat = &qce.Runtime.Flat
	}

	var rawResults []map[string]any
	if expressions.Timestamp != nil {
		rawResults, err = qce.queryInstant(ctx, queryString, expressions)
	} else {
		rawResults, err = qce.queryRange(ctx, queryString, expressions)
	}

	if err != nil {
		return nil, err
	}

	result, err := utils.EvalObjectsWithColumnSelection(qce.Request.Query.Fields, rawResults)
	if err != nil {
		return nil, err
	}

	return &schema.RowSet{
		Aggregates: schema.RowSetAggregates{},
		Rows:       result,
	}, nil
}

func (qce *QueryCollectionExecutor) queryInstant(ctx context.Context, queryString string, predicate *CollectionRequest) ([]map[string]any, error) {
	params := &client.QueryParams{
		Query:     queryString,
		Time:      predicate.Timestamp,
		Direction: predicate.GetQueryDirection(),
	}
	if qce.Request.Query.Limit != nil {
		params.Limit = *qce.Request.Query.Limit
	}
	resp, err := qce.Client.Query(ctx, params)
	if err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	vector := resp.Vector
	sortVector(vector, predicate.OrderBy)
	vector = paginateVector(vector, qce.Request.Query)
	results := serializeMetricVector(vector, getLabelInfosFromModelLabels(qce.Model.Labels), qce.Model.GetFields(), qce.Runtime, predicate.GetFlat())

	return results, nil
}

func (qce *QueryCollectionExecutor) queryRange(ctx context.Context, queryString string, predicate *CollectionRequest) ([]map[string]any, error) {
	params := &client.QueryRangeParams{
		Query:     queryString,
		Start:     predicate.Start,
		End:       predicate.End,
		Since:     predicate.Since,
		Interval:  predicate.Interval,
		Step:      predicate.Step,
		Direction: predicate.GetQueryDirection(),
	}

	if qce.Request.Query.Limit != nil {
		params.Limit = *qce.Request.Query.Limit
	}

	resp, err := qce.Client.QueryRange(ctx, params)
	if err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	labels := getLabelInfosFromModelLabels(qce.Model.Labels)
	if resp.ResultType == client.ResultTypeMatrix {
		matrix := resp.Matrix
		sortMatrix(matrix, predicate.OrderBy)
		results := serializeMetricMatrix(matrix, labels, qce.Model.GetFields(), qce.Runtime, predicate.GetFlat())

		return paginateQueryResults(results, qce.Request.Query), nil
	}

	results := serializeStreamValues(resp.Stream, labels, qce.Model.GetFields(), qce.Runtime, predicate.GetFlat())

	return results, nil
}

func (qce *QueryCollectionExecutor) buildQueryString(predicate *CollectionRequest) (string, bool, error) {
	logQuery, ok, err := qce.buildLogQuery(predicate)
	if err != nil {
		return "", false, err
	}
	if qce.QueryType != metadata.QueryTypeMetric {
		return logQuery, ok, nil
	}

	query, err := qce.buildMetricQuery(logQuery, predicate)
	if err != nil {
		return "", false, err
	}

	return query, ok, nil
}

func (qce *QueryCollectionExecutor) evalValueComparisonCondition(operator *schema.ExpressionBinaryComparisonOperator) (string, error) {
	if operator == nil {
		return "", nil
	}
	v, err := getComparisonValueFloat64(operator.Value, qce.Variables)
	if err != nil {
		return "", fmt.Errorf("invalid value expression: %s", err)
	}
	if v == nil {
		return "", nil
	}

	op, ok := valueBinaryOperators[operator.Operator]
	if !ok {
		return "", fmt.Errorf("value: unsupported comparison operator `%s`", operator)
	}

	return fmt.Sprintf(" %s %f", op, *v), nil
}

func (qce *QueryCollectionExecutor) buildDefaultLabelFilter() string {
	return qce.Model.GetFirstLabelName() + `!=""`
}
