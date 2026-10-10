package typesense

import (
	"context"
	"time"
)

// Health checks server health within the supplied timeout.
func (c *Client) Health(ctx context.Context, timeout time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	response, err := c.c.HealthWithResponse(ctx)
	if err != nil {
		return false, err
	}
	if response.JSON200 == nil {
		return false, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Ok, nil
}
