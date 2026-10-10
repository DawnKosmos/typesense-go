package typesense_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

type failureReader struct{ err error }

func (r failureReader) Read([]byte) (int, error) { return 0, r.err }

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func responseClient(t *testing.T, status int, contentType, payload string) *typesense.Client {
	t.Helper()
	return injectedClient(t, requestDoerFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(payload)), Request: r}, nil
	}))
}

func TestStreamingReadFailuresCloseBodies(t *testing.T) {
	sentinel := errors.New("response read failed")
	for _, operation := range []string{"Export", "ImportJsonl", "StemmingJsonl"} {
		t.Run(operation, func(t *testing.T) {
			body := &trackedBody{Reader: failureReader{err: sentinel}}
			client := injectedClient(t, requestDoerFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 404, Body: body, Header: make(http.Header), Request: r}, nil
			}))
			var err error
			switch operation {
			case "Export":
				_, err = client.Collection[data.Product](data.CollectionName).Documents().Export(t.Context(), nil)
			case "ImportJsonl":
				_, err = client.Collection[data.Product](data.CollectionName).Documents().ImportJsonl(t.Context(), strings.NewReader("{}"), nil)
			case "StemmingJsonl":
				_, err = client.Stemming().Dictionaries().UpsertJsonl(t.Context(), data.DictionaryID, strings.NewReader("{}"))
			}
			if !errors.Is(err, sentinel) || !body.closed {
				t.Fatalf("read error or body cleanup lost: %v, closed=%v", err, body.closed)
			}
		})
	}
}

func TestImportPartialDecoding(t *testing.T) {
	for _, tail := range []string{"invalid JSON", "null"} {
		t.Run(tail, func(t *testing.T) {
			client := responseClient(t, 200, "text/plain", "{\"success\":true,\"id\":\"first\"}\n"+tail)
			result, err := client.Collection[data.Product](data.CollectionName).Documents().Import(t.Context(), data.Products(), nil)
			if err == nil || len(result) != 1 || !result[0].Success || result[0].ID != "first" {
				t.Fatalf("partial results lost: %+v, %v", result, err)
			}
			client = responseClient(t, 200, "text/plain", "{\"word\":\"people\",\"root\":\"person\"}\n"+tail)
			words, err := client.Stemming().Dictionaries().Upsert(t.Context(), data.DictionaryID, []api.StemmingDictionaryWord{{Word: "people", Root: "person"}})
			if err == nil || len(words) != 1 || words[0].Root != "person" {
				t.Fatalf("partial stemming results lost: %+v, %v", words, err)
			}
		})
	}
	t.Run("LargeJSONLRecord", func(t *testing.T) {
		payload := "{\"success\":true,\"document\":\"" + strings.Repeat("x", 128*1024) + "\"}"
		client := responseClient(t, 200, "text/plain", payload)
		result, err := client.Collection[data.Product](data.CollectionName).Documents().Import(t.Context(), data.Products()[:1], nil)
		requireNoError(t, err)
		if len(result) != 1 || len(result[0].Document.(string)) != 128*1024 {
			t.Fatal("large JSONL record truncated")
		}
	})
	t.Run("EncodingFailure", func(t *testing.T) {
		client := responseClient(t, 200, "text/plain", "")
		if _, err := client.Collection[data.Product](data.CollectionName).Documents().Import(t.Context(), []any{make(chan int)}, nil); err == nil {
			t.Fatal("non-JSON document accepted")
		}
	})
}

func TestMultiSearchDecodingErrors(t *testing.T) {
	for _, operation := range []string{"Perform", "Union", "Custom"} {
		t.Run(operation, func(t *testing.T) {
			client := responseClient(t, 200, "text/plain", "")
			var err error
			switch operation {
			case "Perform":
				_, err = client.MultiSearch().Perform(t.Context(), nil, api.MultiSearchSearchesParameter{})
			case "Union":
				_, err = client.MultiSearch().PerformUnion(t.Context(), nil, api.MultiSearchSearchesParameter{})
			case "Custom":
				_, err = client.MultiSearch().PerformWithContentType(t.Context(), nil, api.MultiSearchSearchesParameter{}, "application/json")
			}
			requireHTTPError(t, err, 200)
		})
	}
	client := responseClient(t, 200, "application/json", `{"code":400,"error":"invalid parameters"}`)
	_, err := client.MultiSearch().Perform(t.Context(), nil, api.MultiSearchSearchesParameter{})
	requireHTTPError(t, err, 400)
	client = responseClient(t, 200, "application/json", `{"found":"invalid"}`)
	if _, err := client.MultiSearch().PerformUnion(t.Context(), nil, api.MultiSearchSearchesParameter{}); err == nil {
		t.Fatal("invalid union result decoded")
	}
	if _, err := client.MultiSearch().PerformWithContentType(t.Context(), nil, api.MultiSearchSearchesParameter{Union: optional.Bool(255)}, "application/json"); err == nil {
		t.Fatal("invalid boolean encoded")
	}
}

func TestDocumentDirtyValues(t *testing.T) {
	dirty := api.DirtyValues("coerce_or_reject")
	options := &api.DocumentIndexParameters{DirtyValues: &dirty}
	client := injectedClient(t, requestDoerFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("dirty_values") != string(dirty) {
			t.Fatalf("dirty_values missing: %s", r.URL)
		}
		status := 200
		if r.Method == http.MethodPost {
			status = 201
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"product-1"}`)), Request: r}, nil
	}))
	_, err := client.Collection[data.Product](data.CollectionName).Documents().Create(t.Context(), data.Products()[0], options)
	requireNoError(t, err)
	_, err = client.Collection[data.Product](data.CollectionName).Document(data.FirstID).Update(t.Context(), map[string]any{}, options)
	requireNoError(t, err)
}

func TestSlowRequestLogLegacyStatus(t *testing.T) {
	client := responseClient(t, 200, "application/json", `{"success":true}`)
	ok, err := client.Operations().ToggleSlowRequestLog(t.Context(), -1)
	requireNoError(t, err)
	if !ok {
		t.Fatal("legacy 200 success rejected")
	}
}

func TestAnalyticsRuleDecodingErrors(t *testing.T) {
	for _, payload := range []string{`["invalid rule"]`, `[{"name":42}]`, `[null]`} {
		t.Run(payload, func(t *testing.T) {
			client := responseClient(t, 200, "application/json", payload)
			if _, err := client.Analytics().Rules().Create(t.Context(), []*api.AnalyticsRuleCreate{{Name: data.RuleName}}); err == nil {
				t.Fatal("invalid rule response accepted")
			}
		})
	}
	client := responseClient(t, 200, "application/json", `{ "name":"clicks", "collection":"products", "type":"log", "event_type":"click" }`)
	rules, err := client.Analytics().Rules().Create(t.Context(), []*api.AnalyticsRuleCreate{{Name: data.RuleName}})
	requireNoError(t, err)
	if len(rules) != 1 || rules[0].Name != "clicks" {
		t.Fatalf("single rule response not decoded: %+v", rules)
	}
	if _, err := client.Analytics().Rules().Create(t.Context(), []*api.AnalyticsRuleCreate{{Params: &api.AnalyticsRuleCreate_Params{ExpandQuery: optional.Bool(255)}}}); err == nil {
		t.Fatal("invalid rule boolean encoded")
	}
}
