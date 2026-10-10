package typesense

import "fmt"

// HTTPError preserves the status and body of an unexpected HTTP response.
type HTTPError struct {
	Status int
	Body   []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("status: %d response: %s", e.Status, e.Body)
}
