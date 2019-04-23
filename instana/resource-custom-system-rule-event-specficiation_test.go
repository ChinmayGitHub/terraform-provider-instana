package instana_test

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	. "github.com/gessnerfl/terraform-provider-instana/instana"
	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	"github.com/gessnerfl/terraform-provider-instana/mocks"
	"github.com/gessnerfl/terraform-provider-instana/testutils"
)

var testCustomSystemEventProviders = map[string]terraform.ResourceProvider{
	"instana": Provider(),
}

const resourceCustomSystemEventDefinitionTemplate = `
provider "instana" {
  api_token = "test-token"
  endpoint = "localhost:{{PORT}}"
}

resource "instana_custom_event_spec_system_rule" "example" {
  name = "name"
  entity_type = "entity_type"
  query = "query"
  enabled = true
  triggering = true
  description = "description"
  expiration_time = "60000"
	rule_severity = "warning"
	rule_system_rule_id = "system-rule-id"
	downstream_integration_ids = [ "integration-id-1", "integration-id-2" ]
	downstream_broadcast_to_all_alerting_configs = true
}
`

const (
	customSystemEventApiPath        = restapi.CustomEventSpecificationResourcePath + "/{id}"
	testCustomSystemEventDefinition = "instana_custom_event_spec_system_rule.example"

	customSystemEventID                       = "custom-system-event-id"
	customSystemEventName                     = "name"
	customSystemEventMetricName               = "metric_name"
	customSystemEventEntityType               = "entity_type"
	customSystemEventQuery                    = "query"
	customSystemEventExpirationTime           = 60000
	customSystemEventDescription              = "description"
	customSystemEventRuleSystemRuleId         = "system-rule-id"
	customSystemEventDownStringIntegrationId1 = "integration-id-1"
	customSystemEventDownStringIntegrationId2 = "integration-id-2"
)

var customSystemEventRuleSeverity = restapi.SeverityWarning.GetTerraformRepresentation()

func TestCRUDOfCustomSystemEventResourceWithMockServer(t *testing.T) {
	testutils.DeactivateTLSServerCertificateVerification()
	httpServer := testutils.NewTestHTTPServer()
	httpServer.AddRoute(http.MethodPut, customSystemEventApiPath, testutils.EchoHandlerFunc)
	httpServer.AddRoute(http.MethodDelete, customSystemEventApiPath, testutils.EchoHandlerFunc)
	httpServer.AddRoute(http.MethodGet, customSystemEventApiPath, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		json := strings.ReplaceAll(`
		{
			"id" : "{{id}}",
			"name" : "name",
			"entityType" : "entity_type",
			"query" : "query",
			"enabled" : true,
			"triggering" : true,
			"description" : "description",
			"expirationTime" : 60000,
			"rules" : [ { "ruleType" : "system", "severity" : 5, "systemRuleId" : "system-rule-id" } ],
			"downstream" : {
				"integrationIds" : ["integration-id-1", "integration-id-2"],
				"broadcastToAllAlertingConfigs" : true
			}
		}
		`, "{{id}}", vars["id"])
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	})
	httpServer.Start()
	defer httpServer.Close()

	resourceCustomSystemEventDefinition := strings.ReplaceAll(resourceCustomSystemEventDefinitionTemplate, "{{PORT}}", strconv.Itoa(httpServer.GetPort()))

	resource.UnitTest(t, resource.TestCase{
		Providers: testCustomSystemEventProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: resourceCustomSystemEventDefinition,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testCustomSystemEventDefinition, "id"),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationFieldName, customSystemEventName),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationFieldEntityType, customSystemEventEntityType),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationFieldQuery, customSystemEventQuery),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationFieldTriggering, "true"),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationFieldDescription, customSystemEventDescription),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationFieldExpirationTime, strconv.Itoa(customSystemEventExpirationTime)),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationFieldEnabled, "true"),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationDownstreamIntegrationIds+".0", customSystemEventDownStringIntegrationId1),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationDownstreamIntegrationIds+".1", customSystemEventDownStringIntegrationId2),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs, "true"),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, CustomEventSpecificationRuleSeverity, customSystemEventRuleSeverity),
					resource.TestCheckResourceAttr(testCustomSystemEventDefinition, SystemRuleSpecificationSystemRuleID, customSystemEventRuleSystemRuleId),
				),
			},
		},
	})
}

