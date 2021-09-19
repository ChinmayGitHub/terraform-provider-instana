package instana_test

import (
	"context"
	"fmt"
	"github.com/gessnerfl/terraform-provider-instana/utils"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"

	. "github.com/gessnerfl/terraform-provider-instana/instana"
	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	"github.com/gessnerfl/terraform-provider-instana/testutils"
)

const resourceApplicationConfigWithMatchSpecificationDefinitionTemplate = `
provider "instana" {
  api_token = "test-token"
  endpoint = "localhost:%d"
  default_name_prefix = "prefix"
  default_name_suffix = "suffix"
}

resource "instana_application_config" "example" {
  label = "name %d"
  scope = "INCLUDE_ALL_DOWNSTREAM"
  boundary_scope = "ALL"
  match_specification = "%s"
}
`

const serverResponseWithMatchSpecificationTemplate = `
{
	"id" : "%s",
	"label" : "prefix name %d suffix",
	"scope" : "INCLUDE_ALL_DOWNSTREAM",
	"boundaryScope" : "ALL",
	"matchSpecification" : {
		"type" : "BINARY_OP",
		"left" : {
			"type" : "BINARY_OP",
			"left" : {
				"type" : "LEAF",
				"key" : "entity.name",
				"entity" : "DESTINATION",
				"operator" : "CONTAINS",
				"value" : "foo"
			},
			"conjunction" : "AND",
			"right" : {
				"type" : "LEAF",
				"key" : "entity.type",
				"entity" : "DESTINATION",
				"operator" : "EQUALS",
				"value" : "mysql"
			}
		},
		"conjunction" : "OR",
		"right" : {
			"type" : "LEAF",
			"key" : "entity.type",
			"entity" : "SOURCE",
			"operator" : "EQUALS",
			"value" : "elasticsearch"
		}
	}
}
`

const resourceApplicationConfigWithTagFilterDefinitionTemplate = `
provider "instana" {
  api_token = "test-token"
  endpoint = "localhost:%d"
  default_name_prefix = "prefix"
  default_name_suffix = "suffix"
}

resource "instana_application_config" "example" {
  label = "name %d"
  scope = "INCLUDE_ALL_DOWNSTREAM"
  boundary_scope = "ALL"
  tag_filter = "%s"
}
`

const serverResponseWithTagFilterTemplate = `
{
	"id" : "%s",
	"label" : "prefix name %d suffix",
	"scope" : "INCLUDE_ALL_DOWNSTREAM",
	"boundaryScope" : "ALL",
	"tagFilterExpression" : {
		"type" : "EXPRESSION",
		"logicalOperator": "OR",
		"elements" : [
			{
				"type" : "EXPRESSION",
				"logicalOperator": "AND",
				"elements" : [
					{
						"type" : "TAG_FILTER",
						"name" : "entity.name",
						"entity" : "DESTINATION",
						"operator" : "CONTAINS",
						"stringValue" : "foo",
						"value" : "foo"
					},
					{
						"type" : "TAG_FILTER",
						"name" : "entity.type",
						"entity" : "DESTINATION",
						"operator" : "EQUALS",
						"stringValue" : "mysql",
						"value" : "mysql"
					}
				]
			},
			{
				"type" : "TAG_FILTER",
				"name" : "entity.type",
				"entity" : "SOURCE",
				"operator" : "EQUALS",
				"stringValue" : "elasticsearch",
				"value" : "elasticsearch"
			}
		]
	}
}
`

const (
	testApplicationConfigDefinition = "instana_application_config.example"
	defaultExpression               = "entity.name CONTAINS 'foo' AND entity.type EQUALS 'mysql' OR entity.type@src EQUALS 'elasticsearch'"
	defaultNormalizedExpression     = "entity.name@dest CONTAINS 'foo' AND entity.type@dest EQUALS 'mysql' OR entity.type@src EQUALS 'elasticsearch'"
	validMatchSpecification         = "entity.type EQUALS 'foo'"
	invalidMatchSpecification       = "entity.type bla bla bla"
	validTagFilter                  = "entity.type EQUALS 'foo'"
	invalidTagFilter                = "entity.type bla bla bla"
	defaultLabel                    = "label"
	entityName                      = "entity.name"
	entityType                      = "entity.type"
)

