package tests

import (
	"testing"
	"typesense"
)

func TestClient(t *testing.T) {
	if testing.Short() {
		t.Log("Skipping container integration tests in short mode")
		return
	}
	client, server := startClient(t)
	for _, feature := range []struct {
		name string
		run  func(*testing.T, *typesense.Client)
	}{
		{"Health", health},
		{"Collections", collections},
		{"Collection", collection},
		{"Aliases", aliases},
		{"Document", document},
		{"Documents", documents},
		{"Analytics", analytics},
		{"Keys", keys},
		{"Joins", joins},
		{"ScopedKeys", func(t *testing.T, client *typesense.Client) { scopedKeys(t, client, server) }},
		{"SearchFeatures", searchFeatures},
		{"Presets", presets},
		{"Stopwords", stopwords},
		{"SynonymSets", synonymSets},
		{"CurationSets", curationSets},
		{"Stemming", stemming},
		{"MultiSearch", multiSearch},
		{"Diagnostics", diagnostics},
		{"LLMModels", llmModels},
		{"Operations", operations},
	} {
		t.Run(feature.name, func(t *testing.T) { feature.run(t, client) })
	}
}
