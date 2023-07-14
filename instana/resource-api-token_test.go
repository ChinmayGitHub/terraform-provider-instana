package instana_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"

	. "github.com/gessnerfl/terraform-provider-instana/instana"
	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	"github.com/gessnerfl/terraform-provider-instana/testutils"
)

const resourceAPITokenDefinitionTemplate = `
resource "instana_api_token" "example" {
  name = "name %d"
  can_configure_service_mapping = true
  can_configure_eum_applications = true
  can_configure_mobile_app_monitoring = true
  can_configure_users = true
  can_install_new_agents = true
  can_see_usage_information = true
  can_configure_integrations = true
  can_see_on_premise_license_information = true
  can_configure_custom_alerts = true
  can_configure_api_tokens = true
  can_configure_agent_run_mode = true
  can_view_audit_log = true
  can_configure_agents = true
  can_configure_authentication_methods = true
  can_configure_applications = true
  can_configure_teams = true
  can_configure_releases = true
  can_configure_log_management = true
  can_create_public_custom_dashboards = true
  can_view_logs = true
  can_view_trace_details = true
  can_configure_session_settings = true
  can_configure_service_level_indicators = true
  can_configure_global_alert_payload = true
  can_configure_global_alert_configs = true
  can_view_account_and_billing_information = true
  can_edit_all_accessible_custom_dashboards = true
}
`

const (
	apiTokenApiPath             = restapi.APITokensResourcePath + "/{internal-id}"
	testAPITokenDefinition      = "instana_api_token.example"
	apiTokenID                  = "api-token-id"
	apiTokenNameFieldValue      = resourceName
	apiTokenFullNameFieldValue  = resourceFullName
	apiTokenAccessGrantingToken = "api-token-access-granting-token"
	apiTokenInternalID          = "api-token-internal-id"
)

var apiTokenPermissionFields = []string{
	APITokenFieldCanConfigureServiceMapping,
	APITokenFieldCanConfigureEumApplications,
	APITokenFieldCanConfigureMobileAppMonitoring,
	APITokenFieldCanConfigureUsers,
	APITokenFieldCanInstallNewAgents,
	APITokenFieldCanSeeUsageInformation,
	APITokenFieldCanConfigureIntegrations,
	APITokenFieldCanSeeOnPremiseLicenseInformation,
	APITokenFieldCanConfigureCustomAlerts,
	APITokenFieldCanConfigureAPITokens,
	APITokenFieldCanConfigureAgentRunMode,
	APITokenFieldCanViewAuditLog,
	APITokenFieldCanConfigureAgents,
	APITokenFieldCanConfigureAuthenticationMethods,
	APITokenFieldCanConfigureApplications,
	APITokenFieldCanConfigureTeams,
	APITokenFieldCanConfigureReleases,
	APITokenFieldCanConfigureLogManagement,
	APITokenFieldCanCreatePublicCustomDashboards,
	APITokenFieldCanViewLogs,
	APITokenFieldCanViewTraceDetails,
	APITokenFieldCanConfigureSessionSettings,
	APITokenFieldCanConfigureServiceLevelIndicators,
	APITokenFieldCanConfigureGlobalAlertPayload,
	APITokenFieldCanConfigureGlobalAlertConfigs,
	APITokenFieldCanViewAccountAndBillingInformation,
	APITokenFieldCanEditAllAccessibleCustomDashboards,
}

