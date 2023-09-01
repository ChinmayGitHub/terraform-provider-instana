package restapi

import (
	"strings"
)

// AlertsResourcePath path to Alerts resource of Instana RESTful API
const AlertsResourcePath = EventSettingsBasePath + "/alerts"

// AlertEventType type definition of EventTypes of an Instana Alert
type AlertEventType string

// Equals checks if the alert event type is equal to the provided alert event type. It compares the string representation of both case insensitive
func (t AlertEventType) Equals(other AlertEventType) bool {
	return strings.EqualFold(string(t), string(other))
}

const (
	//IncidentAlertEventType constant value for alert event type incident
	IncidentAlertEventType = AlertEventType("incident")
	//CriticalAlertEventType constant value for alert event type critical
	CriticalAlertEventType = AlertEventType("critical")
	//WarningAlertEventType constant value for alert event type warning
	WarningAlertEventType = AlertEventType("warning")
	//ChangeAlertEventType constant value for alert event type change
	ChangeAlertEventType = AlertEventType("change")
	//OnlineAlertEventType constant value for alert event type online
	OnlineAlertEventType = AlertEventType("online")
	//OfflineAlertEventType constant value for alert event type offline
	OfflineAlertEventType = AlertEventType("offline")
	//NoneAlertEventType constant value for alert event type none
	NoneAlertEventType = AlertEventType("none")
	//AgentMonitoringIssueEventType constant value for alert event type none
	AgentMonitoringIssueEventType = AlertEventType("agent_monitoring_issue")
)

// SupportedAlertEventTypes list of supported alert event types of Instana API
var SupportedAlertEventTypes = []AlertEventType{
	IncidentAlertEventType,
	CriticalAlertEventType,
	WarningAlertEventType,
	ChangeAlertEventType,
	OnlineAlertEventType,
	OfflineAlertEventType,
	NoneAlertEventType,
	AgentMonitoringIssueEventType,
}

// IsSupportedAlertEventType checks if the given alert type is supported by Instana API
func IsSupportedAlertEventType(t AlertEventType) bool {
	for _, supported := range SupportedAlertEventTypes {
		if supported.Equals(t) {
			return true
		}
	}
	return false
}

// EventFilteringConfiguration type definiton of an EventFilteringConfiguration of a AlertingConfiguration of the Instana ReST AOI
type EventFilteringConfiguration struct {
	Query      *string          `json:"query"`
	RuleIDs    []string         `json:"ruleIds"`
	EventTypes []AlertEventType `json:"eventTypes"`
}

// AlertingConfiguration type definition of an Alertinng Configruation in Instana REST API
type AlertingConfiguration struct {
	ID                          string                      `json:"id"`
	AlertName                   string                      `json:"alertName"`
	IntegrationIDs              []string                    `json:"integrationIds"`
	EventFilteringConfiguration EventFilteringConfiguration `json:"eventFilteringConfiguration"`
}

// GetIDForResourcePath implementation of the interface InstanaDataObject
func (c *AlertingConfiguration) GetIDForResourcePath() string {
	return c.ID
}