func TestResourceCustomSystemEventDefinition(t *testing.T) {
	resource := CreateResourceCustomSystemEventSpecification()

	validateCustomSystemEventResourceSchema(resource.Schema, t)

	if resource.Create == nil {
		t.Fatal("Create function expected")
	}
	if resource.Update == nil {
		t.Fatal("Update function expected")
	}
	if resource.Read == nil {
		t.Fatal("Read function expected")
	}
	if resource.Delete == nil {
		t.Fatal("Delete function expected")
	}
}

func validateCustomSystemEventResourceSchema(schemaMap map[string]*schema.Schema, t *testing.T) {
	schemaAssert := testutils.NewTerraformSchemaAssert(schemaMap, t)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(CustomEventSpecificationFieldName)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(CustomEventSpecificationFieldEntityType)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeString(CustomEventSpecificationFieldQuery)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(CustomEventSpecificationFieldTriggering, false)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeString(CustomEventSpecificationFieldDescription)
	schemaAssert.AssertSchemaIsOptionalAndOfTypeInt(CustomEventSpecificationFieldExpirationTime)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(CustomEventSpecificationFieldEnabled, true)
	schemaAssert.AssertSChemaIsRequiredAndOfTypeListOfStrings(CustomEventSpecificationDownstreamIntegrationIds)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs, true)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(CustomEventSpecificationRuleSeverity)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(SystemRuleSpecificationSystemRuleID)
}

func TestShouldSuccessfullyReadCustomSystemEventFromInstanaAPIWhenBaseDataIsReturned(t *testing.T) {
	expectedModel := createBaseTestCustomSystemEventModel()
	testShouldSuccessfullyReadCustomSystemEventFromInstanaAPI(expectedModel, t)
}

func TestShouldSuccessfullyReadCustomSystemEventFromInstanaAPIWhenFullDataIsReturned(t *testing.T) {
	expectedModel := createTestCustomSystemEventModelWithFullDataSet()
	testShouldSuccessfullyReadCustomSystemEventFromInstanaAPI(expectedModel, t)
}

func testShouldSuccessfullyReadCustomSystemEventFromInstanaAPI(expectedModel restapi.CustomEventSpecification, t *testing.T) {
	resourceData := NewTestHelper(t).CreateEmptyCustomSystemEventSpecificationResourceData()
	resourceData.SetId(customSystemEventID)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().GetOne(gomock.Eq(customSystemEventID)).Return(expectedModel, nil).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Read(resourceData, mockInstanaAPI)

	if err != nil {
		t.Fatalf(testutils.ExpectedNoErrorButGotMessage, err)
	}
	verifyCustomSystemEventModelAppliedToResource(expectedModel, resourceData, t)
}

func TestShouldFailToReadCustomSystemEventFromInstanaAPIWhenIDIsMissing(t *testing.T) {
	resourceData := NewTestHelper(t).CreateEmptyCustomSystemEventSpecificationResourceData()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Read(resourceData, mockInstanaAPI)

	if err == nil || !strings.HasPrefix(err.Error(), "ID of custom event specification") {
		t.Fatal("Expected error to occur because of missing id")
	}
}

func TestShouldFailToReadCustomSystemEventFromInstanaAPIAndDeleteResourceWhenCustomEventDoesNotExist(t *testing.T) {
	resourceData := NewTestHelper(t).CreateEmptyCustomSystemEventSpecificationResourceData()
	resourceData.SetId(customSystemEventID)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().GetOne(gomock.Eq(customSystemEventID)).Return(restapi.CustomEventSpecification{}, restapi.ErrEntityNotFound).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Read(resourceData, mockInstanaAPI)

	if err != nil {
		t.Fatalf(testutils.ExpectedNoErrorButGotMessage, err)
	}
	if len(resourceData.Id()) > 0 {
		t.Fatal("Expected ID to be cleaned to destroy resource")
	}
}

