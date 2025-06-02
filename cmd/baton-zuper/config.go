package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	apiUrlField = field.StringField(
		"api-url",
		field.WithDescription("The URL of the API."),
		field.WithDefaultValue("https://staging.zuperpro.com"),
		field.WithRequired(false),
	)
	tokenField = field.StringField(
		"token",
		field.WithDescription("Your Zuper token"),
		field.WithRequired(true),
	)

	ConfigurationFields = []field.SchemaField{
		apiUrlField,
		tokenField,
	}

	FieldRelationships = []field.SchemaFieldRelationship{}
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	return nil
}
