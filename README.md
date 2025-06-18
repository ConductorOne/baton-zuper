![Baton Logo](./baton-logo.png)

# `baton-zuper` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-zuper.svg)](https://pkg.go.dev/github.com/conductorone/baton-zuper) ![main ci](https://github.com/conductorone/baton-zuper/actions/workflows/main.yaml/badge.svg)

`baton-zuper` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

## Connector Capabilities

1. **Resources synced**:

   - Users
   - Teams
   - Roles
   - Access Roles

2. **Account provisioning**

   - Users

3. **Entitlement provisioning**

   - Assign User To Team
   - Unassign User To Team
   - Update a User's Role
   - Update a User's Access Role

## Connector Credentials

1. **API URL**
2. **API KEY**

### Obtaining Credentials

1. Log in to [Zuper Pro](https://staging.zuperpro.com/login).
2. Navigate to **Settings** → **Account Settings** → **API Keys**.
3. Click on **New API Key**, enter a name for your key, and click **Generate**. The API key will be displayed—make sure to copy and save it securely, as it may not be shown again.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-zuper
baton-zuper
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-zuper:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-zuper/cmd/baton-zuper@main

baton-zuper

baton resources
```

# Data Model

`baton-zuper` will pull down information about the following resources:

- Users
- Teams
- Roles
- Access Roles

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-zuper` Command Line Usage

```
baton-zuper

Usage:
  baton-zuper [flags]
  baton-zuper [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --api-url   string             the API URL provided by Zuper
      --api-key   string             the API key generated in Zuper
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-zuper
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-zuper

Use "baton-zuper [command] --help" for more information about a command.
```