func TestShouldFailToReadCustomSystemEventFromInstanaAPIAndReturnErrorWhenAPICallFails(t *testing.T) {
	resourceData := NewTestHelper(t).CreateEmptyCustomSystemEventSpecificationResourceData()
	resourceData.SetId(customSystemEventID)
	expectedError := errors.New("test")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().GetOne(gomock.Eq(customSystemEventID)).Return(restapi.CustomEventSpecification{}, expectedError).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Read(resourceData, mockInstanaAPI)

	if err == nil || err != expectedError {
		t.Fatal("Expected error should be returned")
	}
	if len(resourceData.Id()) == 0 {
		t.Fatal("Expected ID should still be set")
	}
}

func TestShouldFailToReadCustomSystemEventFromInstanaAPIWhenSeverityFromAPICannotBeMappedToSeverityOfTerraformState(t *testing.T) {
	expectedModel := createTestCustomSystemEventModelWithFullDataSet()
	expectedModel.Rules[0].Severity = 999
	resourceData := NewTestHelper(t).CreateEmptyCustomSystemEventSpecificationResourceData()
	resourceData.SetId(customSystemEventID)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().GetOne(gomock.Eq(customSystemEventID)).Return(expectedModel, nil).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Read(resourceData, mockInstanaAPI)

	if err == nil || !strings.Contains(err.Error(), "not a valid severity") {
		t.Fatal("Expected to get error that the provided severity is not valid")
	}
}

func TestShouldCreateCustomSystemEventThroughInstanaAPI(t *testing.T) {
	data := createFullTestCustomSystemEventData()
	resourceData := NewTestHelper(t).CreateCustomSystemEventSpecificationResourceData(data)
	expectedModel := createTestCustomSystemEventModelWithFullDataSet()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().Upsert(gomock.AssignableToTypeOf(restapi.CustomEventSpecification{})).Return(expectedModel, nil).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Create(resourceData, mockInstanaAPI)

	if err != nil {
		t.Fatalf(testutils.ExpectedNoErrorButGotMessage, err)
	}
	verifyCustomSystemEventModelAppliedToResource(expectedModel, resourceData, t)
}

func TestShouldReturnErrorWhenCreateCustomSystemEventFailsThroughInstanaAPI(t *testing.T) {
	data := createFullTestCustomSystemEventData()
	resourceData := NewTestHelper(t).CreateCustomSystemEventSpecificationResourceData(data)
	expectedError := errors.New("test")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().Upsert(gomock.AssignableToTypeOf(restapi.CustomEventSpecification{})).Return(restapi.CustomEventSpecification{}, expectedError).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Create(resourceData, mockInstanaAPI)

	if err == nil || expectedError != err {
		t.Fatal("Expected definned error to be returned")
	}
}

func TestShouldReturnErrorWhenCreateCustomSystemEventFailsBecauseOfInvalidSeverityConfiguredInTerraform(t *testing.T) {
	data := createFullTestCustomSystemEventData()
	data[CustomEventSpecificationRuleSeverity] = "invalid"
	resourceData := NewTestHelper(t).CreateCustomSystemEventSpecificationResourceData(data)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Create(resourceData, mockInstanaAPI)

	if err == nil || !strings.Contains(err.Error(), "not a valid severity") {
		t.Fatal("Expected to get error that the provided severity is not valid")
	}
}

func TestShouldReturnErrorWhenCreateCustomSystemEventFailsBecauseOfInvalidSeverityReturnedFromInstanaAPI(t *testing.T) {
	data := createFullTestCustomSystemEventData()
	resourceData := NewTestHelper(t).CreateCustomSystemEventSpecificationResourceData(data)
	expectedModel := createTestCustomSystemEventModelWithFullDataSet()
	expectedModel.Rules[0].Severity = 999

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().Upsert(gomock.AssignableToTypeOf(restapi.CustomEventSpecification{})).Return(expectedModel, nil).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Create(resourceData, mockInstanaAPI)

	if err == nil || !strings.Contains(err.Error(), "not a valid severity") {
		t.Fatal("Expected to get error that the provided severity is not valid")
	}
}

