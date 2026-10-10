// Package testsdata contains shared integration test names and fixtures.
package testsdata

import (
	"embed"
	"time"
)

const (
	ContainerImage                  = "typesense/typesense:30.2"
	APIKey                          = "integration-test-key"
	ContainerTimeout                = 2 * time.Minute
	RequestTimeout                  = 10 * time.Second
	CollectionsName                 = "collections"
	ReferenceCollectionName         = "reference_products"
	StructCollectionName            = "product"
	CollectionName                  = "collection"
	AliasCollectionName             = "alias_products"
	AliasTargetName                 = "alias_target"
	DocumentCollectionName          = "document_products"
	DocumentsCollectionName         = "documents_products"
	AnalyticsCollectionName         = "analytics_products"
	MultiSearchCollectionName       = "multisearch_products"
	ScopedCollectionName            = "scoped_products"
	SynonymCollectionName           = "synonym_products"
	StopwordCollectionName          = "stopword_products"
	StemmedCollectionName           = "stemmed_products"
	SearchKeyDescription            = "integration search key"
	CurationSetName                 = "integration_curations"
	SynonymSetName                  = "integration_synonyms"
	SetItemID                       = "integration_item"
	PresetName                      = "integration_preset"
	StopwordsID                     = "integration_stopwords"
	DictionaryID                    = "integration_dictionary"
	JsonlDictionaryID               = "integration_jsonl_dictionary"
	SnapshotPath                    = "/tmp/integration_snapshot"
	AliasName                       = "products_alias"
	RuleName                        = "product_clicks"
	PartialRuleName                 = "product_visits"
	MissingName                     = "does_not_exist"
	RuleTag                         = "integration"
	UpdatedRuleTag                  = "updated_integration"
	UserID                          = "user-1"
	FirstID                         = "product-1"
	SecondID                        = "product-2"
	InvalidID                       = "invalid-product"
	LargeInteger              int64 = 9007199254740993
)

// Product also exercises schema generation from JSON names and tyse options.
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name" tyse:"index"`
	Category string  `json:"category" tyse:"facet"`
	Price    float64 `json:"price" tyse:"index,sort"`
	Sequence int64   `json:"sequence" tyse:"index"`
	Secret   string  `json:"secret,omitempty" tyse:"optional,nostore"`
}

// Products returns a fresh fixture slice for each test.
func Products() []Product {
	return []Product{
		{ID: FirstID, Name: "Blue hiking backpack", Category: "outdoors", Price: 49.5, Sequence: LargeInteger},
		{ID: SecondID, Name: "Red coffee mug", Category: "kitchen", Price: 12.5, Sequence: 2},
	}
}

//go:embed *.jsonl
var Fixtures embed.FS
