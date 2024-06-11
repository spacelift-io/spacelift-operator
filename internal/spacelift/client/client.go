package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	v1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/spacelift-io/spacelift-operator/internal/spacelift/client/session"
)

// spaceliftClient is the authenticated client that can be used to interact with Spacelift
var spaceliftClient Client

const (
	SecretName                 = "spacelift-credentials"      //nolint:gosec
	SpaceliftApiKeyEndpointKey = "SPACELIFT_API_KEY_ENDPOINT" //nolint:gosec
	SpaceliftApiKeyIDKey       = "SPACELIFT_API_KEY_ID"       //nolint:gosec
	SpaceliftApiKeySecretKey   = "SPACELIFT_API_KEY_SECRET"   //nolint:gosec
)

var DefaultClient = GetSpaceliftClient

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

	apiEndpoint := string(secret.Data[SpaceliftApiKeyEndpointKey])
	apiKeyID := string(secret.Data[SpaceliftApiKeyIDKey])
	apiKeySecret := string(secret.Data[SpaceliftApiKeySecretKey])

	session, err := func() (session.Session, error) {
		sessionCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		return session.New(sessionCtx, http.DefaultClient, apiEndpoint, apiKeyID, apiKeySecret)
	}()
	if err != nil {
		return nil, errors.Wrap(err, "could not create session from Spacelift API key")
	}

	spaceliftClient = New(http.DefaultClient, session)

	return spaceliftClient, nil
}

type client struct {
	wraps   *http.Client
	session session.Session
}

// New returns a new instance of a Spacelift Client.
func New(wraps *http.Client, session session.Session) Client {
	return &client{wraps: wraps, session: session}
}

func (c *client) Mutate(ctx context.Context, mutation interface{}, variables map[string]interface{}, opts ...graphql.RequestOption) error {
	logger := log.FromContext(ctx)

	apiClient, err := c.apiClient(ctx)
	if err != nil {
		return nil
	}

	err = apiClient.Mutate(ctx, mutation, variables, opts...)
	if err != nil && strings.Contains(err.Error(), "unauthorized") {
		logger.Error(err, "Server returned an unauthorized response - retrying request with a new token")
		c.session.RefreshToken(ctx)

		// Try again in case refreshing the token fixes the problem
		apiClient, err = c.apiClient(ctx)
		if err != nil {
			return err
		}

		err = apiClient.Mutate(ctx, mutation, variables, opts...)
	}

	return err
}

func (c *client) Query(ctx context.Context, query interface{}, variables map[string]interface{}, opts ...graphql.RequestOption) error {
	logger := log.FromContext(ctx)

	apiClient, err := c.apiClient(ctx)
	if err != nil {
		return err
	}

	err = apiClient.Query(ctx, query, variables)
	if err != nil && strings.Contains(err.Error(), "unauthorized") {
		logger.Error(err, "Server returned an unauthorized response - retrying request with a new token")
		c.session.RefreshToken(ctx)

		// Try again in case refreshing the token fixes the problem
		apiClient, err = c.apiClient(ctx)
		if err != nil {
			return err
		}

		err = apiClient.Query(ctx, query, variables, opts...)
	}

	return err
}

func (c *client) URL(format string, a ...interface{}) string {
	endpoint := c.session.Endpoint()

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		panic(err) // Impossible condition.
	}

	endpointURL.Path = fmt.Sprintf(format, a...)

	return endpointURL.String()
}

func (c *client) apiClient(ctx context.Context) (*graphql.Client, error) {
	bearerToken, err := c.session.BearerToken(ctx)
	if err != nil {
		return nil, err
	}

	return graphql.NewClient(c.session.Endpoint(), oauth2.NewClient(
		context.WithValue(ctx, oauth2.HTTPClient, c.wraps), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: bearerToken},
		)), graphql.WithHeader("Spacelift-Client-Type", "k8s-operator")), nil
}
