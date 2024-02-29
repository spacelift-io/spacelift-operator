package client

import (
	"context"
	"net/http"

	"github.com/shurcooL/graphql"

	spacectl "github.com/spacelift-io/spacectl/client"
	"github.com/spacelift-io/spacectl/client/session"
)

type client struct {
	wraps spacectl.Client
}

// New returns a new instance of a Spacelift Client.
func New(endpoint, keyID, keySecret string) (*client, error) {
	ctx, httpClient := session.Defaults()

	session, err := session.FromAPIKey(ctx, httpClient)(endpoint, keyID, keySecret)
	if err != nil {
		return nil, err
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

func (c *client) Do(req *http.Request) (*http.Response, error) {
	return c.wraps.Do(req)
}
