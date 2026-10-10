package typesense_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

func TestClientNilInputs(t *testing.T) {
	client := injectedClient(t, requestDoerFunc(func(*http.Request) (*http.Response, error) {
		t.Error("invalid input made a server request")
		return nil, nil
	}))
	for _, tc := range []struct {
		name string
		run  func(context.Context, *typesense.Client) error
	}{
		{"ConversationCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Conversations().Models().Create(ctx, nil)
			return err
		}},
		{"ConversationUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Conversations().Model(data.MissingName).Update(ctx, nil)
			return err
		}},
		{"NLCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.NLSearchModels().Create(ctx, nil)
			return err
		}},
		{"NLUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.NLSearchModel(data.MissingName).Update(ctx, nil)
			return err
		}},
		{"CurationSet", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Upsert(ctx, nil)
			return err
		}},
		{"CurationItem", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Item(data.SetItemID).Upsert(ctx, nil)
			return err
		}},
		{"SynonymSet", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Upsert(ctx, nil)
			return err
		}},
		{"SynonymItem", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Item(data.SetItemID).Upsert(ctx, nil)
			return err
		}},
		{"Keys", func(ctx context.Context, c *typesense.Client) error { _, err := c.Keys().Create(ctx, nil); return err }},
		{"Presets", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Presets().Upsert(ctx, data.PresetName, nil)
			return err
		}},
		{"Stopwords", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stopwords().Upsert(ctx, data.StopwordsID, nil)
			return err
		}},
		{"StemmingWords", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stemming().Dictionaries().Upsert(ctx, data.DictionaryID, nil)
			return err
		}},
		{"StemmingReader", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stemming().Dictionaries().UpsertJsonl(ctx, data.DictionaryID, nil)
			return err
		}},
		{"InvalidSchema", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collections().CreateFromStruct(ctx, 123)
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(t.Context(), client); err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

/* LLM wrappers are tested locally; no provider is contacted or configured. */
func TestLLMModelResponses(t *testing.T) {
	client := injectedClient(t, requestDoerFunc(func(r *http.Request) (*http.Response, error) {
		status := http.StatusOK
		if r.Method == http.MethodPost {
			status = http.StatusCreated
		}
		var payload any = map[string]any{"id": data.SetItemID, "model_name": "openai/example"}
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/models") {
			payload = []any{payload}
		}
		body, err := json.Marshal(payload)
		requireNoError(t, err)
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(string(body))), Request: r}, nil
	}))
	ctx := t.Context()
	created, err := client.Conversations().Models().Create(ctx, &api.ConversationModelCreateSchema{ModelName: "openai/example"})
	requireNoError(t, err)
	if created.ID != data.SetItemID {
		t.Fatalf("wrong model: %+v", created)
	}
	retrieved, err := client.Conversations().Model(data.SetItemID).Retrieve(ctx)
	requireNoError(t, err)
	if retrieved.ID != created.ID {
		t.Fatal("wrong conversation model")
	}
	updated, err := client.Conversations().Model(data.SetItemID).Update(ctx, &api.ConversationModelUpdateSchema{ModelName: "openai/example"})
	requireNoError(t, err)
	if updated.ID != created.ID {
		t.Fatal("wrong updated conversation model")
	}
	deleted, err := client.Conversations().Model(data.SetItemID).Delete(ctx)
	requireNoError(t, err)
	if deleted.ID != created.ID {
		t.Fatal("wrong deleted conversation model")
	}
	nl, err := client.NLSearchModels().Create(ctx, &api.NLSearchModelCreateSchema{ModelName: "openai/example"})
	requireNoError(t, err)
	if nl.ID != data.SetItemID {
		t.Fatal("wrong NL model")
	}
	nl, err = client.NLSearchModel(data.SetItemID).Retrieve(ctx)
	requireNoError(t, err)
	if nl.ID != data.SetItemID {
		t.Fatal("wrong retrieved NL model")
	}
	nl, err = client.NLSearchModel(data.SetItemID).Update(ctx, &api.NLSearchModelUpdateSchema{ModelName: "openai/example"})
	requireNoError(t, err)
	if nl.ID != data.SetItemID {
		t.Fatal("wrong updated NL model")
	}
	removed, err := client.NLSearchModel(data.SetItemID).Delete(ctx)
	requireNoError(t, err)
	if removed.ID != data.SetItemID {
		t.Fatal("wrong deleted NL model")
	}
}

func TestHTTPErrorMessage(t *testing.T) {
	err := &typesense.HTTPError{Status: 409, Body: []byte("already exists")}
	if err.Error() != "status: 409 response: already exists" {
		t.Fatalf("wrong error message: %s", err)
	}
}
