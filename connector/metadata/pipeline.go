package metadata

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v3"
)

var grafanaArrayStringRegex = regexp.MustCompile(`^{?([\w-,]+)}?$`)

// LogPipelineType represents a log pipeline type of the model
type LogPipelineType string

const (
	PipelineTypeLineFilter  LogPipelineType = "line_filter"
	PipelineTypeLabelFilter LogPipelineType = "label_filter"
	PipelineTypeJSON        LogPipelineType = "json"
	PipelineTypeLogFormat   LogPipelineType = "logfmt"
	PipelineTypePattern     LogPipelineType = "pattern"
	PipelineTypeRegexp      LogPipelineType = "regexp"
	PipelineTypeUnpack      LogPipelineType = "unpack"
	PipelineTypeLineFormat  LogPipelineType = "line_format"
	PipelineTypeLabelFormat LogPipelineType = "label_format"
	PipelineTypeKeep        LogPipelineType = "keep"
	PipelineTypeDrop        LogPipelineType = "drop"
)

var enumValues_LogPipelineType = []LogPipelineType{
	PipelineTypeLineFilter,
	PipelineTypeLabelFilter,
	PipelineTypeJSON,
	PipelineTypeLogFormat,
	PipelineTypePattern,
	PipelineTypeRegexp,
	PipelineTypeUnpack,
	PipelineTypeLineFormat,
	PipelineTypeLabelFormat,
	PipelineTypeKeep,
	PipelineTypeDrop,
}

// ParseLogPipelineType parses the LogPipelineType from a string.
func ParseLogPipelineType(input string) (LogPipelineType, error) {
	result := LogPipelineType(input)
	if !slices.Contains(enumValues_LogPipelineType, result) {
		return "", fmt.Errorf("LogPipelineType must be one of %v, got: %s", enumValues_LogPipelineType, input)
	}

	return result, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *LogPipelineType) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	result, err := ParseLogPipelineType(raw)
	if err != nil {
		return err
	}

	*m = result

	return nil
}

// LogPipeline abstracts the a log pipeline interface.
type LogPipeline interface {
	// GetType returns the pipeline type.
	GetType() LogPipelineType
	// Render return a pipeline expression string
	Render() (string, error)
}

// ModelPipeline represents a log pipeline
type ModelPipeline struct {
	LogPipeline
}

// JSONSchema defines the json schema for the model pipeline
func (m ModelPipeline) JSONSchema() *jsonschema.Schema {
	result := &jsonschema.Schema{
		Description: "LogPipeline abstracts the a log pipeline interface.",
		OneOf:       []*jsonschema.Schema{},
	}

	for _, pipeline := range []LogPipeline{PipelineLineFilter{}, PipelineLabelFilter{}, PipelineJSON{}, PipelineLogFormat{}, PipelinePattern{}, PipelineRegexp{}, PipelineUnpack{}, PipelineLineFormat{}, PipelineLabelFormat{}, PipelineKeep{}, PipelineDrop{}} {
		jc := jsonschema.Reflect(pipeline)
		for _, def := range jc.Definitions {
			result.OneOf = append(result.OneOf, def)
		}
	}

	return result
}

// MarshalYAML implements yaml.Marshaler interface.
func (m ModelPipeline) MarshalYAML() (interface{}, error) {
	return m.LogPipeline, nil
}

// UnmarshalYAML implements yaml.Marshaler interface.
func (m *ModelPipeline) UnmarshalYAML(value *yaml.Node) error {
	var plType string
	for i, node := range value.Content {
		if node.Value != "type" {
			continue
		}
		if len(value.Content) > i+1 {
			plType = value.Content[i+1].Value
		}

		break
	}

	var result LogPipeline
	switch LogPipelineType(plType) {
	case PipelineTypeLineFilter:
		var mp PipelineLineFilter
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeLabelFilter:
		var mp PipelineLabelFilter
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeJSON:
		var mp PipelineJSON
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeLogFormat:
		var mp PipelineLogFormat
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypePattern:
		var mp PipelinePattern
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeRegexp:
		var mp PipelineRegexp
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeUnpack:
		var mp PipelineUnpack
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeLineFormat:
		var mp PipelineLineFormat
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeLabelFormat:
		var mp PipelineLabelFormat
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeKeep:
		var mp PipelineKeep
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	case PipelineTypeDrop:
		var mp PipelineDrop
		if err := value.Decode(&mp); err != nil {
			return err
		}
		result = &mp
	default:
		return fmt.Errorf("invalid pipeline type `%s`", plType)
	}

	m.LogPipeline = result

	return nil
}

