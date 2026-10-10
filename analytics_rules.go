package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"typesense/api"
)

// AnalyticsRules provides analytics rule creation and listing.
type AnalyticsRules struct {
	apiClient api.ClientWithResponsesInterface
}

// Create returns created rules. Partial failures return the successful rules and an error.
func (a *AnalyticsRules) Create(ctx context.Context, rules []*api.AnalyticsRuleCreate) ([]*api.AnalyticsRule, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("typesense: analytics rules must not be empty")
	}
	values := make([]api.AnalyticsRuleCreate, len(rules))
	for i, rule := range rules {
		if rule == nil {
			return nil, fmt.Errorf("typesense: analytics rule at index %d must not be nil", i)
		}
		values[i] = *rule
	}
	var body api.CreateAnalyticsRuleJSONBody
	if err := body.FromCreateAnalyticsRuleJSONBody1(values); err != nil {
		return nil, err
	}
	response, err := a.apiClient.CreateAnalyticsRuleWithResponse(ctx, api.CreateAnalyticsRuleJSONRequestBody(body))
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	// The success schema permits a single rule or an array containing rules and
	// per-rule errors. Inspect error entries before decoding them as rule models.
	data := bytes.TrimSpace(response.Body)
	entries := []json.RawMessage{data}
	if len(data) > 0 && data[0] == '[' {
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, err
		}
	}
	result := make([]*api.AnalyticsRule, 0, len(entries))
	var failures []error
	for i, entry := range entries {
		var failure struct {
			Error *string `json:"error"`
		}
		if err := json.Unmarshal(entry, &failure); err != nil {
			return result, err
		}
		if failure.Error != nil {
			failures = append(failures, fmt.Errorf("typesense: analytics rule at index %d: %s", i, *failure.Error))
			continue
		}
		var rule *api.AnalyticsRule
		if err := json.Unmarshal(entry, &rule); err != nil {
			return result, err
		}
		if rule == nil {
			return result, fmt.Errorf("typesense: null analytics rule at index %d", i)
		}
		result = append(result, rule)
	}
	return result, errors.Join(failures...)
}

// Retrieve lists all analytics rules.
func (a *AnalyticsRules) Retrieve(ctx context.Context) ([]*api.AnalyticsRule, error) {
	response, err := a.apiClient.RetrieveAnalyticsRulesWithResponse(ctx, &api.RetrieveAnalyticsRulesParams{})
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	rules := make([]*api.AnalyticsRule, len(response.JSON200))
	for i := range response.JSON200 {
		rules[i] = &response.JSON200[i]
	}
	return rules, nil
}