var defaultMatchSpecificationModel = restapi.NewBinaryOperator(
	restapi.NewBinaryOperator(
		restapi.NewComparisonExpression(entityName, restapi.MatcherExpressionEntityDestination, restapi.ContainsOperator, "foo"),
		restapi.LogicalAnd,
		restapi.NewComparisonExpression(entityType, restapi.MatcherExpressionEntityDestination, restapi.EqualsOperator, "mysql"),
	),
	restapi.LogicalOr,
	restapi.NewComparisonExpression(entityType, restapi.MatcherExpressionEntitySource, restapi.EqualsOperator, "elasticsearch"))

var defaultTagFilterModel = restapi.NewLogicalOrTagFilter([]restapi.TagFilterExpressionElement{
	restapi.NewLogicalAndTagFilter([]restapi.TagFilterExpressionElement{
		restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, entityName, restapi.ContainsOperator, utils.StringPtr("foo")),
		restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, entityType, restapi.EqualsOperator, utils.StringPtr("mysql")),
	}),
	restapi.NewStringTagFilter(restapi.TagFilterEntitySource, entityType, restapi.EqualsOperator, utils.StringPtr("elasticsearch")),
})

const applicationConfigID = "application-config-id"

func TestCRUDOfApplicationConfigWithMatchSpecificationResourceWithMockServer(t *testing.T) {
	httpServer := createMockHttpServerForResource(restapi.ApplicationConfigsResourcePath, serverResponseWithMatchSpecificationTemplate)
	httpServer.Start()
	defer httpServer.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			createApplicationConfigWithMatchSpecificationResourceTestStep(httpServer.GetPort(), 0),
			testStepImport(testApplicationConfigDefinition),
			createApplicationConfigWithMatchSpecificationResourceTestStep(httpServer.GetPort(), 1),
			testStepImport(testApplicationConfigDefinition),
		},
	})
}

func createApplicationConfigWithMatchSpecificationResourceTestStep(httpPort int, iteration int) resource.TestStep {
	config := fmt.Sprintf(resourceApplicationConfigWithMatchSpecificationDefinitionTemplate, httpPort, iteration, defaultExpression)
	return resource.TestStep{
		Config: config,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(testApplicationConfigDefinition, "id"),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldLabel, formatResourceName(iteration)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldFullLabel, formatResourceFullName(iteration)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeAllDownstream)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldMatchSpecification, defaultNormalizedExpression),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldTagFilter, ""),
		),
	}
}

func TestCRUDOfApplicationConfigWithTagFilterResourceWithMockServer(t *testing.T) {
	httpServer := createMockHttpServerForResource(restapi.ApplicationConfigsResourcePath, serverResponseWithTagFilterTemplate)
	httpServer.Start()
	defer httpServer.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			createApplicationConfigWithTagFilterResourceTestStep(httpServer.GetPort(), 0),
			testStepImport(testApplicationConfigDefinition),
			createApplicationConfigWithTagFilterResourceTestStep(httpServer.GetPort(), 1),
			testStepImport(testApplicationConfigDefinition),
		},
	})
}

func createApplicationConfigWithTagFilterResourceTestStep(httpPort int, iteration int) resource.TestStep {
	config := fmt.Sprintf(resourceApplicationConfigWithTagFilterDefinitionTemplate, httpPort, iteration, defaultExpression)
	return resource.TestStep{
		Config: config,
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttrSet(testApplicationConfigDefinition, "id"),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldLabel, formatResourceName(iteration)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldFullLabel, formatResourceFullName(iteration)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeAllDownstream)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll)),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldTagFilter, defaultNormalizedExpression),
			resource.TestCheckResourceAttr(testApplicationConfigDefinition, ApplicationConfigFieldMatchSpecification, ""),
		),
	}
}

func TestApplicationConfigSchemaDefinitionIsValid(t *testing.T) {
	schema := NewApplicationConfigResourceHandle().MetaData().Schema

	schemaAssert := testutils.NewTerraformSchemaAssert(schema, t)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(ApplicationConfigFieldLabel)
	schemaAssert.AssertSchemaIsComputedAndOfTypeString(ApplicationConfigFieldFullLabel)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeStringWithDefault(ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	schemaAssert.AssertSchemaIsOptionalAndOfTypeStringWithDefault(ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeDefault))
	schemaAssert.AssertSchemaIsOptionalAndOfTypeString(ApplicationConfigFieldMatchSpecification)
	require.Equal(t, []string{ApplicationConfigFieldMatchSpecification, ApplicationConfigFieldTagFilter}, schema[ApplicationConfigFieldMatchSpecification].ExactlyOneOf)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeString(ApplicationConfigFieldTagFilter)
	require.Equal(t, []string{ApplicationConfigFieldMatchSpecification, ApplicationConfigFieldTagFilter}, schema[ApplicationConfigFieldTagFilter].ExactlyOneOf)
}

