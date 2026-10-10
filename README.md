# Typesense Go

**WIP — version 0.1.** This client is under active development. Its API and
generated types may change before a stable release.

Models generated from the pinned `typesense-api-spec/openapi.yml` submodule
using `oapi-codegen` v2.8.0. Models and an HTTP client are generated.

Generated types live in `api/type.gen.go`, in package `api`. They use
named inline objects, plain optional values, and acronym normalization.
All schemas are retained. Inline type names are derived from schema paths.

`api/version.gen.go` exposes `api.Version`, generated from the spec's
`info.version` (currently `"30.0"`). This identifies the pinned API specification,
not the running Typesense server or oapi-codegen version.

Optional zero values are omitted through `omitempty,omitzero` tags (Go 1.24+).
Optional booleans use the three-state type below to preserve explicit false.
Selected optional numeric fields use pointers to preserve explicit zero. Remaining
optional numeric zeros, empty strings, and empty slices/maps are omitted, except
collection-update metadata and synonym sets, which use only `omitzero` to preserve
explicit empty values. Explicit pointer types supplied through upstream Go
extensions may remain.

## Regenerate

```sh
git submodule update --init --recursive
go generate .
go build ./...
```

The upstream generator is pinned as a Go tool dependency in `go.mod` at v2.8.0.
`go tool oapi-codegen` and `go generate .` use that version.
Edit `codegen/values.yaml` to change generation options and
`codegen/models.overlay.yaml` to adjust explicit Go type references without editing the upstream spec.
The overlay also updates the spec's hardcoded `ApiKey` Go reference to `APIKey`
to match acronym normalization.

This is a playground, not a finalized public API. Query objects remain as supplied
by upstream. Inline types use the generator's schema-path naming.

The normalization step broadens free-form object maps whose value schema is
exactly `type: object` to `additionalProperties: true`. This fixes
`SearchResultHit.Document` to `map[string]any`.
This is a deliberate schema correction, not a semantics-neutral
generator option. Typed maps, structured objects, references, and constrained
value schemas remain unchanged. Currently only the document field matches.

## Three-state booleans

`optional.Bool` is a `uint8` with `None` (0), `False` (1), and `True` (2).
`codegen/generate/main.go` automatically annotates optional boolean properties
and primitive optional query/header/cookie parameters throughout the current
spec. Required booleans and boolean array/map elements retain the normal Go type.
Existing explicit upstream Go type overrides are preserved. The current spec's
boolean fields are declared inline; referenced boolean schemas or requiredness
split across composed schemas would need additional handling if introduced.

The generation command is now:

```sh
go run ./codegen/generate
```

It reads the original spec, annotates optional booleans, normalizes free-form maps,
writes a temporary spec, and runs:

```sh
go tool oapi-codegen --config codegen/values.yaml <temporary-spec>
go tool oapi-codegen --config codegen/client.yaml <temporary-spec>
```
The type-reference overlay still applies inside oapi-codegen. The temporary directory
is removed afterward. `go generate .` runs this complete pipeline; running
oapi-codegen directly against the original spec bypasses boolean annotation.

After oapi-codegen writes the models, the wrapper runs
`gofmt -w -r 'interface{} -> any' api/type.gen.go api/client.gen.go`. This syntax-aware rewrite
uses `any` consistently without changing type semantics, strings, or comments.

```go
field := api.Field{
    Name: "secret",
    Type: "string",
    Index: optional.False,
    Store: optional.False,
}
```

Generated fields have `omitempty,omitzero` tags. `None` is omitted, while
`False` and `True` serialize as JSON booleans, not integers. Standalone `None`
serializes as `null`; JSON `null` decodes to `None`. Invalid integer states
and JSON inputs other than booleans or null return errors.
Use `optional.FromBool(value)` to convert a regular boolean and `Value()`
to retrieve `(value, set)`.

## Collection field tags

