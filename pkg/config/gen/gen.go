package main

import (
	"github.com/conductorone/baton-sdk/pkg/config"
	cfg "github.com/conductorone/baton-zuper/pkg/config"
)

func main() {
	config.Generate("zuper", cfg.Config)
}
