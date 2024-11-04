package metadata

import (
	"testing"

	"github.com/hasura/ndc-sdk-go/utils"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

func TestUnmarshalModelPipeline(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          string
		Expected       LogPipeline
		ExpectedString string
	}{
		{
			Name: "line_filter",
			Input: `
type: line_filter
operator: _ilike
value: '"type":"http-log"'`,
			Expected: &PipelineLineFilter{
				Type:     PipelineTypeLineFilter,
				Operator: "_ilike",
				Value:    `"type":"http-log"`,
			},
			ExpectedString: "|~ `(?i)\"type\":\"http-log\"`",
		},
		{
			Name: "json",
			Input: `
type: json
fields:
  type:
    selector: type
  level:
    selector: level
  http_status:
    selector: detail.http_info.status`,
			Expected: &PipelineJSON{
				Type: PipelineTypeJSON,
				Fields: map[string]ModelField{
					"type": {
						Selector: utils.ToPtr("type"),
					},
					"level": {
						Selector: utils.ToPtr("level"),
					},
					"http_status": {
						Selector: utils.ToPtr("detail.http_info.status"),
					},
				},
			},
			ExpectedString: "| json http_status=`detail.http_info.status`, level=`level`, type=`type`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var result ModelPipeline
			assert.NilError(t, yaml.Unmarshal([]byte(tc.Input), &result))
			assert.DeepEqual(t, tc.Expected, result.LogPipeline)
			str, err := result.Render()
			assert.NilError(t, err)
			assert.Equal(t, tc.ExpectedString, str)
		})
	}
}

func TestDecodeStringSlice(t *testing.T) {
	testCases := []struct {
		Input    string
		Expected []string
	}{
		{
			Input:    "info",
			Expected: []string{"info"},
		},
		{
			Input:    "{info,error}",
			Expected: []string{"info", "error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Input, func(t *testing.T) {
			result, err := DecodeStringSlice(tc.Input)
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.Expected, result)
		})
	}
}