func TestCRUDOfAPITokenResourceWithMockServer(t *testing.T) {
	id := RandomID()
	accessGrantingToken := RandomID()
	internalID := RandomID()
	httpServer := testutils.NewTestHTTPServer()
	httpServer.AddRoute(http.MethodPost, restapi.APITokensResourcePath, func(w http.ResponseWriter, r *http.Request) {
		apiToken := &restapi.APIToken{}
		err := json.NewDecoder(r.Body).Decode(apiToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			r.Write(bytes.NewBufferString("Failed to get request"))
		} else {
			apiToken.ID = id
			apiToken.AccessGrantingToken = accessGrantingToken
			apiToken.InternalID = internalID
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(apiToken)
		}
	})
	httpServer.AddRoute(http.MethodPut, apiTokenApiPath, testutils.EchoHandlerFunc)
	httpServer.AddRoute(http.MethodDelete, apiTokenApiPath, testutils.EchoHandlerFunc)
	httpServer.AddRoute(http.MethodGet, apiTokenApiPath, func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		modCount := httpServer.GetCallCount(http.MethodPut, restapi.APITokensResourcePath+"/"+internalID)
		json := fmt.Sprintf(`
		{
			"id" : "%s",
			"accessGrantingToken": "%s",
			"internalId" : "%s",
			"name" : "name %d",
			"canConfigureServiceMapping" : true,
			"canConfigureEumApplications" : true,
			"canConfigureMobileAppMonitoring" : true,
			"canConfigureUsers" : true,
			"canInstallNewAgents" : true,
			"canSeeUsageInformation" : true,
			"canConfigureIntegrations" : true,
			"canSeeOnPremLicenseInformation" : true,
			"canConfigureCustomAlerts" : true,
			"canConfigureApiTokens" : true,
			"canConfigureAgentRunMode" : true,
			"canViewAuditLog" : true,
			"canConfigureAgents" : true,
			"canConfigureAuthenticationMethods" : true,
			"canConfigureApplications" : true,
			"canConfigureTeams" : true,
			"canConfigureReleases" : true,
			"canConfigureLogManagement" : true,
			"canCreatePublicCustomDashboards" : true,
			"canViewLogs" : true,
			"canViewTraceDetails" : true,
			"canConfigureSessionSettings" : true,
			"canConfigureServiceLevelIndicators" : true,
			"canConfigureGlobalAlertPayload" : true,
			"canConfigureGlobalAlertConfigs" : true,
			"canViewAccountAndBillingInformation" : true,
			"canEditAllAccessibleCustomDashboards" : true
		}
		`, id, accessGrantingToken, vars["internal-id"], modCount)
		w.Header().Set(contentType, r.Header.Get(contentType))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	})
	httpServer.Start()
	defer httpServer.Close()

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testProviderFactory,
		Steps: []resource.TestStep{
			createAPITokenConfigResourceTestStep(httpServer.GetPort(), 0, id, accessGrantingToken, internalID),
			testStepImportWithCustomID(testAPITokenDefinition, internalID),
			createAPITokenConfigResourceTestStep(httpServer.GetPort(), 1, id, accessGrantingToken, internalID),
			testStepImportWithCustomID(testAPITokenDefinition, internalID),
		},
	})
}

func createAPITokenConfigResourceTestStep(httpPort int64, iteration int, id string, accessGrantingToken string, internalID string) resource.TestStep {
	return resource.TestStep{
		Config: appendProviderConfig(fmt.Sprintf(resourceAPITokenDefinitionTemplate, iteration), httpPort),
		Check: resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(testAPITokenDefinition, "id", id),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldAccessGrantingToken, accessGrantingToken),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldInternalID, internalID),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldName, formatResourceName(iteration)),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldFullName, formatResourceFullName(iteration)),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureServiceMapping, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureEumApplications, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureMobileAppMonitoring, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureUsers, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanInstallNewAgents, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanSeeUsageInformation, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureIntegrations, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanSeeOnPremiseLicenseInformation, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureCustomAlerts, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureAPITokens, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureAgentRunMode, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanViewAuditLog, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureAgents, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureAuthenticationMethods, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureApplications, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureTeams, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureReleases, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureLogManagement, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanCreatePublicCustomDashboards, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanViewLogs, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanViewTraceDetails, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureSessionSettings, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureServiceLevelIndicators, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureGlobalAlertPayload, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanConfigureGlobalAlertConfigs, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanViewAccountAndBillingInformation, trueAsString),
			resource.TestCheckResourceAttr(testAPITokenDefinition, APITokenFieldCanEditAllAccessibleCustomDashboards, trueAsString),
		),
	}
}

func TestAPITokenSchemaDefinitionIsValid(t *testing.T) {
	schema := NewAPITokenResourceHandle().MetaData().Schema

	schemaAssert := testutils.NewTerraformSchemaAssert(schema, t)
	schemaAssert.AssertSchemaIsComputedAndOfTypeString(APITokenFieldAccessGrantingToken)
	schemaAssert.AssertSchemaIsComputedAndOfTypeString(APITokenFieldInternalID)
	schemaAssert.AssertSchemaIsRequiredAndOfTypeString(APITokenFieldName)
	schemaAssert.AssertSchemaIsComputedAndOfTypeString(APITokenFieldFullName)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureServiceMapping, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureEumApplications, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureMobileAppMonitoring, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureUsers, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanInstallNewAgents, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanSeeUsageInformation, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureIntegrations, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanSeeOnPremiseLicenseInformation, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureCustomAlerts, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureAPITokens, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureAgentRunMode, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanViewAuditLog, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureAgents, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureAuthenticationMethods, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureApplications, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureTeams, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureReleases, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureLogManagement, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanCreatePublicCustomDashboards, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanViewLogs, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanViewTraceDetails, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureSessionSettings, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureServiceLevelIndicators, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureGlobalAlertPayload, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanConfigureGlobalAlertConfigs, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanViewAccountAndBillingInformation, false)
	schemaAssert.AssertSchemaIsOfTypeBooleanWithDefault(APITokenFieldCanEditAllAccessibleCustomDashboards, false)
}

