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
	tokenField = field.StringField(
		"token",
		field.WithDescription("API token for authenticating requests to Zuper."),
		field.WithRequired(true),
	)

	ConfigurationFields = []field.SchemaField{
		apiUrlField,
		tokenField,
	}

	FieldRelationships = []field.SchemaFieldRelationship{}
)

func ValidateConfig(v *viper.Viper) error {
	return nil
}
