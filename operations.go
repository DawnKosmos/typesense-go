package typesense

import (
	"context"
	"net/http"
	"strings"

	"typesense/api"
)

/* Operations provides snapshot creation and cluster leader election. */
type Operations struct {
	apiClient api.ClientWithResponsesInterface
}

/* Snapshot creates a snapshot in the given directory on the Typesense server. */
func (o *Operations) Snapshot(ctx context.Context, snapshotPath string) (bool, error) {
	response, err := o.apiClient.TakeSnapshotWithResponse(ctx, &api.TakeSnapshotParams{SnapshotPath: snapshotPath})
	if err != nil {
		return false, err
	}
	if response.JSON201 == nil {
		return false, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON201.Success, nil
}

/* Vote asks the contacted node to initiate a cluster leader election. */
func (o *Operations) Vote(ctx context.Context) (bool, error) {
	response, err := o.apiClient.VoteWithResponse(ctx)
	if err != nil {
		return false, err
	}
	if response.JSON200 == nil {
		return false, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Success, nil
}

/* ToggleSlowRequestLog sets the slow request threshold in milliseconds; -1 disables logging. */
func (o *Operations) ToggleSlowRequestLog(ctx context.Context, milliseconds int) (bool, error) {
	response, err := o.apiClient.ToggleSlowRequestLogWithResponse(ctx, api.ToggleSlowRequestLogJSONRequestBody{LogSlowRequestsTimeMs: milliseconds})
	if err != nil {
		return false, err
	}
	if response.JSON201 != nil {
		return response.JSON201.Success, nil
	}
	if response.JSON200 == nil {
		return false, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Success, nil
}

/* ClearCache clears the server's cached search results. */
func (o *Operations) ClearCache(ctx context.Context) (bool, error) {
	response, err := o.apiClient.ClearCacheWithResponse(ctx)
	if err != nil {
		return false, err
	}
	if response.JSON200 == nil {
		return false, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Success, nil
}

/* CompactDB compacts the server's on-disk database. */
func (o *Operations) CompactDB(ctx context.Context) (bool, error) {
	response, err := o.apiClient.CompactDBWithResponse(ctx)
	if err != nil {
		return false, err
	}
	if response.JSON200 == nil {
		return false, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Success, nil
}

/* SchemaChanges lists the progress of ongoing collection schema changes. */
func (o *Operations) SchemaChanges(ctx context.Context) ([]api.SchemaChangeStatus, error) {
	response, err := o.apiClient.GetSchemaChangesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != http.StatusOK || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