`ToField(value)` builds `[]api.Field` from JSON names and `tyse` options and panics on invalid
input. `ToFieldWithName(value)` returns the lowercase struct name, fields, and
an error. `ToSchema(value)` also builds a collection schema and enables nested
fields when object fields are present. Struct values, pointers, and
`reflect.Type` are accepted.

```go
type Product struct {
    Name string `json:"name" tyse:"facet"`
    Owner string `json:"owner" tyse:"join:user.id"`
    Vector []float32 `json:"vector" tyse:"index,dim:384"`
    Location [2]float64 `json:"location" tyse:"index,type:geopoint"`
}

schema, err := typesense.ToSchema(Product{})
```

Field names come from `json`; an absent or empty JSON name defaults to the Go
field name. `tyse` contains only Typesense options, such as `tyse:"index,dim:384"`.
Fields without `tyse`, `tyse:"-"`, and `json:"-"` are skipped. An empty
`tyse:""` tag includes the field using default options.
Untagged embedded structs are flattened. Fields default to `index:false`;
`facet` also enables indexing. Other options are `optional`, `index`, `sort`,
`infix`, `nostore`, `dim:N`, `join:collection.field`,
`reference:collection.field`, and `type:TYPE`. Pointer fields require an explicit
`optional` tag. Go `int` and `int32` map to `int32`, matching the original utility.

Collection names, aliases, schema references, and JOIN expressions are used as
supplied; there is no automatic collection prefixing. Inferred schema names are
lowercase Go type names. Caller schemas and search parameters are not mutated.
Unknown options, unsupported types, invalid dimensions, and duplicate field names
produce errors.

For collection updates, `Metadata` and `SynonymSets` use only `omitzero`: nil
leaves a setting unchanged, while a non-nil empty map or slice clears it. Explicit
`x-omitempty: false` and `x-omitzero: true` overlay rules preserve this behavior
when regenerating. `Fields` retains `omitempty` because an empty field list is
rejected by Typesense.

`api/field.go` also provides Go aliases for scalar field types, including
`api.String`, `api.Int64`, `api.Float`, and `api.Object`. Use slices such as
`[]api.Float` for arrays. Named types
`api.GeoPoint`, `api.Boundary`, and `api.Image` are detected as `geopoint`,
`geopolygon`, and `image` without overrides. `api.GeoPointArray` is an alias for
`[]api.GeoPoint`. `api.StringOrArray` aliases `any` and requires an explicit
`type:string*` tag. Automatic schema detection (`auto`) is intentionally omitted.

## Scoped search keys

Build a scoped key on your backend using a parent key with only
`documents:search` permissions:

```go
key, err := typesense.NewScopedSearchKeyBuilder(parentSearchKey).
    FilterBy("tenant_id:=42 && $users(id:=user-1)").
    IncludeFields("id", "name", "price", "$users(name)").
    ExcludeFields("embedding", "private_notes", "$users(email)").
    LimitHits(200).
    ExpiresAt(time.Now().Add(time.Hour)).
    Param("highlight_fields", "name").
    Build()
```

Use the result as `X-TYPESENSE-API-KEY` for searches. JOINs require the appropriate
schema references and exact collection names in expressions. Set expiry earlier
than the parent's expiry.

The builder also provides `Query`, `QueryBy`, `SortBy`, `FacetBy`, `GroupBy`,
`Page`, `PerPage`, `UseCache`, and `CacheTTL` (seconds). `Param` accepts any
search parameter by its JSON name, preserving explicit `false`, `0`, and empty
values. Repeated setters replace earlier values, including filters and field
lists. Methods mutate the builder; create a separate builder for each scope.
`Build` can be called repeatedly and returns encoding or parent-key errors.
Typesense validates search semantics when the key is used. Parameters are signed,
not encrypted. The existing `GenerateScopedSearchKey` map API remains available.

## Client operations

The root package initializes a concrete `*api.ClientWithResponses`:

```go
client, err := typesense.NewClient(
    typesense.WithServer("http://localhost:8108"),
    typesense.WithAPIKey("your-api-key"),
)
if err != nil {
    return err
}
alias, err := client.Alias("products").Retrieve(ctx)
```

