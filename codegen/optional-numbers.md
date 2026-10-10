# Optional numeric presence

Selected optional numeric fields now use ordinary pointers with `omitempty`,
configured explicitly in `models.overlay.yaml`. Nil omits the field; a pointer to
zero sends zero. Remaining optional numeric fields are values with
`omitempty,omitzero`, so assigning zero still omits those fields from JSON and
object-valued query parameters. Required numeric fields preserve zero. Numbers decoded from responses
are unaffected by request omission, although the models cannot distinguish missing
values from explicit zero when reading responses.

## Confirmed useful zero values

These search fields now use `*int` in `SearchParameters`, `MultiSearchParameters`,
and the inherited `MultiSearchCollectionParameters`:

| Field | Why zero differs from omission |
| --- | --- |
| `DropTokensThreshold` | Disables dropping words from the query. |
| `TypoTokensThreshold` | Disables typo tolerance through this threshold. |
| `FacetQueryNumTypos` | Disables typo tolerance for facet queries. |
| `PerPage` | Requests facets without document hits. |
| `Limit` | Requests zero hits with offset-based pagination. |
| `HighlightAffixNumTokens` | Requests no surrounding context tokens. |
| `SnippetThreshold` | Sets the full-field highlighting threshold to zero. |
| `MaxExtraPrefix`, `MaxExtraSuffix` | Allows zero extra characters in infix matches, rather than the unlimited default. |

The generated NL model types `NLSearchModelCreateSchema` and
`NLSearchModelUpdateSchema` (an alias of the create schema), as well as the
`NLSearchModelBase` component, now preserve `Temperature: 0` and `TopP: 0` through
`*float32` fields. The inherited response model also receives these pointers. Zero is a
supported sampling value; the behavior of an omitted parameter depends on the
configured model/provider. `TopK: 0` needs provider-specific validation.

Sources: [Search parameters](https://typesense.org/docs/30.2/api/search.html),
[Natural language models](https://typesense.org/docs/30.2/api/natural-language-search.html).

## Other request models with optional numeric fields

The following fields also lose explicit zero. This inventory does not imply that
the server accepts zero for every field: many values must be positive, use zero
as their default, or depend on the endpoint/provider.

| Model | Fields |
| --- | --- |
| `AnalyticsRuleCreate_Params`, `AnalyticsRuleUpdate_Params` | `Limit`, `Weight` |
| `APIKeySchema` | `ExpiresAt` |
| `ConversationModelCreateSchema` | `TTL` |
| `ConversationModelUpdateSchema` | `TTL`, `MaxBytes` |
| `CurationItemCreateSchema` | `EffectiveFromTS`, `EffectiveToTS` |
| `Field` | `NumDim` |
| `NLSearchModelCreateSchema`, `NLSearchModelUpdateSchema`, `NLSearchModelBase` | `MaxBytes`, `MaxOutputTokens`, `TopK` (in addition to the sampling fields above) |
| `GetCollectionsParams_GetCollectionsParameters` | `Limit`, `Offset` |
| `DeleteDocumentsParams_DeleteDocumentsParameters` | `BatchSize` |
| `ImportDocumentsParams_ImportDocumentsParameters` | `BatchSize`, `RemoteEmbeddingBatchSize` |

The three search models also omit zero for `CacheTTL`, `FacetSamplePercent`,
`FacetSampleSlope`, `FacetSampleThreshold`, `GroupLimit`, `GroupMaxCandidates`,
`LimitHits`, `MaxFacetValues`, `MaxFilterByCandidates`, `MinLen1Typo`,
`MinLen2Typo`, `Offset`, `Page`, `RemoteEmbeddingNumTries`,
`RemoteEmbeddingTimeoutMs`, `SearchCutoffMs`, and `SynonymNumTypos`.
`SearchParameters` additionally has `MaxCandidates`. Even where zero equals the
server default, omission cannot explicitly override a nonzero preset value.

`NumTypos` is a string, so setting it to `"0"` works with the current generation.
Scoped-key parameters use a map and preserve numeric zero, including values set
through the builder. `PresetUpsertSchema.Value` is a generated union: constructing
it with `FromSearchParameters` or `FromMultiSearchSearchesParameter` now preserves
zero for the selected pointer fields, but still loses zero for the remaining
value fields when those search structs are marshaled. Loading explicit raw JSON through
the union's `UnmarshalJSON` method preserves zero.

## Smallest correction

The selected fields retain ordinary `*int` / `*float32` through
`x-go-type-skip-optional-pointer: false` in the overlay. There is
no need for a custom optional numeric type or to restore pointers on every numeric
field. Removing omission tags instead would send zeros for all unset fields and
replace useful server defaults.

Regenerate with `go generate ./...`. For example, `PerPage: new(0)` sends zero
while leaving `PerPage` nil preserves the server default.
