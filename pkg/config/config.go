package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	apiUrlField = field.StringField(
		"api-url",
		field.WithDisplayName("API URL"),
		field.WithDescription("The URL of the API."),
		field.WithRequired(true),
	)
	apiKeyField = field.StringField(
		"api-key",
		field.WithDisplayName("API key"),
		field.WithDescription("API key for authenticating requests to Zuper."),
		field.WithIsSecret(true),
		field.WithRequired(true),
	)
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	[]field.SchemaField{
		apiUrlField,
		apiKeyField,
	},
	field.WithConnectorDisplayName("Zuper"),
	field.WithHelpUrl("/docs/baton/zuper"),
	field.WithIconUrl("/static/app-icons/zuper.svg"),
)
