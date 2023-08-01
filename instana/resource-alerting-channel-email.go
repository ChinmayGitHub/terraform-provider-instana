package instana

import (
	"context"

	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	"github.com/gessnerfl/terraform-provider-instana/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	//AlertingChannelEmailFieldEmails const for the emails field of the alerting channel
	AlertingChannelEmailFieldEmails = "emails"
	//ResourceInstanaAlertingChannelEmail the name of the terraform-provider-instana resource to manage alerting channels of type email
	ResourceInstanaAlertingChannelEmail = "instana_alerting_channel_email"
)

// AlertingChannelEmailEmailsSchemaField schema definition for instana alerting channel emails field
var AlertingChannelEmailEmailsSchemaField = &schema.Schema{
	Type:     schema.TypeSet,
	MinItems: 1,
	Elem: &schema.Schema{
		Type: schema.TypeString,
	},
	Required:    true,
	Description: "The list of emails of the Email alerting channel",
}

// NewAlertingChannelEmailResourceHandle creates the resource handle for Alerting Channels of type Email
func NewAlertingChannelEmailResourceHandle() ResourceHandle[*restapi.AlertingChannel] {
	return &alertingChannelEmailResource{
		metaData: ResourceMetaData{
			ResourceName: ResourceInstanaAlertingChannelEmail,
			Schema: map[string]*schema.Schema{
				AlertingChannelFieldName:        alertingChannelNameSchemaField,
				AlertingChannelFieldFullName:    alertingChannelFullNameSchemaField,
				AlertingChannelEmailFieldEmails: AlertingChannelEmailEmailsSchemaField,
			},
			SchemaVersion: 1,
		},
	}
}

type alertingChannelEmailResource struct {
	metaData ResourceMetaData
}

func (r *alertingChannelEmailResource) MetaData() *ResourceMetaData {
	return &r.metaData
}

func (r *alertingChannelEmailResource) StateUpgraders() []schema.StateUpgrader {
	return []schema.StateUpgrader{
		{
			Type: r.alertingChannelEmailSchemaV0().CoreConfigSchema().ImpliedType(),
			Upgrade: func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
				return rawState, nil
			},
			Version: 0,
		},
	}
}

func (r *alertingChannelEmailResource) GetRestResource(api restapi.InstanaAPI) restapi.RestResource[*restapi.AlertingChannel] {
	return api.AlertingChannels()
}

func (r *alertingChannelEmailResource) SetComputedFields(d *schema.ResourceData) {
	//No computed fields defined
}

func (r *alertingChannelEmailResource) UpdateState(d *schema.ResourceData, alertingChannel *restapi.AlertingChannel, formatter utils.ResourceNameFormatter) error {
	emails := alertingChannel.Emails
	d.Set(AlertingChannelFieldName, formatter.UndoFormat(alertingChannel.Name))
	d.Set(AlertingChannelFieldFullName, alertingChannel.Name)
	d.Set(AlertingChannelEmailFieldEmails, emails)
	d.SetId(alertingChannel.ID)
	return nil
}

func (r *alertingChannelEmailResource) MapStateToDataObject(d *schema.ResourceData, formatter utils.ResourceNameFormatter) (*restapi.AlertingChannel, error) {
	name := computeFullAlertingChannelNameString(d, formatter)
	return &restapi.AlertingChannel{
		ID:     d.Id(),
		Name:   name,
		Kind:   restapi.EmailChannelType,
		Emails: ReadStringSetParameterFromResource(d, AlertingChannelEmailFieldEmails),
	}, nil
}

func (r *alertingChannelEmailResource) alertingChannelEmailSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			AlertingChannelFieldName:     alertingChannelNameSchemaField,
			AlertingChannelFieldFullName: alertingChannelFullNameSchemaField,
			AlertingChannelEmailFieldEmails: {
				Type:     schema.TypeList,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "The list of emails of the Email alerting channel",
			},
		},
	}
}
