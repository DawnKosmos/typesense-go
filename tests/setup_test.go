package tests

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"typesense"
	data "typesense/tests_data"
)

func startClient(t *testing.T) (*typesense.Client, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), data.ContainerTimeout)
	defer cancel()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        data.ContainerImage,
			ExposedPorts: []string{"8108/tcp"},
			Tmpfs:        map[string]string{"/data": "", "/analytics": ""},
			Cmd:          []string{"--data-dir=/data", "--api-key=" + data.APIKey, "--enable-search-analytics=true", "--analytics-dir=/analytics", "--analytics-flush-interval=60"},
			WaitingFor:   wait.ForHTTP("/health").WithPort("8108/tcp").WithStatusCodeMatcher(func(status int) bool { return status == http.StatusOK }).WithStartupTimeout(data.ContainerTimeout),
		},
		Started: true,
	})
	if container != nil {
		testcontainers.CleanupContainer(t, container)
	}
	requireNoError(t, err)
	host, err := container.Host(ctx)
	requireNoError(t, err)
	port, err := container.MappedPort(ctx, "8108/tcp")
	requireNoError(t, err)
	server := "http://" + net.JoinHostPort(host, port.Port())
	client := authenticatedClient(t, server, data.APIKey)
	return client, server
}
