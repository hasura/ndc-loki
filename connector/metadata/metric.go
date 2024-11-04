package metadata

import (
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

func (scb *connectorSchemaBuilder) createAggregationObjectType(objectName string, labelEnumScalarName string) string {
	aggregationObjectName := objectName + "Aggregation"
	aggregationKInputObjectName := aggregationObjectName + "KInput"
	aggregationGroupByObjectName := aggregationObjectName + "GroupBy"
	aggregationRangeUnwrapObjectName := aggregationObjectName + "RangeUnwrap"
	aggregationQuantileOverTimeObjectName := aggregationObjectName + "QuantileOverTime"

	groupByType := schema.NewNullableType(schema.NewNamedType(aggregationGroupByObjectName)).Encode()
	rangeUnwrapType := schema.NewNullableType(schema.NewNamedType(aggregationRangeUnwrapObjectName)).Encode()
	durationType := schema.NewNullableNamedType(string(ScalarDuration)).Encode()
	orderingType := schema.NewNullableNamedType(string(ScalarOrdering)).Encode()
	scb.ObjectTypes[aggregationKInputObjectName] = createAggregateIntInputObjectType(KKey, labelEnumScalarName)
	scb.ObjectTypes[aggregationGroupByObjectName] = createAggregateGroupByObjectType(labelEnumScalarName)
	scb.ObjectTypes[aggregationRangeUnwrapObjectName] = createAggregateRangeUnwrapInputObjectType(labelEnumScalarName)
	scb.ObjectTypes[aggregationQuantileOverTimeObjectName] = createQuantileOverTimeInputObjectType(labelEnumScalarName)

	// unwrap: rate_counter, sum_over_time,
	aggregationObject := schema.ObjectType{
		Fields: schema.ObjectTypeFields{
			string(Sum): schema.ObjectField{
				Description: utils.ToPtr("Calculate sum over labels"),
				Type:        groupByType,
			},
			string(Avg): schema.ObjectField{
				Description: utils.ToPtr("Calculate the average over labels"),
				Type:        groupByType,
			},
			string(Min): schema.ObjectField{
				Description: utils.ToPtr("Select minimum over labels"),
				Type:        groupByType,
			},
			string(Max): schema.ObjectField{
				Description: utils.ToPtr("Select maximum over labels"),
				Type:        groupByType,
			},
			string(Stddev): schema.ObjectField{
				Description: utils.ToPtr("Calculate the population standard deviation over labels"),
				Type:        groupByType,
			},
			string(Stdvar): schema.ObjectField{
				Description: utils.ToPtr("Calculate the population standard variance over labels"),
				Type:        groupByType,
			},
			string(Count): schema.ObjectField{
				Description: utils.ToPtr("Count number of elements in the vector"),
				Type:        groupByType,
			},
			string(TopK): schema.ObjectField{
				Description: utils.ToPtr("Largest k elements by sample value"),
				Type:        schema.NewNullableNamedType(aggregationKInputObjectName).Encode(),
			},
			string(BottomK): schema.ObjectField{
				Description: utils.ToPtr("Smallest k elements by sample value"),
				Type:        schema.NewNullableNamedType(aggregationKInputObjectName).Encode(),
			},
			string(AbsentOverTime): schema.ObjectField{
				Description: utils.ToPtr("Returns an empty vector if the range vector passed to it has any elements (floats or native histograms) and a 1-element vector with the value 1 if the range vector passed to it has no elements"),
				Type:        durationType,
			},
			string(Rate): schema.ObjectField{
				Description: utils.ToPtr("Calculates the number of entries per second."),
				Type:        durationType,
			},
			string(RateCounter): schema.ObjectField{
				Description: utils.ToPtr("Calculates the number of entries per second."),
				Type:        rangeUnwrapType,
			},
			string(BytesRate): schema.ObjectField{
				Description: utils.ToPtr("Calculates the number of bytes per second for each stream."),
				Type:        durationType,
			},
			string(Sort): schema.ObjectField{
				Description: utils.ToPtr("Returns vector elements sorted by their sample values. Native histograms are sorted by their sum of observations"),
				Type:        orderingType,
			},
			string(AvgOverTime): schema.ObjectField{
				Description: utils.ToPtr("The average of all values in the specified interval."),
				Type:        rangeUnwrapType,
			},
			string(MinOverTime): schema.ObjectField{
				Description: utils.ToPtr("The minimum of all values in the specified interval."),
				Type:        rangeUnwrapType,
			},
			string(MaxOverTime): schema.ObjectField{
				Description: utils.ToPtr("The maximum of all values in the specified interval."),
				Type:        rangeUnwrapType,
			},
			string(SumOverTime): schema.ObjectField{
				Description: utils.ToPtr("The sum of all values in the specified interval"),
				Type:        rangeUnwrapType,
			},
			string(CountOverTime): schema.ObjectField{
				Description: utils.ToPtr("The count of all values in the specified interval"),
				Type:        durationType,
			},
			string(QuantileOverTime): schema.ObjectField{
				Description: utils.ToPtr("The φ-quantile (0 ≤ φ ≤ 1) of the values in the specified interval."),
				Type:        schema.NewNullableNamedType(aggregationQuantileOverTimeObjectName).Encode(),
			},
			string(StddevOverTime): schema.ObjectField{
				Description: utils.ToPtr("The population standard deviation of the values in the specified interval."),
				Type:        rangeUnwrapType,
			},
			string(StdvarOverTime): schema.ObjectField{
				Description: utils.ToPtr("The population standard variance of the values in the specified interval."),
				Type:        rangeUnwrapType,
			},
			string(FirstOverTime): schema.ObjectField{
				Description: utils.ToPtr("The first of all values in the specified interval."),
				Type:        rangeUnwrapType,
			},
			string(LastOverTime): schema.ObjectField{
				Description: utils.ToPtr("The last of all values in the specified interval."),
				Type:        rangeUnwrapType,
			},
			string(BytesOverTime): schema.ObjectField{
				Description: utils.ToPtr("Counts the amount of bytes used by each log stream for a given range."),
				Type:        durationType,
			},
		},
	}

	scb.ObjectTypes[aggregationObjectName] = aggregationObject

	return aggregationObjectName
}

func createQuantileOverTimeInputObjectType(labelEnumScalarName string) schema.ObjectType {
	objectType := createAggregateRangeUnwrapInputObjectType(labelEnumScalarName)
	objectType.Fields[QuantileKey] = schema.ObjectField{
		Type: schema.NewNamedType(string(ScalarFloat64)).Encode(),
	}

	return objectType
}

func createAggregateRangeUnwrapInputObjectType(labelEnumScalarName string) schema.ObjectType {
	return schema.ObjectType{
		Fields: schema.ObjectTypeFields{
			RangeKey: schema.ObjectField{
				Type: schema.NewNamedType(string(ScalarDuration)).Encode(),
			},
			UnwrapKey: schema.ObjectField{
				Type: schema.NewNamedType(labelEnumScalarName).Encode(),
			},
			ConversionFunctionKey: schema.ObjectField{
				Type: schema.NewNullableType(schema.NewNamedType(string(ScalarConversionFunction))).Encode(),
			},
		},
	}
}

func createAggregateGroupByObjectType(labelEnumScalarName string) schema.ObjectType {
	enumType := schema.NewNullableType(schema.NewArrayType(schema.NewNamedType(labelEnumScalarName))).Encode()

	return schema.ObjectType{
		Fields: schema.ObjectTypeFields{
			ByKey: schema.ObjectField{
				Type: enumType,
			},
			WithoutKey: schema.ObjectField{
				Type: enumType,
			},
		},
	}
}

func createAggregateIntInputObjectType(intKey string, labelEnumScalarName string) schema.ObjectType {
	objectType := createAggregateGroupByObjectType(labelEnumScalarName)
	objectType.Fields[intKey] = schema.ObjectField{
		Type: schema.NewNamedType(string(ScalarInt64)).Encode(),
	}

	return objectType
}
