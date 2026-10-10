package typesense_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"typesense"
	"typesense/api"
	data "typesense/tests_data"
)

type requestDoerFunc func(*http.Request) (*http.Response, error)

func (f requestDoerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

func injectedClient(t *testing.T, doer api.HttpRequestDoer) *typesense.Client {
	t.Helper()
	generated, err := api.NewClientWithResponses("http://typesense.invalid", api.WithHTTPClient(doer))
	requireNoError(t, err)
	client, err := typesense.NewClient(typesense.WithAPIClient(generated))
	requireNoError(t, err)
	return client
}

/* These unit tests exercise error propagation without requiring a broken live server. */
func TestClientCallErrors(t *testing.T) {
	sentinel := errors.New("transport unavailable")
	calls := []struct {
		name string
		run  func(context.Context, *typesense.Client) error
	}{
		{"AliasRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Alias(data.AliasName).Retrieve(ctx)
			return err
		}},
		{"AliasDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Alias(data.AliasName).Delete(ctx)
			return err
		}},
		{"AliasesRetrieve", func(ctx context.Context, c *typesense.Client) error { _, err := c.Aliases().Retrieve(ctx); return err }},
		{"AliasesUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Aliases().Upsert(ctx, data.AliasName, &api.CollectionAliasSchema{CollectionName: data.CollectionName})
			return err
		}},
		{"AnalyticsFlush", func(ctx context.Context, c *typesense.Client) error { _, err := c.Analytics().Flush(ctx); return err }},
		{"AnalyticsStatus", func(ctx context.Context, c *typesense.Client) error { _, err := c.Analytics().Status(ctx); return err }},
		{"AnalyticsEventsCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Analytics().Events().Create(ctx, &api.AnalyticsEvent{})
			return err
		}},
		{"AnalyticsEventsRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Analytics().Events().Retrieve(ctx, &api.GetAnalyticsEventsParams{})
			return err
		}},
		{"AnalyticsRuleRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Analytics().Rule(data.RuleName).Retrieve(ctx)
			return err
		}},
		{"AnalyticsRuleDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Analytics().Rule(data.RuleName).Delete(ctx)
			return err
		}},
		{"AnalyticsRuleUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Analytics().Rule(data.RuleName).Update(ctx, &api.AnalyticsRuleUpdate{})
			return err
		}},
		{"AnalyticsRulesCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Analytics().Rules().Create(ctx, []*api.AnalyticsRuleCreate{{Name: data.RuleName}})
			return err
		}},
		{"AnalyticsRulesRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Analytics().Rules().Retrieve(ctx)
			return err
		}},
		{"CollectionsCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collections().Create(ctx, &api.CollectionSchema{Name: data.CollectionName})
			return err
		}},
		{"CollectionsFromStruct", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collections().CreateFromStruct(ctx, data.Product{})
			return err
		}},
		{"CollectionsRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collections().Retrieve(ctx, nil)
			return err
		}},
		{"CollectionRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Retrieve(ctx)
			return err
		}},
		{"CollectionUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Update(ctx, &api.CollectionUpdateSchema{})
			return err
		}},
		{"CollectionDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Delete(ctx)
			return err
		}},
		{"ConversationsCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Conversations().Models().Create(ctx, &api.ConversationModelCreateSchema{})
			return err
		}},
		{"ConversationsList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Conversations().Models().Retrieve(ctx)
			return err
		}},
		{"ConversationRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Conversations().Model(data.MissingName).Retrieve(ctx)
			return err
		}},
		{"ConversationUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Conversations().Model(data.MissingName).Update(ctx, &api.ConversationModelUpdateSchema{})
			return err
		}},
		{"ConversationDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Conversations().Model(data.MissingName).Delete(ctx)
			return err
		}},
		{"CurationsList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSets().Retrieve(ctx)
			return err
		}},
		{"CurationsUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSets().Upsert(ctx, data.CurationSetName, &api.CurationSetCreateSchema{})
			return err
		}},
		{"CurationRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Retrieve(ctx)
			return err
		}},
		{"CurationDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Delete(ctx)
			return err
		}},
		{"CurationItemsList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Items().Retrieve(ctx)
			return err
		}},
		{"CurationItemRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Item(data.SetItemID).Retrieve(ctx)
			return err
		}},
		{"CurationItemDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Item(data.SetItemID).Delete(ctx)
			return err
		}},
		{"CurationItemUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.CurationSet(data.CurationSetName).Items().Upsert(ctx, data.SetItemID, &api.CurationItemCreateSchema{})
			return err
		}},
		{"DocumentRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Document(data.FirstID).Retrieve(ctx)
			return err
		}},
		{"DocumentUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Document(data.FirstID).Update(ctx, map[string]any{}, nil)
			return err
		}},
		{"DocumentDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Document(data.FirstID).Delete(ctx)
			return err
		}},
		{"DocumentsCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().Create(ctx, data.Product{}, nil)
			return err
		}},
		{"DocumentsUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().Upsert(ctx, data.Product{}, nil)
			return err
		}},
		{"DocumentsUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().Update(ctx, map[string]any{}, nil)
			return err
		}},
		{"DocumentsDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().Delete(ctx, nil)
			return err
		}},
		{"DocumentsSearch", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().Search(ctx, &api.SearchCollectionParams{})
			return err
		}},
		{"DocumentsExport", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().Export(ctx, nil)
			return err
		}},
		{"DocumentsImport", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().Import(ctx, data.Products(), nil)
			return err
		}},
		{"DocumentsImportJsonl", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Collection[data.Product](data.CollectionName).Documents().ImportJsonl(ctx, strings.NewReader("{}\n"), nil)
			return err
		}},
		{"KeysCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Keys().Create(ctx, &api.APIKeySchema{})
			return err
		}},
		{"KeysRetrieve", func(ctx context.Context, c *typesense.Client) error { _, err := c.Keys().Retrieve(ctx); return err }},
		{"KeyRetrieve", func(ctx context.Context, c *typesense.Client) error { _, err := c.Key(1).Retrieve(ctx); return err }},
		{"KeyDelete", func(ctx context.Context, c *typesense.Client) error { _, err := c.Key(1).Delete(ctx); return err }},
		{"Metrics", func(ctx context.Context, c *typesense.Client) error { _, err := c.Metrics().Retrieve(ctx); return err }},
		{"Debug", func(ctx context.Context, c *typesense.Client) error { _, err := c.Debug(ctx); return err }},
		{"Stats", func(ctx context.Context, c *typesense.Client) error { _, err := c.Stats().Retrieve(ctx); return err }},
		{"Health", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Health(ctx, data.RequestTimeout)
			return err
		}},
		{"MultiSearchPerform", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.MultiSearch().Perform(ctx, nil, api.MultiSearchSearchesParameter{})
			return err
		}},
		{"MultiSearchCustom", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.MultiSearch().PerformWithContentType(ctx, nil, api.MultiSearchSearchesParameter{}, "application/json")
			return err
		}},
		{"MultiSearchUnion", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.MultiSearch().PerformUnion(ctx, nil, api.MultiSearchSearchesParameter{})
			return err
		}},
		{"NLCreate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.NLSearchModels().Create(ctx, &api.NLSearchModelCreateSchema{})
			return err
		}},
		{"NLList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.NLSearchModels().Retrieve(ctx)
			return err
		}},
		{"NLRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.NLSearchModel(data.MissingName).Retrieve(ctx)
			return err
		}},
		{"NLUpdate", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.NLSearchModel(data.MissingName).Update(ctx, &api.NLSearchModelUpdateSchema{})
			return err
		}},
		{"NLDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.NLSearchModel(data.MissingName).Delete(ctx)
			return err
		}},
		{"Snapshot", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Operations().Snapshot(ctx, data.SnapshotPath)
			return err
		}},
		{"Vote", func(ctx context.Context, c *typesense.Client) error { _, err := c.Operations().Vote(ctx); return err }},
		{"SlowLog", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Operations().ToggleSlowRequestLog(ctx, -1)
			return err
		}},
		{"ClearCache", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Operations().ClearCache(ctx)
			return err
		}},
		{"CompactDB", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Operations().CompactDB(ctx)
			return err
		}},
		{"SchemaChanges", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Operations().SchemaChanges(ctx)
			return err
		}},
		{"PresetsUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Presets().Upsert(ctx, data.PresetName, &api.PresetUpsertSchema{})
			return err
		}},
		{"PresetsList", func(ctx context.Context, c *typesense.Client) error { _, err := c.Presets().Retrieve(ctx); return err }},
		{"PresetRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Preset(data.PresetName).Retrieve(ctx)
			return err
		}},
		{"PresetDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Preset(data.PresetName).Delete(ctx)
			return err
		}},
		{"StemmingUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stemming().Dictionaries().Upsert(ctx, data.DictionaryID, []api.StemmingDictionaryWord{{Word: "people", Root: "person"}})
			return err
		}},
		{"StemmingJsonl", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stemming().Dictionaries().UpsertJsonl(ctx, data.DictionaryID, strings.NewReader("{}\n"))
			return err
		}},
		{"StemmingList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stemming().Dictionaries().Retrieve(ctx)
			return err
		}},
		{"StemmingRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stemming().Dictionary(data.DictionaryID).Retrieve(ctx)
			return err
		}},
		{"StopwordsUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stopwords().Upsert(ctx, data.StopwordsID, &api.StopwordsSetUpsertSchema{})
			return err
		}},
		{"StopwordsList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stopwords().Retrieve(ctx)
			return err
		}},
		{"StopwordRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stopword(data.StopwordsID).Retrieve(ctx)
			return err
		}},
		{"StopwordDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.Stopword(data.StopwordsID).Delete(ctx)
			return err
		}},
		{"SynonymsList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSets().Retrieve(ctx)
			return err
		}},
		{"SynonymsUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSets().Upsert(ctx, data.SynonymSetName, &api.SynonymSetCreateSchema{})
			return err
		}},
		{"SynonymRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Retrieve(ctx)
			return err
		}},
		{"SynonymDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Delete(ctx)
			return err
		}},
		{"SynonymItemsList", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Items().Retrieve(ctx)
			return err
		}},
		{"SynonymItemRetrieve", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Item(data.SetItemID).Retrieve(ctx)
			return err
		}},
		{"SynonymItemDelete", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Item(data.SetItemID).Delete(ctx)
			return err
		}},
		{"SynonymItemUpsert", func(ctx context.Context, c *typesense.Client) error {
			_, err := c.SynonymSet(data.SynonymSetName).Items().Upsert(ctx, data.SetItemID, &api.SynonymItemUpsertSchema{})
			return err
		}},
	}
	for _, call := range calls {
		t.Run(call.name, func(t *testing.T) {
			t.Run("HTTPError", func(t *testing.T) {
				client := injectedClient(t, requestDoerFunc(func(r *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"message":"unavailable"}`)), Request: r}, nil
				}))
				err := call.run(t.Context(), client)
				requireHTTPError(t, err, http.StatusServiceUnavailable)
			})
			t.Run("TransportError", func(t *testing.T) {
				client := injectedClient(t, requestDoerFunc(func(*http.Request) (*http.Response, error) { return nil, sentinel }))
				if err := call.run(t.Context(), client); !errors.Is(err, sentinel) {
					t.Fatalf("transport error not preserved: %v", err)
				}
			})
		})
	}
}