func TestShouldReturnTrueWhenCheckingForSchemaDiffSuppressForMatchSpecificationOfApplicationConfigAndValueCanBeNormalizedAndOldAndNewNormalizedValueAreEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	oldValue := "entity.type@dest EQUALS 'foo'"
	newValue := "entity.type  EQUALS    'foo'"

	require.True(t, schema[ApplicationConfigFieldMatchSpecification].DiffSuppressFunc(ApplicationConfigFieldMatchSpecification, oldValue, newValue, nil))
}

func TestShouldReturnFalseWhenCheckingForSchemaDiffSuppressForMatchSpecificationOfApplicationConfigAndValueCanBeNormalizedAndOldAndNewNormalizedValueAreNotEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	oldValue := "entity.type@src EQUALS 'foo'"
	newValue := validMatchSpecification

	require.False(t, schema[ApplicationConfigFieldMatchSpecification].DiffSuppressFunc(ApplicationConfigFieldMatchSpecification, oldValue, newValue, nil))
}

func TestShouldReturnTrueWhenCheckingForSchemaDiffSuppressForMatchSpecificationOfApplicationConfigAndValueCannotBeNormalizedAndOldAndNewValueAreEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	invalidValue := invalidMatchSpecification

	require.True(t, schema[ApplicationConfigFieldMatchSpecification].DiffSuppressFunc(ApplicationConfigFieldMatchSpecification, invalidValue, invalidValue, nil))
}

func TestShouldReturnFalseWhenCheckingForSchemaDiffSuppressForMatchSpecificationOfApplicationConfigAndValueCannotBeNormalizedAndOldAndNewValueAreNotEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	oldValue := invalidMatchSpecification
	newValue := "entity.type foo foo foo"

	require.False(t, schema[ApplicationConfigFieldMatchSpecification].DiffSuppressFunc(ApplicationConfigFieldMatchSpecification, oldValue, newValue, nil))
}

func TestShouldReturnNormalizedValueForMatchSpecificationOfApplicationConfigWhenStateFuncIsCalledAndValueCanBeNormalized(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	expectedValue := "entity.type@dest EQUALS 'foo'"
	newValue := validMatchSpecification

	require.Equal(t, expectedValue, schema[ApplicationConfigFieldMatchSpecification].StateFunc(newValue))
}

func TestShouldReturnProvidedValueForMatchSpecificationOfApplicationConfigWhenStateFuncIsCalledAndValueCannotBeNormalized(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	value := invalidMatchSpecification

	require.Equal(t, value, schema[ApplicationConfigFieldMatchSpecification].StateFunc(value))
}

func TestShouldReturnNoErrorsAndWarningsWhenValidationOfMatchSpecificationOfApplicationConfigIsCalledAndValueCanBeParsed(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	value := validMatchSpecification

	warns, errs := schema[ApplicationConfigFieldMatchSpecification].ValidateFunc(value, ApplicationConfigFieldMatchSpecification)
	require.Empty(t, warns)
	require.Empty(t, errs)
}

func TestShouldReturnOneErrorAndNoWarningsWhenValidationOfMatchSpecificationOfApplicationConfigIsCalledAndValueCannotBeParsed(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	value := invalidMatchSpecification

	warns, errs := schema[ApplicationConfigFieldMatchSpecification].ValidateFunc(value, ApplicationConfigFieldMatchSpecification)
	require.Empty(t, warns)
	require.Len(t, errs, 1)
}

func TestShouldReturnTrueWhenCheckingForSchemaDiffSuppressForTagFilterOfApplicationConfigAndValueCanBeNormalizedAndOldAndNewNormalizedValueAreEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	oldValue := "entity.type@dest EQUALS 'foo'"
	newValue := "entity.type  EQUALS    'foo'"

	require.True(t, schema[ApplicationConfigFieldTagFilter].DiffSuppressFunc(ApplicationConfigFieldTagFilter, oldValue, newValue, nil))
}

