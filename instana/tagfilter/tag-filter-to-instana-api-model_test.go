package tagfilter_test

import (
	"fmt"
	"github.com/gessnerfl/terraform-provider-instana/utils"
	"testing"

	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	. "github.com/gessnerfl/terraform-provider-instana/instana/tagfilter"
	"github.com/stretchr/testify/require"
)

const (
	entitySpecKey = "key"
)

func TestShouldMapComparisonToRepresentationOfInstanaAPI(t *testing.T) {
	for _, v := range restapi.SupportedComparisonOperators {
		t.Run(fmt.Sprintf("test comparison of string value using operatore %s", v), createTestShouldMapStringComparisonToRepresentationOfInstanaAPI(v))
		t.Run(fmt.Sprintf("test comparison of number value using operatore of %s", v), createTestShouldMapNumberComparisonToRepresentationOfInstanaAPI(v))
		t.Run(fmt.Sprintf("test comparison of boolean value using operatore of %s", v), createTestShouldMapBooleanComparisonToRepresentationOfInstanaAPI(v))
		t.Run(fmt.Sprintf("test comparison of tag using operatore of %s", v), createTestShouldMapTagComparisonToRepresentationOfInstanaAPI(v))
	}
}

func createTestShouldMapStringComparisonToRepresentationOfInstanaAPI(operator restapi.ExpressionOperator) func(*testing.T) {
	return func(t *testing.T) {
		expr := &FilterExpression{
			Expression: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							Comparison: &ComparisonExpression{
								Entity:      &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator:    Operator(operator),
								StringValue: utils.StringPtr("value"),
							},
						},
					},
				},
			},
		}

		expectedResult := restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, operator, "value")
		runTestCaseForMappingToAPI(expr, expectedResult, t)
	}
}

func createTestShouldMapNumberComparisonToRepresentationOfInstanaAPI(operator restapi.ExpressionOperator) func(*testing.T) {
	numberValue := int64(1234)
	return func(t *testing.T) {
		expr := &FilterExpression{
			Expression: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							Comparison: &ComparisonExpression{
								Entity:      &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator:    Operator(operator),
								NumberValue: &numberValue,
							},
						},
					},
				},
			},
		}

		expectedResult := restapi.NewNumberTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, operator, numberValue)
		runTestCaseForMappingToAPI(expr, expectedResult, t)
	}
}

func createTestShouldMapBooleanComparisonToRepresentationOfInstanaAPI(operator restapi.ExpressionOperator) func(*testing.T) {
	boolValue := true
	return func(t *testing.T) {
		expr := &FilterExpression{
			Expression: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							Comparison: &ComparisonExpression{
								Entity:       &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator:     Operator(operator),
								BooleanValue: &boolValue,
							},
						},
					},
				},
			},
		}

		expectedResult := restapi.NewBooleanTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, operator, boolValue)
		runTestCaseForMappingToAPI(expr, expectedResult, t)
	}
}

func createTestShouldMapTagComparisonToRepresentationOfInstanaAPI(operator restapi.ExpressionOperator) func(*testing.T) {
	key := "key"
	value := "value"
	return func(t *testing.T) {
		expr := &FilterExpression{
			Expression: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							Comparison: &ComparisonExpression{
								Entity:      &EntitySpec{Identifier: entitySpecKey, TagKey: &key, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator:    Operator(operator),
								StringValue: &value,
							},
						},
					},
				},
			},
		}

		expectedResult := restapi.NewTagTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, operator, key, value)
		runTestCaseForMappingToAPI(expr, expectedResult, t)
	}
}

func TestShouldMapTagComparisonToRepresentationOfInstanaAPIUsingAStringValue(t *testing.T) {
	key := "key"
	value := "value"
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						Comparison: &ComparisonExpression{
							Entity:      &EntitySpec{Identifier: entitySpecKey, TagKey: &key, Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator:    Operator(restapi.EqualsOperator),
							StringValue: &value,
						},
					},
				},
			},
		},
	}

	expectedResult := restapi.NewTagTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.EqualsOperator, key, value)
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapTagComparisonToRepresentationOfInstanaAPIUsingANumberValue(t *testing.T) {
	key := "key"
	value := int64(1234)
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						Comparison: &ComparisonExpression{
							Entity:      &EntitySpec{Identifier: entitySpecKey, TagKey: &key, Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator:    Operator(restapi.EqualsOperator),
							NumberValue: &value,
						},
					},
				},
			},
		},
	}

	expectedResult := restapi.NewTagTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.EqualsOperator, key, "1234")
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapTagComparisonToRepresentationOfInstanaAPIUsingABooleanValue(t *testing.T) {
	key := "key"
	value := true
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						Comparison: &ComparisonExpression{
							Entity:       &EntitySpec{Identifier: entitySpecKey, TagKey: &key, Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator:     Operator(restapi.EqualsOperator),
							BooleanValue: &value,
						},
					},
				},
			},
		},
	}

	expectedResult := restapi.NewTagTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.EqualsOperator, key, "true")
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapUnaryOperatorToRepresentationOfInstanaAPI(t *testing.T) {
	for _, v := range restapi.SupportedUnaryExpressionOperators {
		t.Run(fmt.Sprintf("test mapping of %s", v), createTestShouldMapUnaryOperatorToRepresentationOfInstanaAPI(v))
	}
}

func createTestShouldMapUnaryOperatorToRepresentationOfInstanaAPI(operatorName restapi.ExpressionOperator) func(*testing.T) {
	return func(t *testing.T) {
		expr := &FilterExpression{
			Expression: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							UnaryOperation: &UnaryOperationExpression{
								Entity:   &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator: Operator(operatorName),
							},
						},
					},
				},
			},
		}

		expectedResult := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, operatorName)
		runTestCaseForMappingToAPI(expr, expectedResult, t)
	}
}

