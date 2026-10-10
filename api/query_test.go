package api

import (
	"net/url"
	"testing"

	"github.com/oapi-codegen/runtime"
	"typesense/optional"
)

func TestQueryOmissionAndBooleanValues(t *testing.T) {
	options := runtime.StyleParamOptions{ParamLocation: runtime.ParamLocationQuery, Type: "object"}
	value := struct {
		Absent   optional.Bool `json:"absent,omitzero"`
		Disabled optional.Bool `json:"disabled,omitzero"`
		Enabled  optional.Bool `json:"enabled,omitzero"`
		Large    int64         `json:"large"`
	}{Disabled: optional.False, Enabled: optional.True, Large: 9007199254740993}
	encoded, err := styleParamWithOptions("form", true, "parameters", value, options)
	if err != nil {
		t.Fatal(err)
	}
	query, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if query.Has("absent") || query.Get("disabled") != "false" || query.Get("enabled") != "true" || query.Get("large") != "9007199254740993" {
		t.Fatalf("query values or omission changed: %s", encoded)
	}
	for _, value := range []any{nil, (*SearchParameters)(nil)} {
		encoded, err := styleParamWithOptions("form", true, "parameters", value, options)
		if err != nil || encoded != "" {
			t.Fatalf("nil object emitted: %q, %v", encoded, err)
		}
	}
	options.Type = "boolean"
	for _, tc := range []struct {
		value optional.Bool
		want  string
	}{{optional.None, ""}, {optional.False, "enabled=false"}, {optional.True, "enabled=true"}} {
		encoded, err := styleParamWithOptions("form", true, "enabled", tc.value, options)
		if err != nil || encoded != tc.want {
			t.Fatalf("boolean %v: %q, %v", tc.value, encoded, err)
		}
	}
	if _, err := styleParamWithOptions("form", true, "enabled", optional.Bool(255), options); err == nil {
		t.Fatal("invalid optional boolean accepted")
	}
	options.Type = "object"
	if _, err := styleParamWithOptions("form", true, "parameters", map[string]any{"bad": make(chan int)}, options); err == nil {
		t.Fatal("non-JSON query accepted")
	}
	if _, err := styleParamWithOptions("invalid", true, "parameters", value, options); err == nil {
		t.Fatal("invalid query style accepted")
	}
}

func TestAnalyticsRequestUnionJSON(t *testing.T) {
	var union CreateAnalyticsRuleJSONBody
	if err := union.FromCreateAnalyticsRuleJSONBody1([]AnalyticsRuleCreate{{Name: "clicks", Collection: "products", EventType: "click", Type: Log}}); err != nil {
		t.Fatal(err)
	}
	body := CreateAnalyticsRuleJSONRequestBody(union)
	encoded, err := body.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var decoded CreateAnalyticsRuleJSONRequestBody
	if err := decoded.UnmarshalJSON(encoded); err != nil {
		t.Fatal(err)
	}
	rules, err := CreateAnalyticsRuleJSONBody(decoded).AsCreateAnalyticsRuleJSONBody1()
	if err != nil || len(rules) != 1 || rules[0].Name != "clicks" {
		t.Fatalf("request union did not round-trip: %+v, %v", rules, err)
	}
	union = CreateAnalyticsRuleJSONBody{}
	if err := union.FromCreateAnalyticsRuleJSONBody1([]AnalyticsRuleCreate{{Params: &AnalyticsRuleCreate_Params{ExpandQuery: optional.Bool(255)}}}); err == nil {
		t.Fatal("union ignored invalid boolean")
	}
}
