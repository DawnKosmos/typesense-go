package tests

import (
	"testing"
	"time"

	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func analytics(t *testing.T, client *typesense.Client) {
	selected := createCollection(t, client, data.AnalyticsCollectionName)
	_, err := selected.Documents().Create(t.Context(), data.Products()[0], nil)
	requireNoError(t, err)
	rules := client.Analytics().Rules()
	rule := &api.AnalyticsRuleCreate{Name: data.RuleName, Collection: data.AnalyticsCollectionName, EventType: "click", Type: api.Log, RuleTag: data.RuleTag}
	created, err := rules.Create(t.Context(), []*api.AnalyticsRuleCreate{rule})
	requireNoError(t, err)
	if len(created) != 1 || created[0].Name != rule.Name || created[0].Type != api.Log {
		t.Fatalf("unexpected created rules: %+v", created)
	}
	cleanupDelete(t, "analytics rule", client.Analytics().Rule(data.RuleName).Delete, 400)
	retrieved, err := client.Analytics().Rule(data.RuleName).Retrieve(t.Context())
	requireNoError(t, err)
	if retrieved.Collection != rule.Collection || retrieved.RuleTag != data.RuleTag {
		t.Fatalf("unexpected rule: %+v", retrieved)
	}
	updated, err := client.Analytics().Rule(data.RuleName).Update(t.Context(), &api.AnalyticsRuleUpdate{RuleTag: data.UpdatedRuleTag})
	requireNoError(t, err)
	if updated.RuleTag != data.UpdatedRuleTag {
		t.Fatalf("rule was not updated: %+v", updated)
	}
	list, err := rules.Retrieve(t.Context())
	requireNoError(t, err)
	if len(list) != 1 || list[0].Name != data.RuleName {
		t.Fatalf("wrong rules list: %+v", list)
	}
	event, err := client.Analytics().Events().Create(t.Context(), &api.AnalyticsEvent{Name: data.RuleName, EventType: "click", Data: api.AnalyticsEvent_Data{DocID: data.FirstID, UserID: data.UserID}})
	requireNoError(t, err)
	if !event.Ok {
		t.Fatal("event rejected")
	}
	status, err := client.Analytics().Status(t.Context())
	requireNoError(t, err)
	if status == nil {
		t.Fatal("missing analytics status")
	}
	ok, err := client.Analytics().Flush(t.Context())
	requireNoError(t, err)
	if !ok {
		t.Fatal("analytics flush failed")
	}
	// Log events can become visible asynchronously; poll with a bounded deadline.
	deadline := time.Now().Add(data.RequestTimeout)
	for {
		events, err := client.Analytics().Events().Retrieve(t.Context(), &api.GetAnalyticsEventsParams{Name: data.RuleName, UserID: data.UserID, N: 10})
		requireNoError(t, err)
		if len(events.Events) > 0 {
			found := false
			for _, event := range events.Events {
				if event.DocID == data.FirstID && event.UserID == data.UserID && event.Name == data.RuleName {
					found = true
				}
			}
			if !found {
				t.Fatalf("wrong logged events: %+v", events)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("analytics event never became visible")
		}
		select {
		case <-t.Context().Done():
			t.Fatal(t.Context().Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
	partial, err := rules.Create(t.Context(), []*api.AnalyticsRuleCreate{
		{Name: data.PartialRuleName, Collection: rule.Collection, EventType: "visit", Type: api.Log},
		{Name: data.MissingName, Collection: rule.Collection, EventType: "unsupported", Type: api.Log},
	})
	if err == nil || len(partial) != 1 || partial[0].Name != data.PartialRuleName {
		t.Fatalf("expected one success and partial error, got %+v, %v", partial, err)
	}
	cleanupDelete(t, "partial analytics rule", client.Analytics().Rule(data.PartialRuleName).Delete, 400)
	_, err = client.Analytics().Rule(data.RuleName).Delete(t.Context())
	requireNoError(t, err)
	_, err = client.Analytics().Rule(data.RuleName).Retrieve(t.Context())
	requireHTTPError(t, err, 400)
	_, err = rules.Create(t.Context(), nil)
	if err == nil {
		t.Fatal("empty rules accepted")
	}
	_, err = rules.Create(t.Context(), []*api.AnalyticsRuleCreate{nil})
	if err == nil {
		t.Fatal("nil rule accepted")
	}
	_, err = client.Analytics().Rule(data.RuleName).Update(t.Context(), nil)
	if err == nil {
		t.Fatal("nil rule update accepted")
	}
	_, err = client.Analytics().Events().Create(t.Context(), nil)
	if err == nil {
		t.Fatal("nil event accepted")
	}
}
