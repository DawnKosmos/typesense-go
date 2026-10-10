package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

type AnalyticsEvents struct {
	apiClient api.ClientWithResponsesInterface
}

func (a *AnalyticsEvents) Create(ctx context.Context, eventSchema *api.AnalyticsEvent) (*api.AnalyticsEventCreateResponse, error) {
	if eventSchema == nil {
		return nil, fmt.Errorf("typesense: analytics event must not be nil")
	}
	response, err := a.apiClient.CreateAnalyticsEventWithResponse(ctx, api.CreateAnalyticsEventJSONRequestBody(*eventSchema))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

func (a *AnalyticsEvents) Retrieve(ctx context.Context, params *api.GetAnalyticsEventsParams) (*api.AnalyticsEventsResponse, error) {
	response, err := a.apiClient.GetAnalyticsEventsWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