func TestShouldDeleteCustomSystemEventThroughInstanaAPI(t *testing.T) {
	id := "test-id"
	data := createFullTestCustomSystemEventData()
	resourceData := NewTestHelper(t).CreateCustomSystemEventSpecificationResourceData(data)
	resourceData.SetId(id)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().DeleteByID(gomock.Eq(id)).Return(nil).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Delete(resourceData, mockInstanaAPI)

	if err != nil {
		t.Fatalf(testutils.ExpectedNoErrorButGotMessage, err)
	}
	if len(resourceData.Id()) > 0 {
		t.Fatal("Expected ID to be cleaned to destroy resource")
	}
}

func TestShouldReturnErrorWhenDeleteCustomSystemEventFailsThroughInstanaAPI(t *testing.T) {
	id := "test-id"
	data := createFullTestCustomSystemEventData()
	resourceData := NewTestHelper(t).CreateCustomSystemEventSpecificationResourceData(data)
	resourceData.SetId(id)
	expectedError := errors.New("test")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCustomEventAPI := mocks.NewMockCustomEventSpecificationResource(ctrl)
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	mockInstanaAPI.EXPECT().CustomEventSpecifications().Return(mockCustomEventAPI).Times(1)
	mockCustomEventAPI.EXPECT().DeleteByID(gomock.Eq(id)).Return(expectedError).Times(1)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Delete(resourceData, mockInstanaAPI)

	if err == nil || err != expectedError {
		t.Fatal("Expected error to be returned")
	}
	if len(resourceData.Id()) == 0 {
		t.Fatal("Expected ID not to be cleaned to avoid resource is destroy")
	}
}

func TestShouldFailToDeleteCustomSystemEventWhenInvalidSeverityIsConfiguredInTerraform(t *testing.T) {
	id := "test-id"
	data := createFullTestCustomSystemEventData()
	data[CustomEventSpecificationRuleSeverity] = "invalid"
	resourceData := NewTestHelper(t).CreateCustomSystemEventSpecificationResourceData(data)
	resourceData.SetId(id)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstanaAPI := mocks.NewMockInstanaAPI(ctrl)

	resource := CreateResourceCustomSystemEventSpecification()
	err := resource.Delete(resourceData, mockInstanaAPI)

	if err == nil || !strings.Contains(err.Error(), "not a valid severity") {
		t.Fatal("Expected to get error that the provided severity is not valid")
	}
}

