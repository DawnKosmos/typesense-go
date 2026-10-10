package tests

import (
	"context"
	"errors"
	"testing"
	"typesense"
	data "typesense/tests_data"
)

func health(t *testing.T, client *typesense.Client) {
	ok, err := client.Health(t.Context(), data.RequestTimeout)
	requireNoError(t, err)
	if !ok {
		t.Fatal("server is not healthy")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = client.Health(ctx, data.RequestTimeout)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if typesense.Version() == "" {
		t.Fatal("missing generated API version")
	}
}
