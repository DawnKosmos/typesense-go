package typesense

import (
	"context"

	"typesense/api"
)

// Analytics provides access to analytics events and rules.
type Analytics struct {
	apiClient api.ClientWithResponsesInterface
}

func (a *Analytics) Events() *AnalyticsEvents {
	return &AnalyticsEvents{apiClient: a.apiClient}
}

func (a *Analytics) Rules() *AnalyticsRules {
	return &AnalyticsRules{apiClient: a.apiClient}
}

func (a *Analytics) Rule(name string) *AnalyticsRule {
	return &AnalyticsRule{apiClient: a.apiClient, ruleName: name}
}

/* Flush writes buffered analytics data to persistent storage. */
func (a *Analytics) Flush(ctx context.Context) (bool, error) {
	response, err := a.apiClient.FlushAnalyticsWithResponse(ctx)
	if err != nil {
		return false, err
	}
	if response.JSON200 == nil {
		return false, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Ok, nil
}

/* Status returns the sizes of internal analytics buffers and queues. */
func (a *Analytics) Status(ctx context.Context) (*api.AnalyticsStatus, error) {
	response, err := a.apiClient.GetAnalyticsStatusWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