func verifyCustomSystemEventModelAppliedToResource(model restapi.CustomEventSpecification, resourceData *schema.ResourceData, t *testing.T) {
	if model.ID != resourceData.Id() {
		t.Fatal("Expected ID to be identical")
	}
	if model.Name != resourceData.Get(CustomEventSpecificationFieldName).(string) {
		t.Fatal("Expected Name to be identical")
	}
	if model.EntityType != resourceData.Get(CustomEventSpecificationFieldEntityType).(string) {
		t.Fatal("Expected EntityType to be identical")
	}
	if model.Query != nil {
		if *model.Query != resourceData.Get(CustomEventSpecificationFieldQuery).(string) {
			t.Fatal("Expected Query to be identical")
		}
	} else {
		if _, ok := resourceData.GetOk(CustomEventSpecificationFieldQuery); ok {
			t.Fatal("Expected Query not to be defined")
		}
	}
	if model.Triggering != resourceData.Get(CustomEventSpecificationFieldTriggering).(bool) {
		t.Fatal("Expected Triggering to be identical")
	}
	if model.Description != nil {
		if *model.Description != resourceData.Get(CustomEventSpecificationFieldDescription).(string) {
			t.Fatal("Expected Description to be identical")
		}
	} else {
		if _, ok := resourceData.GetOk(CustomEventSpecificationFieldDescription); ok {
			t.Fatal("Expected Description not to be defined")
		}
	}
	if model.ExpirationTime != nil {
		if *model.ExpirationTime != resourceData.Get(CustomEventSpecificationFieldExpirationTime).(int) {
			t.Fatal("Expected Expiration Time to be identical")
		}
	} else {
		if _, ok := resourceData.GetOk(CustomEventSpecificationFieldExpirationTime); ok {
			t.Fatal("Expected Expiration Time not to be defined")
		}
	}
	if model.Enabled != resourceData.Get(CustomEventSpecificationFieldEnabled).(bool) {
		t.Fatal("Expected Enabled to be identical")
	}

	if model.Downstream != nil {
		if !cmp.Equal(model.Downstream.IntegrationIds, ReadStringArrayParameterFromResource(resourceData, CustomEventSpecificationDownstreamIntegrationIds)) {
			t.Fatal("Expected Integration IDs to be identical")
		}
		if model.Downstream.BroadcastToAllAlertingConfigs != resourceData.Get(CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs).(bool) {
			t.Fatal("Expected Broadcast to All Alert Configs to be identical")
		}
	} else {
		if _, ok := resourceData.GetOk(CustomEventSpecificationDownstreamIntegrationIds); ok {
			t.Fatal("Expected Integration IDs not to be defined")
		}
		if true != resourceData.Get(CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs) {
			t.Fatalf("Expected Broadcast to All Alert Configs to have the default value set")
		}
	}

	convertedSeverity, err := ConvertSeverityFromInstanaAPIToTerraformRepresentation(model.Rules[0].Severity)
	if err != nil {
		t.Fatalf(testutils.ExpectedNoErrorButGotMessage, err)
	}
	if convertedSeverity != resourceData.Get(CustomEventSpecificationRuleSeverity).(string) {
		t.Fatal("Expected Severity to be identical")
	}

	if model.Rules[0].SystemRuleID != resourceData.Get(SystemRuleSpecificationSystemRuleID).(string) {
		t.Fatal("Expected System Rule ID to be identical")
	}
}

func createTestCustomSystemEventModelWithFullDataSet() restapi.CustomEventSpecification {
	description := customSystemEventDescription
	expirationTime := customSystemEventExpirationTime
	query := customSystemEventQuery

	data := createBaseTestCustomSystemEventModel()
	data.Query = &query
	data.Description = &description
	data.ExpirationTime = &expirationTime
	data.Downstream = &restapi.EventSpecificationDownstream{
		IntegrationIds:                []string{customSystemEventDownStringIntegrationId1, customSystemEventDownStringIntegrationId2},
		BroadcastToAllAlertingConfigs: true,
	}
	return data
}

func createBaseTestCustomSystemEventModel() restapi.CustomEventSpecification {
	return restapi.CustomEventSpecification{
		ID:         customSystemEventID,
		Name:       customSystemEventName,
		EntityType: customSystemEventEntityType,
		Triggering: false,
		Enabled:    true,
		Rules: []restapi.RuleSpecification{
			restapi.NewSystemRuleSpecification(customSystemEventRuleSystemRuleId, restapi.SeverityWarning.GetAPIRepresentation()),
		},
	}
}

func createFullTestCustomSystemEventData() map[string]interface{} {
	data := make(map[string]interface{})
	data[CustomEventSpecificationFieldName] = customSystemEventName
	data[CustomEventSpecificationFieldEntityType] = customSystemEventEntityType
	data[CustomEventSpecificationFieldQuery] = customSystemEventQuery
	data[CustomEventSpecificationFieldTriggering] = "true"
	data[CustomEventSpecificationFieldDescription] = customSystemEventDescription
	data[CustomEventSpecificationFieldExpirationTime] = customSystemEventExpirationTime
	data[CustomEventSpecificationFieldEnabled] = "true"
	data[CustomEventSpecificationDownstreamIntegrationIds] = []string{customSystemEventDownStringIntegrationId1, customSystemEventDownStringIntegrationId2}
	data[CustomEventSpecificationDownstreamBroadcastToAllAlertingConfigs] = "true"
	data[CustomEventSpecificationRuleSeverity] = customSystemEventRuleSeverity
	data[SystemRuleSpecificationSystemRuleID] = customSystemEventRuleSystemRuleId
	return data
}
