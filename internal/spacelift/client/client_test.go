package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/shurcooL/graphql"
	"github.com/spacelift-io/spacelift-operator/internal/spacelift/client"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	t.Skip("skipping test")

	// TODO(michal): Make this test more useful, this is just a simple test that checks if client is working
	endpoint := os.Getenv("SPACELIFT_API_KEY_ENDPOINT")
	keyID := os.Getenv("SPACELIFT_API_KEY_ID")
	keySecret := os.Getenv("SPACELIFT_API_KEY_SECRET")

	spaceliftCli, err := client.New(endpoint, keyID, keySecret)
	assert.Nil(t, err)

	var query struct {
		Stack struct {
			ID             string `graphql:"id"`
			Administrative bool   `graphql:"administrative"`
		} `graphql:"stack(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID("end-to-end-autoconfirm"),
	}

	err = spaceliftCli.Query(context.Background(), &query, variables)
	assert.Nil(t, err)

	assert.Equal(t, query.Stack.ID, "end-to-end-autoconfirm")
	assert.True(t, query.Stack.Administrative)
}