func TestShouldReturnFalseWhenCheckingForSchemaDiffSuppressForTagFilterOfApplicationConfigAndValueCanBeNormalizedAndOldAndNewNormalizedValueAreNotEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	oldValue := "entity.type@src EQUALS 'foo'"
	newValue := validTagFilter

	require.False(t, schema[ApplicationConfigFieldTagFilter].DiffSuppressFunc(ApplicationConfigFieldTagFilter, oldValue, newValue, nil))
}

func TestShouldReturnTrueWhenCheckingForSchemaDiffSuppressForTagFilterOfApplicationConfigAndValueCannotBeNormalizedAndOldAndNewValueAreEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	invalidValue := invalidTagFilter

	require.True(t, schema[ApplicationConfigFieldTagFilter].DiffSuppressFunc(ApplicationConfigFieldTagFilter, invalidValue, invalidValue, nil))
}

func TestShouldReturnFalseWhenCheckingForSchemaDiffSuppressForTagFilterOfApplicationConfigAndValueCannotBeNormalizedAndOldAndNewValueAreNotEqual(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	oldValue := invalidTagFilter
	newValue := "entity.type foo foo foo"

	require.False(t, schema[ApplicationConfigFieldTagFilter].DiffSuppressFunc(ApplicationConfigFieldTagFilter, oldValue, newValue, nil))
}

func TestShouldReturnNormalizedValueForTagFilterOfApplicationConfigWhenStateFuncIsCalledAndValueCanBeNormalized(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	expectedValue := "entity.type@dest EQUALS 'foo'"
	newValue := validTagFilter

	require.Equal(t, expectedValue, schema[ApplicationConfigFieldTagFilter].StateFunc(newValue))
}

func TestShouldReturnProvidedValueForTagFilterOfApplicationConfigWhenStateFuncIsCalledAndValueCannotBeNormalized(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	value := invalidTagFilter

	require.Equal(t, value, schema[ApplicationConfigFieldTagFilter].StateFunc(value))
}

func TestShouldReturnNoErrorsAndWarningsWhenValidationOfTagFilterOfApplicationConfigIsCalledAndValueCanBeParsed(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	value := validTagFilter

	warns, errs := schema[ApplicationConfigFieldTagFilter].ValidateFunc(value, ApplicationConfigFieldTagFilter)
	require.Empty(t, warns)
	require.Empty(t, errs)
}

func TestShouldReturnOneErrorAndNoWarningsWhenValidationOfTagFilterOfApplicationConfigIsCalledAndValueCannotBeParsed(t *testing.T) {
	resourceHandle := NewApplicationConfigResourceHandle()
	schema := resourceHandle.MetaData().Schema
	value := invalidTagFilter

	warns, errs := schema[ApplicationConfigFieldTagFilter].ValidateFunc(value, ApplicationConfigFieldTagFilter)
	require.Empty(t, warns)
	require.Len(t, errs, 1)
}

func TestApplicationConfigResourceShouldHaveSchemaVersionTwo(t *testing.T) {
	require.Equal(t, 3, NewApplicationConfigResourceHandle().MetaData().SchemaVersion)
}

func TestApplicationConfigResourceShouldHaveTwoStateUpgraderForVersionZeroToOne(t *testing.T) {
	resourceHandler := NewApplicationConfigResourceHandle()

	require.Equal(t, 3, len(resourceHandler.StateUpgraders()))
	require.Equal(t, 0, resourceHandler.StateUpgraders()[0].Version)
	require.Equal(t, 1, resourceHandler.StateUpgraders()[1].Version)
	require.Equal(t, 2, resourceHandler.StateUpgraders()[2].Version)
}

func TestShouldMigrateApplicationConfigStateAndAddFullLabelWithSameValueAsLabelWhenMigratingFromVersion0To1(t *testing.T) {
	label := "Test Label"
	rawData := make(map[string]interface{})
	rawData[ApplicationConfigFieldLabel] = label
	meta := "dummy"
	ctx := context.Background()

	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[0].Upgrade(ctx, rawData, meta)

	require.Nil(t, err)
	require.Equal(t, label, result[ApplicationConfigFieldFullLabel])
}

