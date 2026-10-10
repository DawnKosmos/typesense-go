package typesense

import (
	"slices"
	"strings"
	"unicode"

	"typesense/api"
)

/* Prefix references in a copy so creating and updating schemas share the same behavior. */
func prefixFields(fields []api.Field) []api.Field {
	fields = slices.Clone(fields)
	for i := range fields {
		if fields[i].Reference != "" {
			fields[i].Reference = prefixReference(fields[i].Reference)
		}
	}
	return fields
}

/* Prefix JOIN collection names, including nested JOINs, while preserving quoted values. */
func prefixJoins(expression string) string {
	if _prefix == "" || !strings.Contains(expression, "$") {
		return expression
	}
	var result strings.Builder
	var quote byte
	for i := 0; i < len(expression); {
		ch := expression[i]
		if ch == '\\' && i+1 < len(expression) {
			result.WriteString(expression[i : i+2])
			i += 2
			continue
		}
		if quote != 0 {
			if ch == quote {
				quote = 0
			}
		} else if ch == '`' || ch == '"' {
			quote = ch
		} else if ch == '$' {
			end := i + 1
			for end < len(expression) && joinNameByte(expression[end]) {
				end++
			}
			opening := end
			for opening < len(expression) && unicode.IsSpace(rune(expression[opening])) {
				opening++
			}
			if end > i+1 && opening < len(expression) && expression[opening] == '(' {
				result.WriteByte('$')
				result.WriteString(Prefix(expression[i+1 : end]))
				i = end
				continue
			}
		}
		result.WriteByte(ch)
		i++
	}
	return result.String()
}

func joinNameByte(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '-' || ch == '.' || ch >= 128
}

/* Only parameters containing JOIN expressions are rewritten; search text stays literal. */
func prefixSearchParam(name string, value any) any {
	switch name {
	case "filter_by", "include_fields", "exclude_fields", "sort_by", "facet_by":
		if expression, ok := value.(string); ok {
			return prefixJoins(expression)
		}
	}
	return value
}

func prefixSearch(params api.SearchParameters) api.SearchParameters {
	params.FilterBy = prefixJoins(params.FilterBy)
	params.IncludeFields = prefixJoins(params.IncludeFields)
	params.ExcludeFields = prefixJoins(params.ExcludeFields)
	params.SortBy = prefixJoins(params.SortBy)
	params.FacetBy = prefixJoins(params.FacetBy)
	return params
}

func prefixMultiSearch(common *api.MultiSearchParams, body api.MultiSearchSearchesParameter) (*api.MultiSearchParams, api.MultiSearchSearchesParameter) {
	if common != nil {
		copy := *common
		copy.MultiSearchParameters.FilterBy = prefixJoins(copy.MultiSearchParameters.FilterBy)
		copy.MultiSearchParameters.IncludeFields = prefixJoins(copy.MultiSearchParameters.IncludeFields)
		copy.MultiSearchParameters.ExcludeFields = prefixJoins(copy.MultiSearchParameters.ExcludeFields)
		copy.MultiSearchParameters.SortBy = prefixJoins(copy.MultiSearchParameters.SortBy)
		copy.MultiSearchParameters.FacetBy = prefixJoins(copy.MultiSearchParameters.FacetBy)
		common = &copy
	}
	body.Searches = slices.Clone(body.Searches)
	for i := range body.Searches {
		search := &body.Searches[i]
		search.Collection = Prefix(search.Collection)
		search.FilterBy = prefixJoins(search.FilterBy)
		search.IncludeFields = prefixJoins(search.IncludeFields)
		search.ExcludeFields = prefixJoins(search.ExcludeFields)
		search.SortBy = prefixJoins(search.SortBy)
		search.FacetBy = prefixJoins(search.FacetBy)
	}
	return common, body
}
