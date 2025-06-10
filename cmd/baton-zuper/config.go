package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	apiUrlField = field.StringField(
		"api-url",
		field.WithDescription("The URL of the API."),
		field.WithRequired(true),
	)
	apiKeyField = field.StringField(
		"api-key",
		field.WithDescription("API key for authenticating requests to Zuper."),
		field.WithRequired(true),
	)

	ConfigurationFields = []field.SchemaField{
		apiUrlField,
		apiKeyField,
	}

	FieldRelationships = []field.SchemaFieldRelationship{}
)

func ValidateConfig(v *viper.Viper) error {
	return nil
}
