package model

const (
	ChannelTypeSlack     = "slack"
	ChannelTypeWebhook   = "webhook"
	ChannelTypePagerduty = "pagerduty"
	ChannelTypeOpsgenie  = "opsgenie"
	ChannelTypeMSTeams   = "msteams"
	ChannelTypeEmail     = "email"
)

//nolint:gochecknoglobals
var ChannelTypes = []string{
	ChannelTypeSlack,
	ChannelTypeWebhook,
	ChannelTypePagerduty,
	ChannelTypeOpsgenie,
	ChannelTypeMSTeams,
	ChannelTypeEmail,
}

// NotificationChannel model.
type NotificationChannel struct {
	ID               int                `json:"id"`
	Name             string             `json:"name"`
	Type             string             `json:"type"`
	SlackConfigs     []SlackConfig      `json:"slack_configs,omitempty"`
	WebhookConfigs   []WebhookConfig    `json:"webhook_configs,omitempty"`
	PagerdutyConfigs []PagerdutyConfig  `json:"pagerduty_configs,omitempty"`
	OpsgenieConfigs  []OpsgenieConfig   `json:"opsgenie_configs,omitempty"`
	MSTeamsConfigs   []MSTeamsConfig    `json:"msteams_configs,omitempty"`
	EmailConfigs     []EmailConfig      `json:"email_configs,omitempty"`
}

// SlackConfig model.
type SlackConfig struct {
	ApiUrl string `json:"api_url"`
	Channel string `json:"channel,omitempty"`
	Title   string `json:"title,omitempty"`
	Text    string `json:"text,omitempty"`
}

// WebhookConfig model.
type WebhookConfig struct {
	ApiUrl   string `json:"api_url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// PagerdutyConfig model.
type PagerdutyConfig struct {
	RoutingKey string            `json:"routing_key,omitempty"`
	ServiceKey string            `json:"service_key,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
}

// OpsgenieConfig model.
type OpsgenieConfig struct {
	ApiKey string `json:"api_key"`
	ApiUrl string `json:"api_url,omitempty"`
}

// MSTeamsConfig model.
type MSTeamsConfig struct {
	WebhookUrl string `json:"webhook_url"`
}

// EmailConfig model.
type EmailConfig struct {
	SendResolved bool              `json:"send_resolved,omitempty"`
	To           string            `json:"to"`
	Html         string            `json:"html,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
}
