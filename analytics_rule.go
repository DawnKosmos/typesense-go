package typesense

import (
	"context"
	"fmt"

	"typesense/api"
)

type AnalyticsRule struct {
	apiClient api.ClientWithResponsesInterface
	ruleName  string
}

func (a *AnalyticsRule) Delete(ctx context.Context) (*api.AnalyticsRule, error) {
	response, err := a.apiClient.DeleteAnalyticsRuleWithResponse(ctx, a.ruleName)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

func (a *AnalyticsRule) Retrieve(ctx context.Context) (*api.AnalyticsRule, error) {
	response, err := a.apiClient.RetrieveAnalyticsRuleWithResponse(ctx, a.ruleName)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

func (a *AnalyticsRule) Update(ctx context.Context, ruleSchema *api.AnalyticsRuleUpdate) (*api.AnalyticsRule, error) {
	if ruleSchema == nil {
		return nil, fmt.Errorf("typesense: analytics rule update must not be nil")
	}
	response, err := a.apiClient.UpsertAnalyticsRuleWithResponse(ctx, a.ruleName, api.UpsertAnalyticsRuleJSONRequestBody(*ruleSchema))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}