func TestAPITokenResourceShouldHaveSchemaVersionZero(t *testing.T) {
	require.Equal(t, 0, NewAPITokenResourceHandle().MetaData().SchemaVersion)
}

func TestAPITokenResourceShouldHaveNoStateMigrators(t *testing.T) {
	require.Equal(t, 0, len(NewAPITokenResourceHandle().StateUpgraders()))
}

func TestShouldReturnCorrectResourceNameForUserroleResource(t *testing.T) {
	name := NewAPITokenResourceHandle().MetaData().ResourceName

	require.Equal(t, name, "instana_api_token")
}

func TestShouldSetCalculateAccessGrantingTokenAndInternal(t *testing.T) {
	testHelper := NewTestHelper(t)
	sut := NewAPITokenResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	sut.SetComputedFields(resourceData)

	require.NotEmpty(t, resourceData.Get(APITokenFieldInternalID))
	require.NotEmpty(t, resourceData.Get(APITokenFieldAccessGrantingToken))
}

func TestShouldUpdateBasicFieldsOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	testHelper := NewTestHelper(t)
	sut := NewAPITokenResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenFullNameFieldValue,
		InternalID:          apiTokenInternalID,
	}

	err := sut.UpdateState(resourceData, &apiToken, testHelper.ResourceFormatter())

	require.Nil(t, err)
	require.Equal(t, apiTokenID, resourceData.Id())
	require.Equal(t, apiTokenAccessGrantingToken, resourceData.Get(APITokenFieldAccessGrantingToken))
	require.Equal(t, apiTokenInternalID, resourceData.Get(APITokenFieldInternalID))
	require.Equal(t, apiTokenNameFieldValue, resourceData.Get(APITokenFieldName))
	require.Equal(t, apiTokenFullNameFieldValue, resourceData.Get(APITokenFieldFullName))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureServiceMapping).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureEumApplications).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureMobileAppMonitoring).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureUsers).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanInstallNewAgents).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanSeeUsageInformation).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureIntegrations).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanSeeOnPremiseLicenseInformation).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureCustomAlerts).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureAPITokens).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureAgentRunMode).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanViewAuditLog).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureAgents).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureAuthenticationMethods).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureApplications).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureTeams).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureReleases).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureLogManagement).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanCreatePublicCustomDashboards).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanViewLogs).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanViewTraceDetails).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureSessionSettings).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureServiceLevelIndicators).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureGlobalAlertPayload).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanConfigureGlobalAlertConfigs).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanViewAccountAndBillingInformation).(bool))
	require.False(t, resourceData.Get(APITokenFieldCanEditAllAccessibleCustomDashboards).(bool))
}

func TestShouldUpdateCanConfigureServiceMappingPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                         apiTokenID,
		InternalID:                 apiTokenInternalID,
		AccessGrantingToken:        apiTokenAccessGrantingToken,
		Name:                       apiTokenNameFieldValue,
		CanConfigureServiceMapping: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureServiceMapping)
}

func TestShouldUpdateCanConfigureEumApplicationsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                          apiTokenID,
		InternalID:                  apiTokenInternalID,
		AccessGrantingToken:         apiTokenAccessGrantingToken,
		Name:                        apiTokenNameFieldValue,
		CanConfigureEumApplications: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureEumApplications)
}

func TestShouldUpdateCanConfigureMobileAppMonitoringPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                              apiTokenID,
		InternalID:                      apiTokenInternalID,
		AccessGrantingToken:             apiTokenAccessGrantingToken,
		Name:                            apiTokenNameFieldValue,
		CanConfigureMobileAppMonitoring: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureMobileAppMonitoring)
}

func TestShouldUpdateCanConfigureUsersPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		InternalID:          apiTokenInternalID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenNameFieldValue,
		CanConfigureUsers:   true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureUsers)
}

func TestShouldUpdateCanInstallNewAgentsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		InternalID:          apiTokenInternalID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenNameFieldValue,
		CanInstallNewAgents: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanInstallNewAgents)
}

func TestShouldUpdateCanSeeUsageInformationPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                     apiTokenID,
		InternalID:             apiTokenInternalID,
		AccessGrantingToken:    apiTokenAccessGrantingToken,
		Name:                   apiTokenNameFieldValue,
		CanSeeUsageInformation: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanSeeUsageInformation)
}

func TestShouldUpdateCanConfigureIntegrationsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                       apiTokenID,
		InternalID:               apiTokenInternalID,
		AccessGrantingToken:      apiTokenAccessGrantingToken,
		Name:                     apiTokenNameFieldValue,
		CanConfigureIntegrations: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureIntegrations)
}

func TestShouldUpdateCanSeeOnPremiseLicenseInformationPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                                apiTokenID,
		InternalID:                        apiTokenInternalID,
		AccessGrantingToken:               apiTokenAccessGrantingToken,
		Name:                              apiTokenNameFieldValue,
		CanSeeOnPremiseLicenseInformation: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanSeeOnPremiseLicenseInformation)
}

func TestShouldUpdateCanConfigureCustomAlertsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                       apiTokenID,
		InternalID:               apiTokenInternalID,
		AccessGrantingToken:      apiTokenAccessGrantingToken,
		Name:                     apiTokenNameFieldValue,
		CanConfigureCustomAlerts: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureCustomAlerts)
}

func TestShouldUpdateCanConfigureAPITokensPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                    apiTokenID,
		InternalID:            apiTokenInternalID,
		AccessGrantingToken:   apiTokenAccessGrantingToken,
		Name:                  apiTokenNameFieldValue,
		CanConfigureAPITokens: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureAPITokens)
}

func TestShouldUpdateCanConfigureAgentRunModePermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                       apiTokenID,
		InternalID:               apiTokenInternalID,
		AccessGrantingToken:      apiTokenAccessGrantingToken,
		Name:                     apiTokenNameFieldValue,
		CanConfigureAgentRunMode: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureAgentRunMode)
}

func TestShouldUpdateCanViewAuditLogPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		InternalID:          apiTokenInternalID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenNameFieldValue,
		CanViewAuditLog:     true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanViewAuditLog)
}

func TestShouldUpdateCanConfigureAgentsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		InternalID:          apiTokenInternalID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenNameFieldValue,
		CanConfigureAgents:  true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureAgents)
}

func TestShouldUpdateCanConfigureAuthenticationMethodsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                                apiTokenID,
		InternalID:                        apiTokenInternalID,
		AccessGrantingToken:               apiTokenAccessGrantingToken,
		Name:                              apiTokenNameFieldValue,
		CanConfigureAuthenticationMethods: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureAuthenticationMethods)
}

func TestShouldUpdateCanConfigureApplicationsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                       apiTokenID,
		InternalID:               apiTokenInternalID,
		AccessGrantingToken:      apiTokenAccessGrantingToken,
		Name:                     apiTokenNameFieldValue,
		CanConfigureApplications: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureApplications)
}

func TestShouldUpdateCanConfigureTeamsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		InternalID:          apiTokenInternalID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenNameFieldValue,
		CanConfigureTeams:   true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureTeams)
}

func TestShouldUpdateCanConfigureReleasesPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                   apiTokenID,
		InternalID:           apiTokenInternalID,
		AccessGrantingToken:  apiTokenAccessGrantingToken,
		Name:                 apiTokenNameFieldValue,
		CanConfigureReleases: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureReleases)
}

func TestShouldUpdateCanConfigureLogManagementPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                        apiTokenID,
		InternalID:                apiTokenInternalID,
		AccessGrantingToken:       apiTokenAccessGrantingToken,
		Name:                      apiTokenNameFieldValue,
		CanConfigureLogManagement: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureLogManagement)
}

func TestShouldUpdateCanCreatePublicCustomDashboardsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                              apiTokenID,
		InternalID:                      apiTokenInternalID,
		AccessGrantingToken:             apiTokenAccessGrantingToken,
		Name:                            apiTokenNameFieldValue,
		CanCreatePublicCustomDashboards: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanCreatePublicCustomDashboards)
}

func TestShouldUpdateCanViewLogsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		InternalID:          apiTokenInternalID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenNameFieldValue,
		CanViewLogs:         true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanViewLogs)
}

func TestShouldUpdateCanViewTraceDetailsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                  apiTokenID,
		InternalID:          apiTokenInternalID,
		AccessGrantingToken: apiTokenAccessGrantingToken,
		Name:                apiTokenNameFieldValue,
		CanViewTraceDetails: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanViewTraceDetails)
}

func TestShouldUpdateCanConfigureSessionSettingsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                          apiTokenID,
		InternalID:                  apiTokenInternalID,
		AccessGrantingToken:         apiTokenAccessGrantingToken,
		Name:                        apiTokenNameFieldValue,
		CanConfigureSessionSettings: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureSessionSettings)
}

func TestShouldUpdateCanConfigureServiceLevelIndicatorsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                                 apiTokenID,
		InternalID:                         apiTokenInternalID,
		AccessGrantingToken:                apiTokenAccessGrantingToken,
		Name:                               apiTokenNameFieldValue,
		CanConfigureServiceLevelIndicators: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureServiceLevelIndicators)
}

func TestShouldUpdateCanConfigureGlobalAlertPayloadPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                             apiTokenID,
		Name:                           apiTokenNameFieldValue,
		CanConfigureGlobalAlertPayload: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureGlobalAlertPayload)
}

func TestShouldUpdateCanConfigureGlobalAlertConfigsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                             apiTokenID,
		InternalID:                     apiTokenInternalID,
		AccessGrantingToken:            apiTokenAccessGrantingToken,
		Name:                           apiTokenNameFieldValue,
		CanConfigureGlobalAlertConfigs: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanConfigureGlobalAlertConfigs)
}

func TestShouldUpdateCanViewAccountAndBillingInformationPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                                  apiTokenID,
		InternalID:                          apiTokenInternalID,
		AccessGrantingToken:                 apiTokenAccessGrantingToken,
		Name:                                apiTokenNameFieldValue,
		CanViewAccountAndBillingInformation: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanViewAccountAndBillingInformation)
}

func TestShouldUpdateCanEditAllAccessibleCustomDashboardsPermissionOfTerraformResourceStateFromModelForAPIToken(t *testing.T) {
	apiToken := restapi.APIToken{
		ID:                                   apiTokenID,
		InternalID:                           apiTokenInternalID,
		AccessGrantingToken:                  apiTokenAccessGrantingToken,
		Name:                                 apiTokenNameFieldValue,
		CanEditAllAccessibleCustomDashboards: true,
	}

	testSingleAPITokenPermissionSet(t, apiToken, APITokenFieldCanEditAllAccessibleCustomDashboards)
}

func testSingleAPITokenPermissionSet(t *testing.T, apiToken restapi.APIToken, expectedPermissionField string) {
	testHelper := NewTestHelper(t)
	sut := NewAPITokenResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(sut)

	err := sut.UpdateState(resourceData, &apiToken, testHelper.ResourceFormatter())

	require.Nil(t, err)
	require.True(t, resourceData.Get(expectedPermissionField).(bool))
	for _, permissionField := range apiTokenPermissionFields {
		if permissionField != expectedPermissionField {
			require.False(t, resourceData.Get(permissionField).(bool))
		}
	}
}

