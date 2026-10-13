# Typesense Go

A Go client that stays close to the Typesense API specification while making the
Go types and everyday document operations easier to use.

**WIP — version 0.1.** The API and generated types may change before a stable
release. The project currently requires Go 1.27.1, as declared in `go.mod`.

## Why this repository exists

Generating a client from OpenAPI gives us broad API coverage, but the generated
Go code still needs to be practical to develop against. This repository focuses
on four improvements:

1. **A direct link to the upstream specification.** The Typesense API spec is a
   pinned Git submodule. Updating the spec and regenerating the client happen in
   this repository, without manually copying specification files. Local
   configuration and overlays make adjustments repeatable while keeping the
   upstream spec intact.
2. **Less pointer handling in generated code.** Optional fields and containers
   such as `*map[string]any` and `*[]T` add allocation, dereferencing, and nil
   checks to ordinary work. Generation favors plain values, maps, and slices;
   optional booleans use `optional.Bool` to distinguish unset from explicit
   `false`. Pointers remain where they are needed to preserve meaningful values,
   such as an explicitly supplied numeric zero.
3. **Collection schemas defined alongside Go documents.** `tyse` tags describe
   indexing, faceting, references, vectors, and other field options on your
   structs. Helpers build Typesense collection schemas from those structs at
   runtime. Named types such as `api.GeoPoint` supply the information needed to
   infer geographic and image field types without repeating type overrides.
4. **Room for practical client improvements.** A wrapper around the generated
   API makes common operations easier to use. Import already accepts typed
   slices and arrays, writes JSONL, and returns individual document results.
   Further adjustments can live here while the generated API continues to
   follow the spec.

## Getting started

The examples use the current module paths: `typesense`, `typesense/api`, and
`typesense/optional`.

```go
client, err := typesense.NewClient(
    typesense.WithServer("http://localhost:8108"),
    typesense.WithAPIKey("your-api-key"),
)
if err != nil {
    return err
}

type Product struct {
    ID       string       `json:"id" tyse:"index"`
    Name     string       `json:"name" tyse:"index"`
    Category string       `json:"category" tyse:"facet"`
    Location api.GeoPoint `json:"location" tyse:"index"`
}

schema, err := typesense.ToSchema(Product{})
if err != nil {
    return err
}
schema.Name = "products"
if _, err := client.Collections().Create(ctx, schema); err != nil {
    return err
}

products := client.Collection[Product]("products")
_, err = products.Documents().Create(ctx, Product{
    ID:       "1",
    Name:     "Backpack",
    Category: "outdoors",
    Location: api.GeoPoint{52.5200, 13.4050},
}, nil)
```

`Collection[T](name)` selects the document type used by `Document(id).Retrieve`,
`Update`, and `Delete`. `Documents()` provides create, upsert, bulk update/delete,
search, import, and export operations. Collection metadata operations are
`Retrieve`, `Update`, and `Delete`.

## Go types with less pointer handling

The generation configuration skips optional pointers by default. A postprocessing
step also removes outer pointers from map and slice/array response fields,
including named container types, and adjusts their getters and parsers.
Object pointers and pointers to elements inside slices remain unchanged.

Optional booleans retain presence without requiring a `*bool`:

```go
field := api.Field{
    Name:  "secret",
    Type:  string(api.STRING),
    Index: optional.False,
    Store: optional.False,
}
```

| Value | Meaning in an optional JSON field |
| --- | --- |
| `optional.None` | Omit the field. |
| `optional.False` | Send `false`. |
| `optional.True` | Send `true`. |

`optional.Bool` serializes as a JSON boolean, not its underlying integer.
Standalone `None` serializes as `null`, and JSON `null` decodes to `None`.
Use `optional.FromBool(value)` to convert a regular boolean and `Value()` to
retrieve `(value, set)`. Required booleans and boolean array/map elements remain
ordinary Go booleans.

