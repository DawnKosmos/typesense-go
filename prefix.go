package typesense

/*
Automatic collection prefixing is currently removed. Callers supply physical
collection names explicitly in handles, schemas, references, JOIN expressions,
aliases, API key permissions, and saved configurations. Struct-inferred names
remain lowercase Go type names. Scoped keys sign parameters exactly as supplied.

When revisiting prefix support:
  - Choose an explicit namespace owner (client or collection), rather than a
    mutable package-global prefix, and define how cross-collection JOINs work.
  - Centralize name translation in the wrapper layer before sending requests
    or serializing and signing scoped keys. Copy caller-owned data before edits.
  - Writes: cover collection creation and schema updates, references, alias
    names and targets, bulk update/delete filters, key collection permissions
    (including wildcard rules), and collection names in saved presets/analytics.
  - Queries: cover collection/document/alias handles, search and multi-search
    collections, export filters, and JOIN expressions in filter_by,
    include_fields, exclude_fields, sort_by, and facet_by. Decide whether list
    operations should filter namespaces and how response names are exposed.
  - Scoped keys: normalize the same JOIN parameters before HMAC signing; keep
    ordinary field names and text values literal, including the query q.
  - A JOIN scanner must handle nested $collection(...) expressions, quoting,
    escaping, whitespace, and valid collection names. Define how already
    qualified names are recognized so translation does not double-prefix them.
  - Verify writes, queries, and signed keys against Typesense with explicit
    names, nested JOINs, quoted values, and multiple concurrent namespaces.

Typesense's search parameter named "prefix" controls text prefix matching and
is independent of collection namespaces.
*/
