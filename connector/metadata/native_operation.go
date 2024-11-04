package metadata

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/iancoleman/strcase"
)

// The variable syntax for native queries is ${<name>} which is compatible with Grafana
var variableRegex = regexp.MustCompile(`\${(\w+)}`)
var allowedNativeQueryScalars = []ScalarName{ScalarString, ScalarDuration, ScalarInt64, ScalarFloat64}

// QueryType represents the type of the query
type QueryType string

const (
	QueryTypeStream = "stream"
	QueryTypeMetric = "metric"
)

// NativeOperations the list of native query and mutation definitions
type NativeOperations struct {
	// The definition map of native queries
	Queries map[string]NativeQuery `json:"queries" yaml:"queries"`
}

// NativeQueryArgumentInfo the input argument
type NativeQueryArgumentInfo struct {
	// Description of the argument
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string  `json:"type" yaml:"type" jsonschema:"enum=Int64,enum=Float64,enum=String,enum=Duration"`
}

// NativeQuery contains the information a native query
type NativeQuery struct {
	// the type of native query
	Type QueryType `json:"type" yaml:"type" jsonschema:"enum=stream,enum=metric"`
	// The LogQL query string to use for the Native Query.
	// We can interpolate values using `${<varname>}` syntax,
	// such as {job="${<varname>}"}
	Query string `json:"query" yaml:"query"`
	// Description of the query
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	// Labels returned by the native query
	Labels map[string]LabelInfo `json:"labels" yaml:"labels"`
	// Information of input arguments
	Arguments map[string]NativeQueryArgumentInfo `json:"arguments" yaml:"arguments"`
}

func (scb *connectorSchemaBuilder) buildNativeQueries() error {
	for name, nq := range scb.Metadata.NativeOperations.Queries {
		if err := scb.checkDuplicatedOperation(name); err != nil {
			return err
		}
		if err := scb.buildNativeQuery(name, &nq); err != nil {
			return err
		}
	}

	return nil
}

func (scb *connectorSchemaBuilder) buildNativeQuery(name string, query *NativeQuery) error {
	arguments := createCollectionArguments(query.Type)
	for key, arg := range query.Arguments {
		if _, ok := arguments[key]; ok {
			return fmt.Errorf("argument `%s` is already used by the function", key)
		}
		scalarName := arg.Type
		if arg.Type != "" {
			if !slices.Contains(allowedNativeQueryScalars, ScalarName(arg.Type)) {
				return fmt.Errorf("%s: unsupported native query argument type %s; argument: %s ", name, scalarName, key)
			}
		} else {
			scalarName = string(ScalarString)
		}

		arguments[key] = schema.ArgumentInfo{
			Description: arg.Description,
			Type:        schema.NewNamedType(scalarName).Encode(),
		}
	}

	resultType := schema.ObjectType{
		Fields: createQueryResultValuesObjectFields(query.Type),
	}

	for key, label := range query.Labels {
		resultType.Fields[key] = schema.ObjectField{
			Description: label.Description,
			Type:        schema.NewNamedType(string(ScalarLabel)).Encode(),
		}
	}

	objectName := strcase.ToCamel(name)
	if _, ok := scb.ObjectTypes[objectName]; ok {
		objectName = fmt.Sprintf("%sResult", objectName)
	}
	scb.ObjectTypes[objectName] = resultType
	collection := schema.CollectionInfo{
		Name:                  name,
		Type:                  objectName,
		Arguments:             arguments,
		Description:           query.Description,
		ForeignKeys:           schema.CollectionInfoForeignKeys{},
		UniquenessConstraints: schema.CollectionInfoUniquenessConstraints{},
	}

	scb.Collections[name] = collection

	return nil
}

// FindNativeQueryVariableNames find possible variables in the native query
func FindNativeQueryVariableNames(query string) []string {
	matches := variableRegex.FindAllStringSubmatch(query, -1)

	results := make([]string, len(matches))
	for _, m := range matches {
		results = append(results, m[1])
	}

	return results
}

// ReplaceNativeQueryVariable replaces the native query with variable
func ReplaceNativeQueryVariable(query string, name string, value string) string {
	return strings.ReplaceAll(query, fmt.Sprintf("${%s}", name), value)
}