Presence still matters. Most optional values use `omitempty,omitzero`, so zero
numbers, empty strings, and empty containers are omitted. Selected numeric fields
keep pointers: for example, `PerPage: new(0)` sends zero, while nil omits it.
See [the numeric presence audit](codegen/optional-numbers.md) for the fields and
remaining limitations. Explicit upstream Go type overrides may also retain
pointers.

Collection-update `Metadata` and `SynonymSets` use only `omitzero`: nil leaves
settings unchanged, while a non-nil empty map or slice explicitly clears them.
Object-valued query parameters honor these omission rules and `optional.Bool`
through the adapter in [api/query.go](api/query.go).

## Collection schemas from `tyse` tags

The tags generate collection schemas from your application types; they are
separate from the OpenAPI generation that produces the HTTP client and API models.

```go
type Place struct {
    Name     string            `json:"name" tyse:"index"`
    Category string            `json:"category" tyse:"facet"`
    OwnerID  string            `json:"owner_id" tyse:"index,join:users.id"`
    Vector   []float32         `json:"vector" tyse:"index,dim:384"`
    Location api.GeoPoint      `json:"location" tyse:"index"`
    Areas    api.Boundary      `json:"areas" tyse:"index"`
    Image    api.Image         `json:"image" tyse:"optional"`
    Nearby   api.GeoPointArray `json:"nearby" tyse:"index"`
}

schema, err := typesense.ToSchema(Place{})
```

`api.GeoPoint` contains `[latitude, longitude]`; `api.GeoPointArray` is a slice of
points. `api.Boundary` contains alternating latitude and longitude values for a
polygon. `api.Image` holds a base64-encoded image. Their named types let schema
inference select `geopoint`, `geopoint[]`, `geopolygon`, and `image`. For a plain
array, declare the meaning explicitly:

```go
Location [2]float64 `json:"location" tyse:"index,type:geopoint"`
```

| Helper | Result |
| --- | --- |
| `ToField(value)` | `[]api.Field`; panics on invalid input. |
| `ToFieldWithName(value)` | Lowercase struct name, fields, and an error. |
| `ToSchema(value)` | Collection schema and an error; enables nested fields for objects. |
| `client.Collections().CreateFromStruct(ctx, value)` | Infers and creates a collection from a named struct. |

Struct values, pointers, and `reflect.Type` are accepted. Anonymous structs need
an explicit collection name when using `ToSchema`.

Field names come from `json`, defaulting to the Go field name when absent or empty.
Fields without `tyse`, with `tyse:"-"`, or with `json:"-"` are skipped. An empty
`tyse:""` includes the field with default options. Untagged embedded structs are
flattened. Fields default to `index:false`; `facet` also enables indexing.

Supported options are `index`, `facet`, `optional`, `sort`, `infix`, `nostore`,
`dim:N`, `join:collection.field`, `reference:collection.field`, and `type:TYPE`.
Pointer fields require an explicit `optional` tag. Unsupported types, unknown
options, invalid dimensions, and duplicate field names produce errors.

[api/field.go](api/field.go) also supplies scalar aliases and field-type constants.
`api.StringOrArray` requires `tyse:"type:string*"`; automatic schema detection
(`auto`) is intentionally omitted. Collection names, aliases, references, and
JOIN expressions are used as supplied. Automatic collection prefixing is removed;
[prefix.go](prefix.go) records the design considerations for revisiting it.

## Import and other client improvements

Import accepts a nonempty typed slice or array directly:

```go
results, err := products.Documents().Import(ctx, []Product{
    {ID: "1", Name: "Backpack", Category: "outdoors"},
    {ID: "2", Name: "Mug", Category: "kitchen"},
}, nil)
```

The wrapper encodes each document as a JSONL line and decodes individual
`api.ImportDocumentResponse` results. HTTP 200 does not mean every document
succeeded; inspect each result. If response decoding fails, the method returns
results already decoded alongside the error. Defaults are action `create` and
batch size 40, applied without modifying caller parameters.

`Import` buffers the encoded request in memory. For existing JSONL or larger
inputs, `ImportJsonl` accepts an `io.Reader` and returns the raw response stream.
Both that stream and the stream returned by `Export` must be closed by the caller.

