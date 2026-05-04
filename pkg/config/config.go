package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	ApiUrlField = field.StringField(
		"api-url",
		field.WithDisplayName("Zuper API URL"),
		field.WithDescription("The URL of the Zuper API."),
		field.WithRequired(true),
	)
	ApiKeyField = field.StringField(
		"api-key",
		field.WithDisplayName("Zuper API Key"),
		field.WithDescription("API key for authenticating requests to Zuper."),
		field.WithIsSecret(true),
		field.WithRequired(true),
	)

	ConfigurationFields = []field.SchemaField{
		ApiUrlField,
		ApiKeyField,
	}

	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	ConfigurationFields,
	field.WithConstraints(FieldRelationships...),
	field.WithConnectorDisplayName("Zuper"),
	field.WithHelpUrl("/docs/baton/zuper"),
	field.WithIconUrl("/static/app-icons/zuper.svg"),
)
