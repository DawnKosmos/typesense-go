package testsdata

import "typesense/api"

const (
	CatalogCollection            = "catalog"
	MultiSearchCatalogCollection = "multisearch_catalog"
	CatalogPackID                = "catalog-pack"
	CatalogTentID                = "catalog-tent"
	CatalogMugID                 = "catalog-mug"
	CatalogBottleID              = "catalog-bottle"
)

type CatalogDetails struct {
	Color  string `json:"color"`
	Weight int32  `json:"weight"`
}

type CatalogItem struct {
	ID        string         `json:"id"`
	Name      string         `json:"name" tyse:"index"`
	Category  string         `json:"category" tyse:"facet"`
	Price     float64        `json:"price" tyse:"index,sort"`
	InStock   bool           `json:"in_stock" tyse:"facet"`
	Tags      []string       `json:"tags" tyse:"facet"`
	Details   CatalogDetails `json:"details" tyse:"index"`
	Location  api.GeoPoint   `json:"location" tyse:"index"`
	Embedding []float64      `json:"embedding" tyse:"index,dim:3"`
}