// PipelineLineFilter does a distributed grep over the aggregated logs from the matching log streams.
type PipelineLineFilter struct {
	Type     LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=line_filter"`
	Operator string          `json:"operator" yaml:"operator" jsonschema:"enum=_like,enum=_ilike,enum=_nlike,enum=_nilike,enum=_regex,enum=_nregex,enum=_ip,enum=_nip"`
	Value    string          `json:"value" yaml:"value"`
}

var _ LogPipeline = PipelineLineFilter{}

// NewPipelineLineFilter creates a new PipelineLineFilter instance.
func NewPipelineLineFilter(operator string, value string) *PipelineLineFilter {
	return &PipelineLineFilter{
		Type:     PipelineTypeLineFilter,
		Operator: operator,
		Value:    value,
	}
}

// GetType returns the pipeline type.
func (p PipelineLineFilter) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineLineFilter) Render() (string, error) {
	switch p.Operator {
	case ILike, NotILike:
		return fmt.Sprintf("%s `(?i)%s`", lineFilterOperators[p.Operator], p.Value), nil
	case ContainIP, NotContainIP:
		return fmt.Sprintf("%s ip(`%s`)", lineFilterOperators[p.Operator], p.Value), nil
	default:
		op, ok := lineFilterOperators[p.Operator]
		if !ok {
			return "", fmt.Errorf("invalid operator `%s`", p.Operator)
		}

		return fmt.Sprintf("%s `%s`", op, p.Value), nil
	}
}

// PipelineLabelFilter allows filtering log line using their original and extracted labels.
type PipelineLabelFilter struct {
	Type     LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=label_filter"`
	Name     string          `json:"name" yaml:"name"`
	Operator string          `json:"operator" yaml:"operator" jsonschema:"enum=_eq,enum=_neq,enum=_lt,enum=_lte,enum=_gt,enum=_gte,enum=_regex,enum=_nregex,enum=_ip,enum=_nip"`
	Value    any             `json:"value" yaml:"value"`
}

var _ LogPipeline = PipelineLabelFilter{}

// NewPipelineLabelFilter creates a new PipelineLabelFilter instance.
func NewPipelineLabelFilter(name string, operator string, value any) *PipelineLabelFilter {
	return &PipelineLabelFilter{
		Type:     PipelineTypeLabelFilter,
		Name:     name,
		Operator: operator,
		Value:    value,
	}
}

// GetType returns the pipeline type.
func (p PipelineLabelFilter) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineLabelFilter) Render() (string, error) {
	switch p.Operator {
	case In, NotIn:
		values, err := DecodeStringSlice(p.Value)
		if err != nil {
			return "", err
		}
		if values == nil {
			return "", nil
		}
		sb, err := p.createBuilder()
		if err != nil {
			return "", err
		}
		sb.WriteString(" `")
		for i, v := range values {
			if i > 0 {
				sb.WriteRune('|')
			}
			sb.WriteRune('^')
			sb.WriteString(v)
			sb.WriteRune('$')
		}
		sb.WriteRune('`')

		return sb.String(), nil
	case ContainIP, NotContainIP:
		value, err := utils.DecodeNullableString(p.Value)
		if err != nil {
			return "", err
		}
		if value == nil {
			return "", nil
		}

		sb, err := p.createBuilder()
		if err != nil {
			return "", err
		}
		sb.WriteString("ip(`")
		sb.WriteString(*value)
		sb.WriteString("`)")

		return sb.String(), nil
	default:
		if utils.IsNil(p.Value) {
			return "", nil
		}

		sb, err := p.createBuilder()
		if err != nil {
			return "", err
		}
		if str, ok := p.Value.(string); ok {
			sb.WriteRune('`')
			sb.WriteString(str)
			sb.WriteRune('`')
		} else {
			sb.WriteString(fmt.Sprint(str))
		}

		return sb.String(), nil
	}
}