func TestShouldMigrateEmptyApplicationConfigStateFromVersion0To1(t *testing.T) {
	rawData := make(map[string]interface{})
	meta := "dummy"
	ctx := context.Background()

	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[0].Upgrade(ctx, rawData, meta)

	require.Nil(t, err)
	require.Nil(t, result[ApplicationConfigFieldFullLabel])
}

func TestShouldHarmonizeMatchSpecificationWhenMigratingStateFromVersion1To2(t *testing.T) {
	input := "entity.name EQUALS 'foo'"
	expectedResult := "entity.name@dest EQUALS 'foo'"
	rawData := make(map[string]interface{})
	rawData[ApplicationConfigFieldMatchSpecification] = input
	meta := "dummy"
	ctx := context.Background()

	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[1].Upgrade(ctx, rawData, meta)

	require.Nil(t, err)
	require.Equal(t, input, result[ApplicationConfigFieldMatchSpecification])
	require.Equal(t, expectedResult, result[ApplicationConfigFieldNormalizedMatchSpecification])
}

func TestShouldFailToHarmonizeMatchSpecificationWhenMigratingStateFromVersion1To2(t *testing.T) {
	input := "invalid match spec"
	rawData := make(map[string]interface{})
	rawData[ApplicationConfigFieldMatchSpecification] = input
	meta := "dummy"
	ctx := context.Background()

	_, err := NewApplicationConfigResourceHandle().StateUpgraders()[1].Upgrade(ctx, rawData, meta)

	require.Error(t, err)
}

func TestShouldMigrateEmptyApplicationConfigStateFromVersion1To2(t *testing.T) {
	rawData := make(map[string]interface{})
	meta := "dummy"
	ctx := context.Background()

	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[1].Upgrade(ctx, rawData, meta)

	require.Nil(t, err)
	require.Nil(t, result[ApplicationConfigFieldMatchSpecification])
	require.Nil(t, result[ApplicationConfigFieldNormalizedMatchSpecification])
}

func TestShouldRemoveHarmonizedMatchSpecificationWhenMigratingApplicationConfigStateFromVersion2To3AndHarmonizedMatchSpecificationIsSet(t *testing.T) {
	rawData := make(map[string]interface{})
	meta := "dummy"
	ctx := context.Background()

	rawData["id"] = applicationConfigID
	rawData[ApplicationConfigFieldFullLabel] = defaultLabel
	rawData[ApplicationConfigFieldMatchSpecification] = defaultExpression
	rawData[ApplicationConfigFieldScope] = string(restapi.ApplicationConfigScopeIncludeNoDownstream)
	rawData[ApplicationConfigFieldBoundaryScope] = string(restapi.BoundaryScopeAll)
	expectedResult := copyMap(rawData)
	rawData[ApplicationConfigFieldNormalizedMatchSpecification] = defaultNormalizedExpression

	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[2].Upgrade(ctx, rawData, meta)

	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestShouldRemoveHarmonizedMatchSpecificationWhenMigratingApplicationConfigStateFromVersion2To3AndNoHarmonizedMatchSpecificationIsSet(t *testing.T) {
	rawData := make(map[string]interface{})
	meta := "dummy"
	ctx := context.Background()

	rawData["id"] = applicationConfigID
	rawData[ApplicationConfigFieldFullLabel] = defaultLabel
	rawData[ApplicationConfigFieldMatchSpecification] = defaultExpression
	rawData[ApplicationConfigFieldScope] = string(restapi.ApplicationConfigScopeIncludeNoDownstream)
	rawData[ApplicationConfigFieldBoundaryScope] = string(restapi.BoundaryScopeAll)
	expectedResult := copyMap(rawData)

	result, err := NewApplicationConfigResourceHandle().StateUpgraders()[2].Upgrade(ctx, rawData, meta)

	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestShouldReturnCorrectResourceNameForApplicationConfigResource(t *testing.T) {
	name := NewApplicationConfigResourceHandle().MetaData().ResourceName

	require.Equal(t, name, "instana_application_config")
}

func TestShouldUpdateApplicationConfigTerraformResourceStateFromModelWhenMatchSpecificationIsProvided(t *testing.T) {
	fullLabel := "prefix label suffix"
	applicationConfig := restapi.ApplicationConfig{
		ID:                 applicationConfigID,
		Label:              fullLabel,
		MatchSpecification: defaultMatchSpecificationModel,
		Scope:              restapi.ApplicationConfigScopeIncludeNoDownstream,
		BoundaryScope:      restapi.BoundaryScopeAll,
	}

	testHelper := NewTestHelper(t)
	sut := NewApplicationConfigResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &applicationConfig, testHelper.ResourceFormatter())

	require.NoError(t, err)
	require.Equal(t, applicationConfigID, resourceData.Id())
	require.Equal(t, defaultLabel, resourceData.Get(ApplicationConfigFieldLabel))
	require.Equal(t, fullLabel, resourceData.Get(ApplicationConfigFieldFullLabel))
	require.Equal(t, defaultNormalizedExpression, resourceData.Get(ApplicationConfigFieldMatchSpecification))
	_, tagFilterSet := resourceData.GetOk(ApplicationConfigFieldTagFilter)
	require.False(t, tagFilterSet)
	require.Equal(t, string(restapi.ApplicationConfigScopeIncludeNoDownstream), resourceData.Get(ApplicationConfigFieldScope))
	require.Equal(t, string(restapi.BoundaryScopeAll), resourceData.Get(ApplicationConfigFieldBoundaryScope))
}

func TestShouldFailToUpdateApplicationConfigTerraformResourceStateFromModelWhenMatchSpecificationIsNotValid(t *testing.T) {
	comparison := restapi.NewComparisonExpression(entityName, restapi.MatcherExpressionEntityDestination, "INVALID", "foo")
	applicationConfig := restapi.ApplicationConfig{
		ID:                 applicationConfigID,
		Label:              defaultLabel,
		MatchSpecification: comparison,
		Scope:              restapi.ApplicationConfigScopeIncludeNoDownstream,
	}

	testHelper := NewTestHelper(t)
	sut := NewApplicationConfigResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &applicationConfig, testHelper.ResourceFormatter())

	require.Error(t, err)
}