func TestShouldConvertStateOfAPITokenTerraformResourceToDataModel(t *testing.T) {
	testHelper := NewTestHelper(t)
	resourceHandle := NewAPITokenResourceHandle()

	resourceData := testHelper.CreateEmptyResourceDataForResourceHandle(resourceHandle)
	resourceData.SetId(apiTokenID)
	resourceData.Set(APITokenFieldAccessGrantingToken, apiTokenAccessGrantingToken)
	resourceData.Set(APITokenFieldInternalID, apiTokenInternalID)
	resourceData.Set(APITokenFieldName, apiTokenNameFieldValue)
	resourceData.Set(APITokenFieldFullName, apiTokenFullNameFieldValue)
	resourceData.Set(APITokenFieldCanConfigureServiceMapping, true)
	resourceData.Set(APITokenFieldCanConfigureEumApplications, true)
	resourceData.Set(APITokenFieldCanConfigureMobileAppMonitoring, true)
	resourceData.Set(APITokenFieldCanConfigureUsers, true)
	resourceData.Set(APITokenFieldCanInstallNewAgents, true)
	resourceData.Set(APITokenFieldCanSeeUsageInformation, true)
	resourceData.Set(APITokenFieldCanConfigureIntegrations, true)
	resourceData.Set(APITokenFieldCanSeeOnPremiseLicenseInformation, true)
	resourceData.Set(APITokenFieldCanConfigureCustomAlerts, true)
	resourceData.Set(APITokenFieldCanConfigureAPITokens, true)
	resourceData.Set(APITokenFieldCanConfigureAgentRunMode, true)
	resourceData.Set(APITokenFieldCanViewAuditLog, true)
	resourceData.Set(APITokenFieldCanConfigureAgents, true)
	resourceData.Set(APITokenFieldCanConfigureAuthenticationMethods, true)
	resourceData.Set(APITokenFieldCanConfigureApplications, true)
	resourceData.Set(APITokenFieldCanConfigureTeams, true)
	resourceData.Set(APITokenFieldCanConfigureReleases, true)
	resourceData.Set(APITokenFieldCanConfigureLogManagement, true)
	resourceData.Set(APITokenFieldCanCreatePublicCustomDashboards, true)
	resourceData.Set(APITokenFieldCanViewLogs, true)
	resourceData.Set(APITokenFieldCanViewTraceDetails, true)
	resourceData.Set(APITokenFieldCanConfigureSessionSettings, true)
	resourceData.Set(APITokenFieldCanConfigureServiceLevelIndicators, true)
	resourceData.Set(APITokenFieldCanConfigureGlobalAlertPayload, true)
	resourceData.Set(APITokenFieldCanConfigureGlobalAlertConfigs, true)
	resourceData.Set(APITokenFieldCanViewAccountAndBillingInformation, true)
	resourceData.Set(APITokenFieldCanEditAllAccessibleCustomDashboards, true)

	model, err := resourceHandle.MapStateToDataObject(resourceData, testHelper.ResourceFormatter())

	require.Nil(t, err)
	require.IsType(t, &restapi.APIToken{}, model, "Model should be an alerting channel")
	require.Equal(t, apiTokenID, model.(*restapi.APIToken).ID)
	require.Equal(t, apiTokenAccessGrantingToken, model.(*restapi.APIToken).AccessGrantingToken)
	require.Equal(t, apiTokenInternalID, model.GetIDForResourcePath())
	require.Equal(t, apiTokenInternalID, model.(*restapi.APIToken).InternalID)
	require.Equal(t, apiTokenFullNameFieldValue, model.(*restapi.APIToken).Name)
	require.True(t, model.(*restapi.APIToken).CanConfigureServiceMapping)
	require.True(t, model.(*restapi.APIToken).CanConfigureEumApplications)
	require.True(t, model.(*restapi.APIToken).CanConfigureMobileAppMonitoring)
	require.True(t, model.(*restapi.APIToken).CanConfigureUsers)
	require.True(t, model.(*restapi.APIToken).CanInstallNewAgents)
	require.True(t, model.(*restapi.APIToken).CanSeeUsageInformation)
	require.True(t, model.(*restapi.APIToken).CanConfigureIntegrations)
	require.True(t, model.(*restapi.APIToken).CanSeeOnPremiseLicenseInformation)
	require.True(t, model.(*restapi.APIToken).CanConfigureCustomAlerts)
	require.True(t, model.(*restapi.APIToken).CanConfigureAPITokens)
	require.True(t, model.(*restapi.APIToken).CanConfigureAgentRunMode)
	require.True(t, model.(*restapi.APIToken).CanViewAuditLog)
	require.True(t, model.(*restapi.APIToken).CanConfigureAgents)
	require.True(t, model.(*restapi.APIToken).CanConfigureAuthenticationMethods)
	require.True(t, model.(*restapi.APIToken).CanConfigureApplications)
	require.True(t, model.(*restapi.APIToken).CanConfigureTeams)
	require.True(t, model.(*restapi.APIToken).CanConfigureReleases)
	require.True(t, model.(*restapi.APIToken).CanConfigureLogManagement)
	require.True(t, model.(*restapi.APIToken).CanCreatePublicCustomDashboards)
	require.True(t, model.(*restapi.APIToken).CanViewLogs)
	require.True(t, model.(*restapi.APIToken).CanViewTraceDetails)
	require.True(t, model.(*restapi.APIToken).CanConfigureSessionSettings)
	require.True(t, model.(*restapi.APIToken).CanConfigureServiceLevelIndicators)
	require.True(t, model.(*restapi.APIToken).CanConfigureGlobalAlertPayload)
	require.True(t, model.(*restapi.APIToken).CanConfigureGlobalAlertConfigs)
	require.True(t, model.(*restapi.APIToken).CanViewAccountAndBillingInformation)
	require.True(t, model.(*restapi.APIToken).CanEditAllAccessibleCustomDashboards)
}