func (p PipelineLabelFilter) createBuilder() (*strings.Builder, error) {
	op, ok := labelFilterOperators[p.Operator]
	if !ok {
		return nil, fmt.Errorf("invalid label filter operator `%s`", p.Operator)
	}

	sb := strings.Builder{}
	sb.WriteString(p.Name)
	sb.WriteRune(' ')
	sb.WriteString(op)
	sb.WriteRune(' ')

	return &sb, nil
}

// PipelineJSON extracts all json properties as labels if the log line is a valid json document.
type PipelineJSON struct {
	Type LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=json"`
	// The collection of selected fields from the pipeline.
	Fields map[string]ModelField `json:"fields" yaml:"fields"`
}

var _ LogPipeline = PipelineJSON{}

// GetType returns the pipeline type.
func (p PipelineJSON) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineJSON) Render() (string, error) {
	return formatParserPipelineFields(string(PipelineTypeJSON), p.Fields)
}

// PipelineLogFormat extracts keys and values from the logfmt formatted log line.
type PipelineLogFormat struct {
	Type LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=logfmt"`
	// The collection of selected fields from the pipeline.
	Fields map[string]ModelField `json:"fields" yaml:"fields"`
}

var _ LogPipeline = PipelineLogFormat{}

// GetType returns the pipeline type.
func (p PipelineLogFormat) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineLogFormat) Render() (string, error) {
	return formatParserPipelineFields(string(PipelineTypeLogFormat), p.Fields)
}

// PipelinePattern allows the explicit extraction of fields from log lines by defining a pattern expression.
type PipelinePattern struct {
	Type    LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=pattern"`
	Pattern string          `json:"pattern" yaml:"pattern"`
	// The collection of selected fields from the pipeline.
	Fields map[string]ModelField `json:"fields" yaml:"fields"`
}

var _ LogPipeline = PipelinePattern{}

// GetType returns the pipeline type.
func (p PipelinePattern) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelinePattern) Render() (string, error) {
	return fmt.Sprintf("| pattern `%s`", p.Pattern), nil
}

// PipelineRegexp takes a regular expression using the Golang RE2 syntax to parse the log line.
type PipelineRegexp struct {
	Type    LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=regexp"`
	Pattern string          `json:"pattern" yaml:"pattern"`
	// The collection of selected fields from the pipeline.
	Fields map[string]ModelField `json:"fields" yaml:"fields"`
}

var _ LogPipeline = PipelineRegexp{}

// GetType returns the pipeline type.
func (p PipelineRegexp) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineRegexp) Render() (string, error) {
	return fmt.Sprintf("| regexp `%s`", p.Pattern), nil
}

// PipelineUnpack the parser parses a JSON log line, unpacking all embedded labels from Promtail’s pack stage.
// A special property _entry will also be used to replace the original log line.
type PipelineUnpack struct {
	Type LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=unpack"`
	// The collection of selected fields from the pipeline.
	Fields map[string]ModelField `json:"fields" yaml:"fields"`
}

var _ LogPipeline = PipelineUnpack{}

// GetType returns the pipeline type.
func (p PipelineUnpack) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineUnpack) Render() (string, error) {
	return "| unpack", nil
}

// PipelineLineFormat can rewrite the log line content by using the text/template format.
type PipelineLineFormat struct {
	Type     LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=line_format"`
	Template string          `json:"template" yaml:"template"`
}

var _ LogPipeline = PipelineLineFormat{}

// GetType returns the pipeline type.
func (p PipelineLineFormat) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineLineFormat) Render() (string, error) {
	return fmt.Sprintf("| line_format `%s`", p.Template), nil
}