func TestShouldUpdateApplicationConfigTerraformResourceStateFromModelWhenTagFilterIsProvided(t *testing.T) {
	fullLabel := "prefix label suffix"
	applicationConfig := restapi.ApplicationConfig{
		ID:                  applicationConfigID,
		Label:               fullLabel,
		TagFilterExpression: defaultTagFilterModel,
		Scope:               restapi.ApplicationConfigScopeIncludeNoDownstream,
		BoundaryScope:       restapi.BoundaryScopeAll,
	}

	testHelper := NewTestHelper(t)
	sut := NewApplicationConfigResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &applicationConfig, testHelper.ResourceFormatter())

	require.NoError(t, err)
	require.Equal(t, applicationConfigID, resourceData.Id())
	require.Equal(t, defaultLabel, resourceData.Get(ApplicationConfigFieldLabel))
	require.Equal(t, fullLabel, resourceData.Get(ApplicationConfigFieldFullLabel))
	_, matchSpecificationSet := resourceData.GetOk(ApplicationConfigFieldMatchSpecification)
	require.False(t, matchSpecificationSet)
	require.Equal(t, defaultNormalizedExpression, resourceData.Get(ApplicationConfigFieldTagFilter))
	require.Equal(t, string(restapi.ApplicationConfigScopeIncludeNoDownstream), resourceData.Get(ApplicationConfigFieldScope))
	require.Equal(t, string(restapi.BoundaryScopeAll), resourceData.Get(ApplicationConfigFieldBoundaryScope))
}

func TestShouldFailToUpdateApplicationConfigTerraformResourceStateFromModelWhenTagFilterIsNotValid(t *testing.T) {
	comparison := restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, entityName, "INVALID", utils.StringPtr("foo"))
	applicationConfig := restapi.ApplicationConfig{
		ID:                  applicationConfigID,
		Label:               defaultLabel,
		TagFilterExpression: comparison,
		Scope:               restapi.ApplicationConfigScopeIncludeNoDownstream,
	}

	testHelper := NewTestHelper(t)
	sut := NewApplicationConfigResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &applicationConfig, testHelper.ResourceFormatter())

	require.Error(t, err)
}

func TestShouldFailToUpdateApplicationConfigTerraformResourceStateFromModelWhenNeitherMatchExpressionNorTagFilterIsProvided(t *testing.T) {
	applicationConfig := restapi.ApplicationConfig{
		ID:    applicationConfigID,
		Label: defaultLabel,
		Scope: restapi.ApplicationConfigScopeIncludeNoDownstream,
	}

	testHelper := NewTestHelper(t)
	sut := NewApplicationConfigResourceHandle()
	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &applicationConfig, testHelper.ResourceFormatter())

	require.Error(t, err)
	require.Contains(t, err.Error(), "either match specification or tag filter is required")
}

