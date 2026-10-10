package tests

import (
	"testing"
	"typesense"
	data "typesense/tests_data"
)

func diagnostics(t *testing.T, client *typesense.Client) {
	debug, err := client.Debug(t.Context())
	requireNoError(t, err)
	if debug.Version == "" {
		t.Fatal("server version missing")
	}
	metrics, err := client.Metrics().Retrieve(t.Context())
	requireNoError(t, err)
	if len(metrics) == 0 {
		t.Fatal("server metrics missing")
	}
	stats, err := client.Stats().Retrieve(t.Context())
	requireNoError(t, err)
	if stats == nil {
		t.Fatal("server statistics missing")
	}
}

func llmModels(t *testing.T, client *typesense.Client) {
	_, err := client.Conversations().Models().Retrieve(t.Context())
	requireNoError(t, err)
	_, err = client.NLSearchModels().Retrieve(t.Context())
	requireNoError(t, err)
	t.Run("ConversationModelCRUD", func(t *testing.T) {
		t.Skip("requires an LLM provider and credentials; wrapper behavior is covered by unit tests")
	})
	t.Run("NLSearchModelCRUD", func(t *testing.T) {
		t.Skip("requires an LLM provider and credentials; wrapper behavior is covered by unit tests")
	})
}

func operations(t *testing.T, client *typesense.Client) {
	for _, threshold := range []int{100, -1} {
		ok, err := client.Operations().ToggleSlowRequestLog(t.Context(), threshold)
		requireNoError(t, err)
		if !ok {
			t.Fatal("slow request configuration failed")
		}
	}
	ok, err := client.Operations().ClearCache(t.Context())
	requireNoError(t, err)
	if !ok {
		t.Fatal("cache clear failed")
	}
	ok, err = client.Operations().CompactDB(t.Context())
	requireNoError(t, err)
	if !ok {
		t.Fatal("compaction failed")
	}
	_, err = client.Operations().SchemaChanges(t.Context())
	requireNoError(t, err)
	ok, err = client.Operations().Snapshot(t.Context(), data.SnapshotPath)
	requireNoError(t, err)
	if !ok {
		t.Fatal("snapshot failed")
	}
	_, err = client.Operations().Vote(t.Context())
	requireNoError(t, err)
}
