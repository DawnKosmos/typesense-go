package typesense_test

import (
	"errors"
	"testing"

	"typesense"
)

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func requireHTTPError(t *testing.T, err error, status int) {
	t.Helper()
	var httpErr *typesense.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != status {
		t.Fatalf("expected HTTPError %d, got %v", status, err)
	}
}
