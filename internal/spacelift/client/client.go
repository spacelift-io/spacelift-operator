package client

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	spacectl "github.com/spacelift-io/spacectl/client"
	"github.com/spacelift-io/spacectl/client/session"
	v1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// spaceliftClient is the authenticated client that can be used to interact with Spacelift
var spaceliftClient Client

const (
	SecretName                 = "spacelift-credentials"      //nolint:gosec
	SpaceliftApiKeyEndpointKey = "SPACELIFT_API_KEY_ENDPOINT" //nolint:gosec
	SpaceliftApiKeyIDKey       = "SPACELIFT_API_KEY_ID"       //nolint:gosec
	SpaceliftApiKeySecretKey   = "SPACELIFT_API_KEY_SECRET"   //nolint:gosec
)

func GetSpaceliftClient(ctx context.Context, client k8sclient.Client, namespace string) (Client, error) {
	if spaceliftClient != nil {
		return spaceliftClient, nil
	}

	var secret v1.Secret

	if err := client.Get(
		ctx,
		k8sclient.ObjectKey{
			Namespace: namespace,
			Name:      SecretName,
		},
		&secret,
	); err != nil {
		return nil, errors.Wrap(err, "failed to get spacelift credentials secret")
	}

	endpoint := string(secret.Data[SpaceliftApiKeyEndpointKey])
	apiKeyID := string(secret.Data[SpaceliftApiKeyIDKey])
	apiKeySecret := string(secret.Data[SpaceliftApiKeySecretKey])

	newSpaceliftClient, err := New(endpoint, apiKeyID, apiKeySecret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create spacelift client")
	}

	spaceliftClient = newSpaceliftClient

	return spaceliftClient, nil
}

type client struct {
	wraps spacectl.Client
}

// New returns a new instance of a Spacelift Client.
func New(endpoint, keyID, keySecret string) (*client, error) {
	ctx, httpClient := session.Defaults()

	session, err := session.FromAPIKey(ctx, httpClient)(endpoint, keyID, keySecret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create spacelift session")
	}

	spaceliftClient := spacectl.New(httpClient, session)

	return &client{wraps: spaceliftClient}, nil
}

func (c *client) Mutate(ctx context.Context, mutation interface{}, variables map[string]interface{}, opts ...graphql.RequestOption) error {
	return c.wraps.Mutate(ctx, mutation, variables, opts...)
}

func (c *client) Query(ctx context.Context, query interface{}, variables map[string]interface{}, opts ...graphql.RequestOption) error {
	return c.wraps.Query(ctx, query, variables, opts...)
}

func (c *client) URL(format string, a ...interface{}) string {
	return c.wraps.URL(format, a...)
}
