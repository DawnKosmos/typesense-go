package typesense

import "typesense/api"

/* Stemming provides custom word-to-root dictionary operations. */
type Stemming struct {
	apiClient api.ClientWithResponsesInterface
}

/* Dictionaries provides dictionary import and listing operations. */
func (s *Stemming) Dictionaries() *StemmingDictionaries {
	return &StemmingDictionaries{apiClient: s.apiClient}
}

/* Dictionary selects a stemming dictionary by ID. */
func (s *Stemming) Dictionary(id string) *StemmingDictionary {
	return &StemmingDictionary{apiClient: s.apiClient, dictionaryID: id}
}
