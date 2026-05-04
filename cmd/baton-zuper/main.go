package main

import (
	"context"

	cfg "github.com/conductorone/baton-zuper/pkg/config"
	"github.com/conductorone/baton-zuper/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/config"
)

var version = "dev"

func main() {
	ctx := context.Background()
	config.RunConnector(ctx,
		"baton-zuper",
		version,
		cfg.Config,
		connector.New,
	)
}
