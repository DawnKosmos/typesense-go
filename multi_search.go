package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"typesense/api"
	"typesense/optional"
)

/* VIP: Reimplement MultiSearch; this is the initial wrapper around the generated API. */
type MultiSearch struct {
	apiClient api.ClientWithResponsesInterface
}

/* Perform returns separate results for each search, preserving individual search errors in the results. */
func (m *MultiSearch) Perform(ctx context.Context, commonSearchParams *api.MultiSearchParams, searchParams api.MultiSearchSearchesParameter) (*api.MultiSearchResult, error) {
	response, err := m.apiClient.MultiSearchWithResponse(ctx, commonSearchParams, searchParams)
	if err != nil {
		return nil, err
	}
	if err := multiSearchError(response); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

/* PerformWithContentType sends the searches using the specified content type and returns the response wrapper. */
func (m *MultiSearch) PerformWithContentType(ctx context.Context, commonSearchParams *api.MultiSearchParams, searchParams api.MultiSearchSearchesParameter, contentType string) (*api.MultiSearchResponse, error) {
	body, err := json.Marshal(searchParams)
	if err != nil {
		return nil, err
	}
	response, err := m.apiClient.MultiSearchWithBodyWithResponse(ctx, commonSearchParams, contentType, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if err := multiSearchError(response); err != nil {
		return nil, err
	}
	if len(response.Body) == 0 {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response, nil
}

/* PerformUnion merges searches into one result and rejects an explicitly false union option. */
func (m *MultiSearch) PerformUnion(ctx context.Context, commonSearchParams *api.MultiSearchParams, searchParams api.MultiSearchSearchesParameter) (*api.SearchResult, error) {
	if searchParams.Union != optional.None && searchParams.Union != optional.True {
		return nil, fmt.Errorf("typesense: union must be unset or true when calling PerformUnion")
	}
	searchParams.Union = optional.True
	response, err := m.apiClient.MultiSearchWithResponse(ctx, commonSearchParams, searchParams)
	if err != nil {
		return nil, err
	}
	if err := multiSearchError(response); err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	var result api.SearchResult
	if err := json.Unmarshal(response.Body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func multiSearchError(response *api.MultiSearchResponse) error {
	if response == nil {
		return fmt.Errorf("typesense: empty multi-search response")
	}
	if response.StatusCode() != http.StatusOK {
		return &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	var failure struct {
		Code  *int    `json:"code"`
		Error *string `json:"error"`
	}
	if json.Unmarshal(response.Body, &failure) == nil && failure.Code != nil && failure.Error != nil {
		return &HTTPError{Status: *failure.Code, Body: response.Body}
	}
	return nil
}