`client.go` holds initialization and defaults. `client_option.go` holds
`ClientConfig` and options for authentication, nodes, retries, timeouts, circuit
breaking, custom HTTP clients, and concrete API-client injection.
`WithClientConfig` replaces the entire configuration; later options override it.
`WithAPIClient` bypasses the built-in transport configuration. The generated
client is held in the private `c` field; its methods are not promoted onto the
root client. `Alias(name)` returns a concrete `*Alias` with `Retrieve` and `Delete`
methods returning `*api.CollectionAlias`. `Aliases()` returns a concrete
`*Aliases` with `Retrieve` (list all) and `Upsert` methods. The private API field
uses the generated `api.ClientWithResponsesInterface`. Unexpected HTTP responses return
`*typesense.HTTPError` with their status and body.

`Analytics()` returns a concrete `*Analytics`. `Events()` provides `Create` and
`Retrieve`; `Rules()` provides batch `Create` and `Retrieve`; `Rule(name)` provides
`Retrieve`, `Update`, and `Delete`. Rule creation accepts both single-object and
array success responses. If a batch contains per-rule errors, it returns the
successful rules alongside an error describing the failed entries.

`Collections()` returns a concrete `*Collections` with `Create(ctx, schema)`
and `Retrieve(ctx, params)` methods. Creation returns the HTTP 201 collection
body, and retrieval returns the collection list from HTTP 200. Unexpected HTTP
responses return `*typesense.HTTPError`.

`Collection[T](name)` selects a concrete `*Collection[T]` using Go 1.27 generic
methods, for example `client.Collection[Product]("products")`.
Use `client.Collection[map[string]any](name)` for arbitrary JSON objects.
The package function `GenericCollection[T]` remains available as a convenience.
Collections provide `Retrieve`, `Update`, and `Delete`.
Collection methods return collection metadata. `Document(id)` returns a concrete
`*Document[T]` with `Retrieve(ctx)`, `Update(ctx, document, params)`, and
`Delete(ctx)` methods decoding document bodies into `T`. Update accepts partial
JSON objects and optional `*api.DocumentIndexParameters`; nil parameters use
server defaults. HTTP failures return `*typesense.HTTPError`.

`Collection[T].Documents()` returns a concrete `*Documents` with `Create`,
`Upsert`, `Update`, `Delete`, `Search`, `Export`, `Import`, and `ImportJsonl`.
`Import(ctx, documents, params)` accepts any nonempty slice or array, including
`[]Product` and `[N]Product`, and encodes each element as a JSONL line. Each
returned `api.ImportDocumentResponse` reports that document's success or failure;
HTTP 200 does not imply that all imported documents succeeded. Decoding failures
return the results already read alongside the error. Import defaults to action
`create` and batch size 40, without modifying caller parameters. Export and
`ImportJsonl` return streams that callers must close.

The default HTTP timeout is 5 seconds, retry interval is 100 milliseconds, and
unhealthy nodes become eligible again after 1 minute. Retries default to the
number of nodes plus one when a nearest node is configured. Request bodies are
retried only when replayable through `GetBody`; streaming bodies are sent once.
The circuit breaker counts transport failures and final 5xx responses as failures,
while preserving HTTP error responses for the generated response parsers.

All generated code lives in package `api` (import `typesense/api`). Models are
in `api/type.gen.go`. Shared client configuration is in `api/client.gen.go`,
and interfaces are in `api/client_interfaces.gen.go`.

After generation, `codegen/generate/split.go` splits the client by each operation's
first OpenAPI tag. For example, collection methods, request constructors, response
wrappers, and parsers live in `api/client_collections.gen.go`; documents and
analytics have their own files. Declarations and method signatures are preserved.
The splitter removes unused imports and obsolete files bearing its generated
header. It does not split the model file.

