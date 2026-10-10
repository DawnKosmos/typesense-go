package typesense

import (
	"context"

	"typesense/api"
)

/* Debug returns server diagnostics described by the API specification. */
func (c *Client) Debug(ctx context.Context) (*api.Debug200JSONResponseBody, error) {
	response, err := c.c.DebugWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
