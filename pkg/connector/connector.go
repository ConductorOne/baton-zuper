package connector

import (
	"context"
	"io"
	"net/http"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/conductorone/baton-zuper/pkg/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Connector struct {
	client *client.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client),
		newRoleBuilder(d.client),
		newAccessRoleBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Zuper",
		Description: "Connector syncs data users and teams of Zuper",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, apiUrl string, token string) (*Connector, error) {
	l := ctxzap.Extract(ctx)
	httpClient := uhttp.NewBaseHttpClient(&http.Client{})
	zuperClient, err := client.New(ctx, client.NewClient(ctx, apiUrl, token, httpClient))
	if err != nil {
		l.Error("error creating Zuper client", zap.Error(err))
		return nil, err
	}
	return &Connector{
		client: zuperClient,
	}, nil
}
