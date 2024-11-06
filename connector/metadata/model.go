package metadata

import (
	"encoding/json"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/iancoleman/strcase"
)

// ModelFieldValue wraps the ModelField with value.
type ModelFieldValue struct {
	ModelField

	Value *string
}

// ModelInfo represents the metadata of a log model.
type ModelInfo struct {
	// The list of preprocessing pipelines
	Pipelines []ModelPipeline `json:"pipelines" yaml:"pipelines"`
	// The pattern to be used if the parser type is pattern or regexp
	Pattern string `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	// Description of the model.
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// Labels returned by the model.
	Labels map[string]ModelLabelInfo `json:"labels" yaml:"labels"`
	// Stripping ANSI sequences (color codes) from the line.
	Decolorize *bool `json:"decolorize,omitempty" yaml:"decolorize,omitempty"`

	firstLabelName string
	fields         map[string]ModelFieldValue
}

type rawModelInfo ModelInfo

// UnmarshalJSON implements json.Unmarshaler.
func (m *ModelInfo) UnmarshalJSON(b []byte) error {
	var raw rawModelInfo
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	if len(raw.Labels) == 0 {
		return errModelInfoLabelsRequired
	}
	*m = ModelInfo(raw)

	return nil
}

// GetFirstLabelName get the first label name in alphabet order
func (m *ModelInfo) GetFirstLabelName() string {
	if m.firstLabelName != "" {
		return m.firstLabelName
	}
	if len(m.Labels) == 0 {
		return ""
	}

	labelNames := utils.GetSortedKeys(m.Labels)
	label := m.Labels[labelNames[0]]
	if label.Source != "" {
		m.firstLabelName = label.Source
	} else {
		m.firstLabelName = labelNames[0]
	}

	return m.firstLabelName
}

// GetFields get the final model field map
func (m *ModelInfo) GetFields() map[string]ModelFieldValue {
	if m.fields != nil {
		return m.fields
	}

	m.fields = make(map[string]ModelFieldValue)
	for _, pipeline := range m.Pipelines {
		if pipeline.LogPipeline == nil {
			continue
		}
		switch p := pipeline.LogPipeline.(type) {
		case *PipelineJSON:
			m.setFields(p.Fields)
		case *PipelineLogFormat:
			m.setFields(p.Fields)
		case *PipelinePattern:
			m.setFields(p.Fields)
		case *PipelineRegexp:
			m.setFields(p.Fields)
		case *PipelineUnpack:
			m.setFields(p.Fields)
		case *PipelineLabelFormat:
			for key, label := range p.Labels {
				if label.Template != "" {
					m.fields[key] = ModelFieldValue{
						Value: &label.Template,
					}

					continue
				}

				if v, ok := m.fields[label.Source]; ok {
					m.fields[key] = v
					delete(m.fields, label.Source)
				}
			}
		case *PipelineDrop:
			for key, label := range p.Fields {
				if label.Selector != nil && *label.Selector != "" {
					continue
				}
				delete(m.fields, key)
			}
		case *PipelineKeep:
			m.fields = make(map[string]ModelFieldValue)
			for key, label := range p.Fields {
				m.fields[key] = ModelFieldValue{
					ModelField: label,
				}
			}
		}
	}

	return m.fields
}

func (m *ModelInfo) setFields(fields map[string]ModelField) {
	for key, field := range fields {
		m.fields[key] = ModelFieldValue{
			ModelField: field,
		}
	}
}

// LabelInfo the information of a log label
type LabelInfo struct {
	// The source label name.
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// Description of the label.
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
}

// LabelFilterSetting represents the filter setting for the label
type LabelFilterSetting struct {
	// The filter operator to be used.
	Operator string `json:"operator" yaml:"operator" jsonschema:"enum=_eq,enum=_neq,enum=_regex,enum=_nregex,default=_eq"`
	// The filter value.
	Value string `json:"value" yaml:"value"`
	// The static filter can't be changed.
	Static *bool `json:"static,omitempty" yaml:"static,omitempty"`
}

// ModelLabelInfo represents the label information of the model
type ModelLabelInfo struct {
	LabelInfo `yaml:",inline"`

	// Default filter setting for the label
	Filter *LabelFilterSetting `json:"filter,omitempty"`
}

func (scb *connectorSchemaBuilder) buildModels() error {
	for name, info := range scb.Metadata.Models {
		if err := scb.buildStreamItem(name, info); err != nil {
			return err
		}

		if err := scb.buildMetricItem(name, info); err != nil {
			return err
		}
	}

	return nil
}

func (scb *connectorSchemaBuilder) buildStreamItem(name string, info ModelInfo) error {
	objectName, _, err := scb.createModelObjectType(info, name, QueryTypeStream)
	if err != nil {
		return err
	}
	arguments := createCollectionArguments(QueryTypeStream)
	collection := schema.CollectionInfo{
		Name:                  name,
		Type:                  objectName,
		Arguments:             arguments,
		Description:           info.Description,
		ForeignKeys:           schema.CollectionInfoForeignKeys{},
		UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
	}

	scb.Collections[name] = collection

	return nil
}

func (scb *connectorSchemaBuilder) createModelObjectType(info ModelInfo, name string, queryType QueryType) (string, []string, error) {
	if err := scb.checkDuplicatedOperation(name); err != nil {
		return "", nil, err
	}

	objectType := schema.ObjectType{
		Fields: createQueryResultValuesObjectFields(queryType),
	}

	labels := map[string]bool{}
	for key, label := range info.Labels {
		if label.Filter != nil && label.Filter.Static != nil && *label.Filter.Static {
			continue
		}
		labels[key] = true
		objectType.Fields[key] = schema.ObjectField{
			Description: label.Description,
			Type:        schema.NewNamedType(string(ScalarLabel)).Encode(),
		}
	}

	for key, field := range info.GetFields() {
		labels[key] = true
		objectType.Fields[key] = schema.ObjectField{
			Description: field.Description,
			Type:        schema.NewNamedType(string(ScalarString)).Encode(),
		}
	}

	objectName := strcase.ToCamel(name)
	scb.ObjectTypes[objectName] = objectType

	return objectName, utils.GetSortedKeys(labels), nil
}

func (scb *connectorSchemaBuilder) buildMetricItem(name string, info ModelInfo) error {
	name += "_aggregate"

	objectName, labelEnums, err := scb.createModelObjectType(info, name, QueryTypeMetric)
	if err != nil {
		return err
	}
	labelEnumScalarName := objectName + "Label"
	scalarType := schema.NewScalarType()
	scalarType.Representation = schema.NewTypeRepresentationEnum(labelEnums).Encode()
	scb.ScalarTypes[labelEnumScalarName] = *scalarType

	// build aggregation argument
	aggregationObjectName := scb.createAggregationObjectType(objectName, labelEnumScalarName)
	arguments := createCollectionArguments(QueryTypeMetric)
	arguments[ArgumentKeyAggregations] = schema.ArgumentInfo{
		Description: utils.ToPtr("Aggregation operators for " + name),
		Type:        schema.NewArrayType(schema.NewNamedType(aggregationObjectName)).Encode(),
	}

	collection := schema.CollectionInfo{
		Name:                  name,
		Type:                  objectName,
		Arguments:             arguments,
		Description:           info.Description,
		ForeignKeys:           schema.CollectionInfoForeignKeys{},
		UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
	}

	scb.Collections[name] = collection

	return nil
}