func TestShouldSuccessfullyConvertApplicationConfigStateToDataModelWhenMatchSpecificationIsAvailable(t *testing.T) {
	testHelper := NewTestHelper(t)
	resourceHandle := NewApplicationConfigResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.SetId(applicationConfigID)
	resourceData.Set(ApplicationConfigFieldFullLabel, defaultLabel)
	resourceData.Set(ApplicationConfigFieldMatchSpecification, defaultExpression)
	resourceData.Set(ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	resourceData.Set(ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll))

	result, err := resourceHandle.MapStateToDataObject(resourceData, testHelper.ResourceFormatter())

	require.Nil(t, err)
	require.IsType(t, &restapi.ApplicationConfig{}, result)
	require.Equal(t, applicationConfigID, result.GetIDForResourcePath())
	require.Equal(t, defaultLabel, result.(*restapi.ApplicationConfig).Label)
	require.Equal(t, defaultMatchSpecificationModel, result.(*restapi.ApplicationConfig).MatchSpecification)
	require.Nil(t, result.(*restapi.ApplicationConfig).TagFilterExpression)
	require.Equal(t, restapi.ApplicationConfigScopeIncludeNoDownstream, result.(*restapi.ApplicationConfig).Scope)
	require.Equal(t, restapi.BoundaryScopeAll, result.(*restapi.ApplicationConfig).BoundaryScope)
}

func TestShouldFailToConvertApplicationConfigStateToDataModelWhenMatchSpecificationIsNotValid(t *testing.T) {
	testHelper := NewTestHelper(t)
	resourceHandle := NewApplicationConfigResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.SetId(applicationConfigID)
	resourceData.Set(ApplicationConfigFieldFullLabel, defaultLabel)
	resourceData.Set(ApplicationConfigFieldMatchSpecification, "INVALID")
	resourceData.Set(ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	resourceData.Set(ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll))

	_, err := resourceHandle.MapStateToDataObject(resourceData, testHelper.ResourceFormatter())

	require.NotNil(t, err)
}

func TestShouldSuccessfullyConvertApplicationConfigStateToDataModelWhenTagFilterIsAvailable(t *testing.T) {
	testHelper := NewTestHelper(t)
	resourceHandle := NewApplicationConfigResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.SetId(applicationConfigID)
	resourceData.Set(ApplicationConfigFieldFullLabel, defaultLabel)
	resourceData.Set(ApplicationConfigFieldTagFilter, defaultExpression)
	resourceData.Set(ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	resourceData.Set(ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll))

	result, err := resourceHandle.MapStateToDataObject(resourceData, testHelper.ResourceFormatter())

	require.Nil(t, err)
	require.IsType(t, &restapi.ApplicationConfig{}, result)
	require.Equal(t, applicationConfigID, result.GetIDForResourcePath())
	require.Equal(t, defaultLabel, result.(*restapi.ApplicationConfig).Label)
	require.Nil(t, result.(*restapi.ApplicationConfig).MatchSpecification)
	require.Equal(t, defaultTagFilterModel, result.(*restapi.ApplicationConfig).TagFilterExpression)
	require.Equal(t, restapi.ApplicationConfigScopeIncludeNoDownstream, result.(*restapi.ApplicationConfig).Scope)
	require.Equal(t, restapi.BoundaryScopeAll, result.(*restapi.ApplicationConfig).BoundaryScope)
}

func TestShouldFailToConvertApplicationConfigStateToDataModelWhenTagFilterIsNotValid(t *testing.T) {
	testHelper := NewTestHelper(t)
	resourceHandle := NewApplicationConfigResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.SetId(applicationConfigID)
	resourceData.Set(ApplicationConfigFieldFullLabel, defaultLabel)
	resourceData.Set(ApplicationConfigFieldTagFilter, "INVALID")
	resourceData.Set(ApplicationConfigFieldScope, string(restapi.ApplicationConfigScopeIncludeNoDownstream))
	resourceData.Set(ApplicationConfigFieldBoundaryScope, string(restapi.BoundaryScopeAll))

	_, err := resourceHandle.MapStateToDataObject(resourceData, testHelper.ResourceFormatter())

	require.NotNil(t, err)
}