func TestShouldMapLogicalAndExpression(t *testing.T) {
	logicalAnd := Operator(restapi.LogicalAnd)
	primaryExpression := PrimaryExpression{
		UnaryOperation: &UnaryOperationExpression{
			Entity:   &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
			Operator: Operator(restapi.IsEmptyOperator),
		},
	}
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left:     &BracketExpression{Primary: &primaryExpression},
				Operator: &logicalAnd,
				Right: &LogicalAndExpression{
					Left: &BracketExpression{Primary: &primaryExpression},
				},
			},
		},
	}

	expectedPrimaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.IsEmptyOperator)
	expectedResult := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedPrimaryExpression})
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapLogicalAndExpressionWithNestedAnd(t *testing.T) {
	logicalAnd := Operator(restapi.LogicalAnd)
	primaryExpression := PrimaryExpression{
		UnaryOperation: &UnaryOperationExpression{
			Entity:   &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
			Operator: Operator(restapi.IsEmptyOperator),
		},
	}
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left:     &BracketExpression{Primary: &primaryExpression},
				Operator: &logicalAnd,
				Right: &LogicalAndExpression{
					Left:     &BracketExpression{Primary: &primaryExpression},
					Operator: &logicalAnd,
					Right: &LogicalAndExpression{
						Left: &BracketExpression{Primary: &primaryExpression},
					},
				},
			},
		},
	}

	expectedPrimaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.IsEmptyOperator)
	expectedResult := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedPrimaryExpression, expectedPrimaryExpression})
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapLogicalAndExpressionWithNestedOrInBrackets(t *testing.T) {
	logicalAnd := Operator(restapi.LogicalAnd)
	logicalOr := Operator(restapi.LogicalOr)
	primaryExpression := PrimaryExpression{
		UnaryOperation: &UnaryOperationExpression{
			Entity:   &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
			Operator: Operator(restapi.IsEmptyOperator),
		},
	}
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left:     &BracketExpression{Primary: &primaryExpression},
				Operator: &logicalAnd,
				Right: &LogicalAndExpression{
					Left: &BracketExpression{
						Bracket: &LogicalOrExpression{
							Left:     &LogicalAndExpression{Left: &BracketExpression{Primary: &primaryExpression}},
							Operator: &logicalOr,
							Right: &LogicalOrExpression{
								Left: &LogicalAndExpression{Left: &BracketExpression{Primary: &primaryExpression}},
							},
						},
					},
				},
			},
		},
	}

	expectedPrimaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.IsEmptyOperator)
	expectedOrExpression := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedPrimaryExpression})
	expectedResult := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedOrExpression})
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapLogicalOrExpression(t *testing.T) {
	logicalOr := Operator(restapi.LogicalOr)
	primaryExpression := PrimaryExpression{
		UnaryOperation: &UnaryOperationExpression{
			Entity:   &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
			Operator: Operator(restapi.IsEmptyOperator),
		},
	}
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{Primary: &primaryExpression},
			},
			Operator: &logicalOr,
			Right: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{Primary: &primaryExpression},
				},
			},
		},
	}

	expectedPrimaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.IsEmptyOperator)
	expectedResult := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedPrimaryExpression})
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapLogicalOrExpressionWithNestedOr(t *testing.T) {
	logicalOr := Operator(restapi.LogicalOr)
	primaryExpression := PrimaryExpression{
		UnaryOperation: &UnaryOperationExpression{
			Entity:   &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
			Operator: Operator(restapi.IsEmptyOperator),
		},
	}
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{Primary: &primaryExpression},
			},
			Operator: &logicalOr,
			Right: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{Primary: &primaryExpression},
				},
				Operator: &logicalOr,
				Right: &LogicalOrExpression{
					Left: &LogicalAndExpression{
						Left: &BracketExpression{Primary: &primaryExpression},
					},
				},
			},
		},
	}

	expectedPrimaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.IsEmptyOperator)
	expectedResult := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedPrimaryExpression, expectedPrimaryExpression})
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func TestShouldMapLogicalOrExpressionWithNestedAndInBrackets(t *testing.T) {
	logicalOr := Operator(restapi.LogicalOr)
	logicalAnd := Operator(restapi.LogicalAnd)
	primaryExpression := PrimaryExpression{
		UnaryOperation: &UnaryOperationExpression{
			Entity:   &EntitySpec{Identifier: entitySpecKey, Origin: utils.StringPtr(EntityOriginDestination.Key())},
			Operator: Operator(restapi.IsEmptyOperator),
		},
	}
	expr := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{Primary: &primaryExpression},
			},
			Operator: &logicalOr,
			Right: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Bracket: &LogicalOrExpression{
							Left: &LogicalAndExpression{
								Left:     &BracketExpression{Primary: &primaryExpression},
								Operator: &logicalAnd,
								Right: &LogicalAndExpression{
									Left: &BracketExpression{Primary: &primaryExpression},
								},
							},
						},
					},
				},
			},
		},
	}

	expectedPrimaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, entitySpecKey, restapi.IsEmptyOperator)
	expectedAndExpression := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedPrimaryExpression})
	expectedResult := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{expectedPrimaryExpression, expectedAndExpression})
	runTestCaseForMappingToAPI(expr, expectedResult, t)
}

func runTestCaseForMappingToAPI(input *FilterExpression, expectedResult *restapi.TagFilter, t *testing.T) {
	mapper := NewMapper()
	result := mapper.ToAPIModel(input)

	require.Equal(t, expectedResult, result)
}