Before splitting, `codegen/generate/containers.go` removes outer pointers from
map and array/slice response fields, including named container types. Getters
return the containers directly, and parsers assign decoded values directly.
Pointers to objects and pointers inside slices remain unchanged. A nil container
can mean JSON null or an unmatched response, so wrappers use status and content
type to recognize successful container responses.

`Client` methods return raw `*http.Response` values. `ClientWithResponses`
provides parsed response wrappers with fields for each declared status, such as
`JSON201`, `JSON400`, and `JSON409` for collection creation. HTTP error statuses
are represented in those wrappers; transport and decoding failures return Go errors.
Authentication is configured through `WithRequestEditorFn` by setting the
`X-TYPESENSE-API-KEY` header. Use `WithHTTPClient` to supply a client with a timeout.

The wrapper rewrites generated `runtime.StyleParamWithOptions` calls to
`styleParamWithOptions` in `api/query.go`. For object-valued query parameters, this
adapter first uses JSON encoding to honor omission tags and `optional.Bool`,
then lets the runtime serialize the resulting map. Other parameters use normal
runtime styling, with direct `optional.Bool` values handled as booleans.
This avoids emitting unset search fields and numeric three-state values.

`tests/client_test.go` contains the single integration entry point, `TestClient`.
It starts an isolated Typesense 30.2 container, waits for health, and runs feature
subtests from separate files. These exercise health, collections, aliases, typed
document CRUD, search, import/export, analytics, API and scoped keys, presets,
stopwords, curation sets and items, synonym sets and items, stemming, MultiSearch,
diagnostics, snapshots, leader voting, cache clearing, database compaction, and
slow request configuration. LLM model listing runs normally; creating and updating
LLM models is explicitly skipped because it requires a provider and credentials.
Shared names, IDs, typed fixtures, and embedded JSONL fixtures live in
`tests_data/`. Shared helpers load and seed fixtures, register cleanup, and check
search results and field projections. The container and its data are removed
automatically.

Organizations, users, projects, and access grants exercise forward, reverse, nested,
and many-to-many joins, including sorting and filtering across references. Scoped
keys restrict users to permitted projects and fields; tests attempt to widen those
restrictions, tamper with signatures, access other collections, and use expired or
revoked keys. Catalog fixtures cover facets, grouping, nested objects, geo filters,
and vector search with supplied embeddings. Separate collections check synonym,
stopword, and stemming behavior and searches across different collection schemas.

Unit tests live beside the code: `key_scoped_test.go` checks signatures and the
builder, and `collection_field_test.go` checks schema tags and type inference.
Client unit tests check configuration, retries, cancellation, error propagation,
stream ownership, partial JSONL decoding, and LLM wrappers using local responses.
`optional/bool_test.go` and `api/query_test.go` check JSON presence and query encoding.

Run all tests with Docker running:

```sh
go test ./...
```

Run a single feature, or run the checks that do not require Docker:

```sh
go test -v ./tests -run 'TestClient/Documents'
go test -v ./tests -run 'TestClient/(Joins|ScopedKeys|SearchFeatures)'
go test -short ./...
```

Run the race detector and record coverage for all packages:

```sh
go test -race -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

Coverage distinguishes the client runtime from generated API code and generation
tooling. The current suite covers 100% of the field parser, scoped-key generation
and builder, and optional booleans; client runtime statement coverage is 99.81%.
The two uncovered runtime statements are defensive error returns: the generated
client constructor's fixed options cannot fail, and encoding a stemming word
containing only strings into a bytes buffer cannot fail. They remain in production
code. Generated API code and generation tooling do not have 100% coverage.

Selected zero-sensitive search controls and NL sampling settings are generated
as ordinary pointers through explicit overlay overrides: `PerPage: new(0)` sends
zero; nil omits it. Remaining optional numeric values still omit zero. The models
and fields are listed in [the numeric presence audit](codegen/optional-numbers.md).

oapi-codegen warns that the client-only config references named models it does
not emit itself. Those types are provided by the separate models generation in the same
package; `go build ./...` checks that the references resolve.
