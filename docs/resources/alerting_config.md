# Alerting Configuration

Management of alert configurations. Alert configurations define how either event types or 
event (aka rules) are reported to integrated services (Alerting Channels).

API Documentation: <https://instana.github.io/openapi/#operation/putAlert>

The ID of the resource which is also used as unique identifier in Instana is auto generated!

---
**Note:**

Instana web UI provides the option to select for applications or a dynamic filter query. This is a UI feature only. To setup
altering configuration for applications you need to express this by a dynamic filter query:

`entity.application.id:\"my-application-perspective-id\""`

---

## Example Usage

### Rule ids

```hcl
resource "instana_alerting_config" "example" {
  alert_name            = "name"
  integration_ids       = [ "alerting-channel-id1", "alerting-channel-id2" ]
  event_filter_query    = "query"
  event_filter_rule_ids = [ "rule-1", "rule-2" ]
}
``` 

### Event types

```hcl
resource "instana_alerting_config" "example" {
  alert_name               = "name"
  integration_ids          = [ "alerting-channel-id1", "alerting-channel-id2" ]
  event_filter_query       = "query"
  event_filter_event_types = [ "incident", "critical" ]
}
``` 

## Argument Reference

* `alert_name` - Required - the name of the alerting configuration
* `integration_ids` - Optional - the list of target alerting channel ids
* `event_filter_query` - Optional - a dynamic focus query to restrict the alert configuration to a sub set of entities
* `event_filter_rule_ids` - Optional - list of rule IDs which are included by the alerting config.
* `event_filter_event_types` - Optional - list of event types which are included by the alerting config.
Allowed values: `incident`, `critical`, `warning`, `change`, `online`, `offline`, `agent_monitoring_issue`, `none`

## Import

Alerting configs can be imported using the `id`, e.g.:

```
$ terraform import instana_alerting_config.my_alerting_config 60845e4e5e6b9cf8fc2868da
```