// LabelFormatOperation the format operation of a label
type LabelFormatRule struct {
	Source   string `json:"source" yaml:"source" jsonschema:"oneof_required=source_label"`
	Template string `json:"template" yaml:"template" jsonschema:"oneof_required=template"`
}

// PipelineLabelFormat can rename, modify or add labels.
type PipelineLabelFormat struct {
	Type   LogPipelineType            `json:"type" yaml:"type" jsonschema:"enum=label_format"`
	Labels map[string]LabelFormatRule `json:"labels" yaml:"labels"`
}

var _ LogPipeline = PipelineLabelFormat{}

// GetType returns the pipeline type.
func (p PipelineLabelFormat) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineLabelFormat) Render() (string, error) {
	var sb strings.Builder
	sb.WriteString("| label_format ")
	for key, label := range p.Labels {
		sb.WriteString(key)
		sb.WriteRune('=')
		if label.Source != "" {
			sb.WriteString(label.Source)
		} else {
			sb.WriteRune('`')
			sb.WriteString(label.Template)
			sb.WriteRune('`')
		}
		sb.WriteRune(' ')
	}

	return sb.String(), nil
}

// PipelineKeep will keep only the specified labels in the pipeline and drop all the other labels.
type PipelineKeep struct {
	Type LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=keep"`
	// The collection of selected fields from the pipeline.
	Fields map[string]ModelField `json:"fields" yaml:"fields"`
}

var _ LogPipeline = PipelineKeep{}

// GetType returns the pipeline type.
func (p PipelineKeep) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineKeep) Render() (string, error) {
	return formatParserPipelineFields(string(PipelineTypeKeep), p.Fields)
}

// PipelineDrop will drop the given labels in the pipeline.
type PipelineDrop struct {
	Type LogPipelineType `json:"type" yaml:"type" jsonschema:"enum=drop"`
	// The collection of selected fields from the pipeline.
	Fields map[string]ModelField `json:"fields" yaml:"fields"`
}

var _ LogPipeline = PipelineDrop{}

// GetType returns the pipeline type.
func (p PipelineDrop) GetType() LogPipelineType {
	return p.Type
}

// String implement the fmt.Stringer interface.
func (p PipelineDrop) Render() (string, error) {
	return formatParserPipelineFields(string(PipelineTypeDrop), p.Fields)
}

// ModelField the metadata of a field.
type ModelField struct {
	// Description of the field.
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// The path to select the value from JSON or logfmt log body.
	Selector *string `json:"selector,omitempty" yaml:"selector,omitempty"`
}

func formatParserPipelineFields(operator string, fields map[string]ModelField) (string, error) {
	var sb strings.Builder
	sb.WriteString("| ")
	sb.WriteString(operator)
	sb.WriteRune(' ')

	for i, key := range utils.GetSortedKeys(fields) {
		if i > 0 {
			sb.WriteString(", ")
		}
		field := fields[key]
		sb.WriteString(key)
		if field.Selector != nil && *field.Selector != "" {
			sb.WriteString("=`")
			sb.WriteString(*field.Selector)
			sb.WriteRune('`')
		}
	}

	return sb.String(), nil
}

// DecodeStringSlice decodes string slice from an unknown argument value
func DecodeStringSlice(value any) ([]string, error) {
	if utils.IsNil(value) {
		return nil, nil
	}
	var err error
	sliceValue := []string{}
	if str, ok := value.(string); ok {
		if str == "" {
			return nil, nil
		}
		matches := grafanaArrayStringRegex.FindStringSubmatch(str)
		if len(matches) > 1 {
			sliceValue = strings.Split(matches[1], ",")
			for i, str := range sliceValue {
				sliceValue[i] = strings.TrimSpace(str)
			}

			return sliceValue, nil
		}

		// try to parse the slice from the json string
		err = json.Unmarshal([]byte(str), &sliceValue)
	} else {
		sliceValue, err = utils.DecodeStringSlice(value)
	}
	if err != nil {
		return nil, err
	}

	return sliceValue, nil
}
