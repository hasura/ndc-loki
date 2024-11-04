package internal

import (
	"context"
	"strings"
	"testing"

	"github.com/hasura/ndc-loki/connector/metadata"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"gotest.tools/v3/assert"
)

func TestCollectionExplain(t *testing.T) {
	testCases := []struct {
		Name      string
		Request   schema.QueryRequest
		Model     metadata.ModelInfo
		Query     string
		QueryType metadata.QueryType
		ErrorMsg  string
	}{
		{
			Name:      "ip",
			Query:     "{job_name=\"myapp\"} != ip(`192.168.4.5-192.168.4.20`)",
			QueryType: metadata.QueryTypeStream,
			Model: metadata.ModelInfo{
				Labels: map[string]metadata.ModelLabelInfo{
					"job_name": {},
				},
			},
			Request: schema.QueryRequest{
				Collection:              "myapp",
				Arguments:               schema.QueryRequestArguments{},
				CollectionRelationships: schema.QueryRequestCollectionRelationships{},
				Query: schema.Query{
					Fields: schema.QueryFields{},
					Predicate: schema.NewExpressionAnd(
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("job_name", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("myapp"),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("log_line", nil, nil),
							"_nip",
							schema.NewComparisonValueScalar("192.168.4.5-192.168.4.20"),
						),
					).Encode(),
				},
			},
		},
		{
			Name:      "logfmt",
			Query:     "{job_name=\"myapp\"} | logfmt addr=`addr` | (addr = ip(`192.168.4.5/16`) and addr != ip(`192.168.4.2`))",
			QueryType: metadata.QueryTypeStream,
			Model: metadata.ModelInfo{
				Labels: map[string]metadata.ModelLabelInfo{
					"job_name": {},
				},
				Pipelines: []metadata.ModelPipeline{
					{
						LogPipeline: metadata.PipelineLogFormat{
							Type: metadata.PipelineTypeLogFormat,
							Fields: map[string]metadata.ModelField{
								"addr": {
									Selector: utils.ToPtr("addr"),
								},
							},
						},
					},
				},
			},
			Request: schema.QueryRequest{
				Collection:              "myapp",
				Arguments:               schema.QueryRequestArguments{},
				CollectionRelationships: schema.QueryRequestCollectionRelationships{},
				Query: schema.Query{
					Fields: schema.QueryFields{},
					Predicate: schema.NewExpressionAnd(
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("job_name", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("myapp"),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("addr", nil, nil),
							"_ip",
							schema.NewComparisonValueScalar("192.168.4.5/16"),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("addr", nil, nil),
							"_nip",
							schema.NewComparisonValueScalar("192.168.4.2"),
						),
					).Encode(),
				},
			},
		},
		{
			Name:      "regexp",
			Query:     "{job=\"security\"} |~ `Invalid user.*` | regexp `(^(?P<user>\\S+ {1,2}){8})` | regexp `(^(?P<ip>\\S+ {1,2}){10})` | line_format `IP = {{.ip}}\\tUSER = {{.user}}`",
			QueryType: metadata.QueryTypeStream,
			Model: metadata.ModelInfo{
				Labels: map[string]metadata.ModelLabelInfo{
					"job": {},
				},
				Pipelines: []metadata.ModelPipeline{
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_regex",
							Value:    `Invalid user.*`,
						},
					},
					{
						LogPipeline: metadata.PipelineRegexp{
							Type:    metadata.PipelineTypeRegexp,
							Pattern: `(^(?P<user>\S+ {1,2}){8})`,
							Fields: map[string]metadata.ModelField{
								"user": {},
							},
						},
					},
					{
						LogPipeline: metadata.PipelineRegexp{
							Type:    metadata.PipelineTypeRegexp,
							Pattern: `(^(?P<ip>\S+ {1,2}){10})`,
							Fields: map[string]metadata.ModelField{
								"ip": {},
							},
						},
					},
					{
						LogPipeline: metadata.PipelineLineFormat{
							Type:     metadata.PipelineTypeLineFormat,
							Template: `IP = {{.ip}}\tUSER = {{.user}}`,
						},
					},
				},
			},
			Request: schema.QueryRequest{
				Collection:              "myapp",
				Arguments:               schema.QueryRequestArguments{},
				CollectionRelationships: schema.QueryRequestCollectionRelationships{},
				Query: schema.Query{
					Fields: schema.QueryFields{},
					Predicate: schema.NewExpressionAnd(
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("job", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("security"),
						),
					).Encode(),
				},
			},
		},
		{
			Name:      "regexp2",
			Query:     "{job=\"security\"} != `grafana_com` |= `session opened` != `sudo: ` | regexp `(^(?P<user>\\S+ {1,2}){11})` | line_format `USER = {{.user}}`",
			QueryType: metadata.QueryTypeStream,
			Model: metadata.ModelInfo{
				Labels: map[string]metadata.ModelLabelInfo{
					"job": {},
				},
				Pipelines: []metadata.ModelPipeline{
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_nlike",
							Value:    `grafana_com`,
						},
					},
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_like",
							Value:    `session opened`,
						},
					},
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_nlike",
							Value:    `sudo: `,
						},
					},
					{
						LogPipeline: metadata.PipelineRegexp{
							Type:    metadata.PipelineTypeRegexp,
							Pattern: `(^(?P<user>\S+ {1,2}){11})`,
							Fields: map[string]metadata.ModelField{
								"user": {},
							},
						},
					},
					{
						LogPipeline: metadata.PipelineLineFormat{
							Type:     metadata.PipelineTypeLineFormat,
							Template: `USER = {{.user}}`,
						},
					},
				},
			},
			Request: schema.QueryRequest{
				Collection:              "myapp",
				Arguments:               schema.QueryRequestArguments{},
				CollectionRelationships: schema.QueryRequestCollectionRelationships{},
				Query: schema.Query{
					Fields: schema.QueryFields{},
					Predicate: schema.NewExpressionAnd(
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("job", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("security"),
						),
					).Encode(),
				},
			},
		},
		{
			Name:      "sum",
			Query:     "sum by (host) (rate({job=\"mysql\"} |= `error` != `timeout` | json duration, host | (duration = `10s`) [1m0s]))",
			QueryType: metadata.QueryTypeMetric,
			Model: metadata.ModelInfo{
				Labels: map[string]metadata.ModelLabelInfo{
					"job": {},
				},
				Pipelines: []metadata.ModelPipeline{
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_like",
							Value:    `error`,
						},
					},
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_nlike",
							Value:    `timeout`,
						},
					},
					{
						LogPipeline: metadata.PipelineJSON{
							Type: metadata.PipelineTypeJSON,
							Fields: map[string]metadata.ModelField{
								"duration": {},
								"host":     {},
							},
						},
					},
				},
			},
			Request: schema.QueryRequest{
				Collection: "myapp",
				Arguments: schema.QueryRequestArguments{
					metadata.ArgumentKeyAggregations: schema.NewArgumentLiteral([]map[string]any{
						{"rate": "1m"},
						{
							"sum": map[string]any{
								"by": []string{"host"},
							},
						},
					}).Encode(),
				},
				CollectionRelationships: schema.QueryRequestCollectionRelationships{},
				Query: schema.Query{
					Fields: schema.QueryFields{},
					Predicate: schema.NewExpressionAnd(
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("job", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("mysql"),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("duration", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("10s"),
						),
					).Encode(),
				},
			},
		},
		{
			Name:      "multiple_parsers",
			Query:     "{job=\"loki-ops/query-frontend\"} | logfmt  | line_format `{{.msg}}` | regexp `(?P<method>\\w+) (?P<path>[\\w|/]+) \\((?P<status>\\d+?)\\) (?P<duration>.*)`",
			QueryType: metadata.QueryTypeStream,
			Model: metadata.ModelInfo{
				Labels: map[string]metadata.ModelLabelInfo{
					"job": {
						Filter: &metadata.LabelFilterSetting{
							Operator: "_eq",
							Value:    "loki-ops/query-frontend",
							Static:   utils.ToPtr(true),
						},
					},
				},
				Pipelines: []metadata.ModelPipeline{
					{
						LogPipeline: metadata.PipelineLogFormat{
							Type:   metadata.PipelineTypeLogFormat,
							Fields: map[string]metadata.ModelField{},
						},
					},
					{
						LogPipeline: metadata.PipelineLineFormat{
							Type:     metadata.PipelineTypeLineFormat,
							Template: `{{.msg}}`,
						},
					},
					{
						LogPipeline: metadata.PipelineRegexp{
							Type:    metadata.PipelineTypeRegexp,
							Pattern: `(?P<method>\w+) (?P<path>[\w|/]+) \((?P<status>\d+?)\) (?P<duration>.*)`,
							Fields:  map[string]metadata.ModelField{},
						},
					},
				},
			},
			Request: schema.QueryRequest{
				Collection:              "myapp",
				Arguments:               schema.QueryRequestArguments{},
				CollectionRelationships: schema.QueryRequestCollectionRelationships{},
				Query: schema.Query{
					Fields:    schema.QueryFields{},
					Predicate: schema.NewExpressionAnd().Encode(),
				},
			},
		},
		{
			Name:      "label_format",
			Query:     "{cluster=\"ops-tools1\", name=\"querier\", namespace=\"loki-dev\"} | decolorize |= `metrics.go` != `loki-canary` | logfmt  | query != `` | label_format query=`{{ Replace .query \"\\n\" \"\" -1 }}`  | line_format `{{ .ts}} {{.duration}} traceID = {{.traceID}} {{ printf \"%-100.100s\" .query }}`",
			QueryType: metadata.QueryTypeStream,
			Model: metadata.ModelInfo{
				Decolorize: utils.ToPtr(true),
				Labels: map[string]metadata.ModelLabelInfo{
					"cluster":   {},
					"name":      {},
					"namespace": {},
				},
				Pipelines: []metadata.ModelPipeline{
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_like",
							Value:    "metrics.go",
						},
					},
					{
						LogPipeline: metadata.PipelineLineFilter{
							Type:     metadata.PipelineTypeLineFilter,
							Operator: "_nlike",
							Value:    "loki-canary",
						},
					},
					{
						LogPipeline: metadata.PipelineLogFormat{
							Type:   metadata.PipelineTypeLogFormat,
							Fields: map[string]metadata.ModelField{},
						},
					},
					{
						LogPipeline: metadata.PipelineLabelFilter{
							Type:     metadata.PipelineTypeLabelFilter,
							Name:     "query",
							Operator: "_neq",
							Value:    "",
						},
					},
					{
						LogPipeline: metadata.PipelineLabelFormat{
							Type: metadata.PipelineTypeLabelFormat,
							Labels: map[string]metadata.LabelFormatRule{
								"query": {
									Template: `{{ Replace .query "\n" "" -1 }}`,
								},
							},
						},
					},
					{
						LogPipeline: metadata.PipelineLineFormat{
							Type:     metadata.PipelineTypeLineFormat,
							Template: `{{ .ts}} {{.duration}} traceID = {{.traceID}} {{ printf "%-100.100s" .query }}`,
						},
					},
				},
			},
			Request: schema.QueryRequest{
				Collection:              "myapp",
				Arguments:               schema.QueryRequestArguments{},
				CollectionRelationships: schema.QueryRequestCollectionRelationships{},
				Query: schema.Query{
					Fields: schema.QueryFields{},
					Predicate: schema.NewExpressionAnd(
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("cluster", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("ops-tools1"),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("name", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("querier"),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("namespace", nil, nil),
							"_eq",
							schema.NewComparisonValueScalar("loki-dev"),
						),
					).Encode(),
				},
			},
		},
		{
			Name:      "array_in",
			Query:     "{container_name=\"ndc-loki-graphql-engine-1\", service_name=\"ndc-loki-graphql-engine-1\"} | decolorize | json level, type | (level =~  `^error$|^info$` and level =~  `^http-log$`)",
			QueryType: metadata.QueryTypeStream,
			Model: metadata.ModelInfo{
				Decolorize: utils.ToPtr(true),
				Labels: map[string]metadata.ModelLabelInfo{
					"container_name": {
						Filter: &metadata.LabelFilterSetting{
							Operator: "_eq",
							Value:    "ndc-loki-graphql-engine-1",
						},
					},
					"service_name": {
						Filter: &metadata.LabelFilterSetting{
							Operator: "_eq",
							Value:    "ndc-loki-graphql-engine-1",
							Static:   utils.ToPtr(true),
						},
					},
				},
				Pipelines: []metadata.ModelPipeline{
					{
						LogPipeline: metadata.PipelineJSON{
							Type: metadata.PipelineTypeJSON,
							Fields: map[string]metadata.ModelField{
								"type":  {},
								"level": {},
							},
						},
					},
				},
			},
			Request: schema.QueryRequest{
				Collection: "hasura_log",
				Query: schema.Query{
					Fields: schema.QueryFields{},
					Predicate: schema.NewExpressionAnd(
						schema.NewExpressionAnd(
							schema.NewExpressionBinaryComparisonOperator(
								*schema.NewComparisonTargetColumn("timestamp", nil, nil),
								"_gt",
								schema.NewComparisonValueScalar(1730697040059),
							),
							schema.NewExpressionBinaryComparisonOperator(
								*schema.NewComparisonTargetColumn("timestamp", nil, nil),
								"_lt",
								schema.NewComparisonValueScalar(1730718640059),
							),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("level", nil, nil),
							"_in",
							schema.NewComparisonValueScalar("{error,info}"),
						),
						schema.NewExpressionBinaryComparisonOperator(
							*schema.NewComparisonTargetColumn("level", nil, nil),
							"_in",
							schema.NewComparisonValueScalar("http-log"),
						),
					).Encode(),
				},
				Arguments: schema.QueryRequestArguments{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			arguments, err := utils.ResolveArgumentVariables(tc.Request.Arguments, nil)
			assert.NilError(t, err)

			executor := QueryCollectionExecutor{
				Request:   &tc.Request,
				Model:     tc.Model,
				QueryType: tc.QueryType,
				Arguments: arguments,
				Runtime:   &metadata.RuntimeSettings{},
			}
			_, query, _, err := executor.Explain(context.TODO())
			if tc.ErrorMsg != "" {
				assert.ErrorContains(t, err, tc.ErrorMsg)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, tc.Query, strings.TrimSpace(query))
			}
		})
	}
}
