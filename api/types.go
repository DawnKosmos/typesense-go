package api

/* StemmingDictionaryWord maps a word to its root form. */
type StemmingDictionaryWord = StemmingDictionary_Words

// MarshalJSON preserves the union encoder on the generated request body type.
// A new Go type does not inherit methods from CreateAnalyticsRuleJSONBody.
func (body CreateAnalyticsRuleJSONRequestBody) MarshalJSON() ([]byte, error) {
	return CreateAnalyticsRuleJSONBody(body).MarshalJSON()
}

func (body *CreateAnalyticsRuleJSONRequestBody) UnmarshalJSON(data []byte) error {
	return (*CreateAnalyticsRuleJSONBody)(body).UnmarshalJSON(data)
}

// DocumentIndexParameters provides shared options for individual document writes.
type DocumentIndexParameters struct {
	DirtyValues *DirtyValues `json:"dirty_values,omitempty"`
}

// ImportDocumentResponse describes one result in a document import's JSONL response.
type ImportDocumentResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Document any    `json:"document,omitempty"`
	ID       string `json:"id,omitempty"`
}
