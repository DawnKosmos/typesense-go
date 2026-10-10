package typesense

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"typesense/api"
)

/* A response-only adapter exercises validation of responses supplied by an API implementation. */
type analyticsResponseClient struct {
	api.ClientWithResponsesInterface
	response *api.CreateAnalyticsRuleResponse
}

func (c *analyticsResponseClient) CreateAnalyticsRuleWithResponse(context.Context, api.CreateAnalyticsRuleJSONRequestBody, ...api.RequestEditorFn) (*api.CreateAnalyticsRuleResponse, error) {
	return c.response, nil
}

func TestAnalyticsInvalidArrayResponse(t *testing.T) {
	rules := &AnalyticsRules{apiClient: &analyticsResponseClient{response: &api.CreateAnalyticsRuleResponse{
		Body: []byte("[invalid"), JSON200: &api.CreateAnalyticsRule200JSONResponseBody{},
		HTTPResponse: &http.Response{StatusCode: 200},
	}}}
	if _, err := rules.Create(t.Context(), []*api.AnalyticsRuleCreate{{Name: "clicks"}}); err == nil {
		t.Fatal("malformed analytics array accepted")
	}
}

func TestEmptyMultiSearchResponse(t *testing.T) {
	if err := multiSearchError(nil); err == nil {
		t.Fatal("empty response accepted")
	}
}

func TestTransportInvalidRetryCount(t *testing.T) {
	transport := &clientTransport{retries: -1}
	req, err := http.NewRequest(http.MethodGet, "http://example", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transport.do(req); err == nil || !strings.Contains(err.Error(), "no request attempts") {
		t.Fatalf("invalid retry count accepted: %v", err)
	}
}