The wrapper also provides:

- Scoped search-key generation and a builder for signed filters, projections,
  expiry, and other search parameters.
- Separate and union multi-search results.
- Collection aliases, analytics, API keys, presets, synonyms, stopwords,
  curation sets, stemming, model configuration, and server operations.
- Configurable authentication, nodes, retries, timeouts, circuit breaking,
  custom HTTP clients, and API-client injection.

Unexpected HTTP responses return `*typesense.HTTPError` with their status and body.
The default HTTP timeout is 5 seconds. Streaming requests are sent once unless
an HTTP request body supports replay through `GetBody`.

### Scoped search keys

Create keys on your backend using a parent key with `documents:search` permissions:

```go
key, err := typesense.NewScopedSearchKeyBuilder(parentSearchKey).
    FilterBy("tenant_id:=42 && $users(id:=user-1)").
    IncludeFields("id", "name", "$users(name)").
    ExcludeFields("embedding", "private_notes").
    LimitHits(200).
    ExpiresAt(time.Now().Add(time.Hour)).
    Build()
```

Use the result as `X-TYPESENSE-API-KEY`. JOINs need matching schema references and
exact collection names. Set expiry earlier than the parent's expiry. Parameters
are signed, not encrypted, and Typesense validates their meaning when searching.
The builder's `Param` method and the map-based `GenerateScopedSearchKey` API
preserve explicit `false`, zero, and empty values.

## Regenerating from the upstream spec

```sh
git submodule update --init --recursive
go generate .
go build ./...
```

The submodule pins the input spec; `go.mod` pins `oapi-codegen` v2.8.0 as a Go tool.
To target another spec revision, update the submodule checkout, regenerate, and
commit the submodule revision together with the generated changes.

`go generate .` runs [codegen/generate](codegen/generate/main.go), which:

1. Reads `typesense-api-spec/openapi.yml` and annotates optional booleans.
2. Normalizes free-form object-valued maps so document payloads can use
   `map[string]any`, then writes a temporary spec.
3. Generates models and the HTTP client using the local configs and overlay.
4. Adapts query encoding, removes outer response-container pointers, uses `any`
   consistently, and splits the generated client by OpenAPI operation tag.
5. Generates `api.Version` from the spec's `info.version` and removes the
   temporary spec.

All schemas are retained. Inline objects receive named types derived from their
schema paths; acronyms are normalized. The free-form map normalization is a
local schema correction; typed and constrained map values are preserved. The
original submodule spec is not modified by generation.

| File | Purpose |
| --- | --- |
| [codegen/values.yaml](codegen/values.yaml) | Model generation options. |
| [codegen/client.yaml](codegen/client.yaml) | HTTP client generation options. |
| [codegen/models.overlay.yaml](codegen/models.overlay.yaml) | Local type references, presence rules, and response compatibility fixes. |
| [api/type.gen.go](api/type.gen.go) | Generated API models. |
| [api/client.gen.go](api/client.gen.go) | Shared generated client configuration. |
| [api/client_interfaces.gen.go](api/client_interfaces.gen.go) | Generated client interfaces. |

Operation-specific client files live beside them in `api/`. `api.Version`
identifies the pinned specification, rather than the running server. Running
`oapi-codegen` directly on the original spec bypasses the local pipeline.

## Testing

Run unit tests without Docker:

```sh
go test -short ./...
```

Run the full suite with Docker running, or select an integration feature:

```sh
go test ./...
go test -v ./tests -run 'TestClient/(Joins|ScopedKeys|SearchFeatures)'
```

The integration suite starts an isolated Typesense 30.2 container and removes it
afterward. It exercises document CRUD, import/export, collection schemas,
forward/reverse/nested JOINs, scoped-key restrictions, search features, and other
client operations. Creating and updating LLM models is skipped because it needs
provider credentials. Unit tests cover serialization, schema inference,
scoped-key signing, transport behavior, errors, and partial import responses.

To record coverage and check for races:

```sh
go test -race -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
