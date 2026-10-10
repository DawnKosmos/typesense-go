package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"typesense/api"
)

/* StemmingDictionaries provides dictionary import and listing operations. */
type StemmingDictionaries struct {
	apiClient api.ClientWithResponsesInterface
}

/* Upsert imports word-to-root mappings and returns the decoded JSONL results. */
func (s *StemmingDictionaries) Upsert(ctx context.Context, id string, words []api.StemmingDictionaryWord) ([]*api.StemmingDictionaryWord, error) {
	if len(words) == 0 {
		return nil, fmt.Errorf("typesense: stemming dictionary words must not be empty")
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	for i, word := range words {
		if err := encoder.Encode(word); err != nil {
			return nil, fmt.Errorf("typesense: encode stemming word at index %d: %w", i, err)
		}
	}
	response, err := s.UpsertJsonl(ctx, id, &buf)
	if err != nil {
		return nil, err
	}
	defer response.Close()
	var result []*api.StemmingDictionaryWord
	decoder := json.NewDecoder(response)
	for {
		var word *api.StemmingDictionaryWord
		err := decoder.Decode(&word)
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return result, fmt.Errorf("typesense: decode stemming word at index %d: %w", len(result), err)
		}
		if word == nil {
			return result, fmt.Errorf("typesense: null stemming word at index %d", len(result))
		}
		result = append(result, word)
	}
}

/* UpsertJsonl imports JSONL mappings and returns a stream the caller must close. */
func (s *StemmingDictionaries) UpsertJsonl(ctx context.Context, id string, body io.Reader) (io.ReadCloser, error) {
	if body == nil {
		return nil, fmt.Errorf("typesense: stemming dictionary body must not be nil")
	}
	raw, ok := s.apiClient.(api.ClientInterface)
	if !ok {
		return nil, fmt.Errorf("typesense: API client does not support raw stemming dictionary import")
	}
	response, err := raw.ImportStemmingDictionaryWithBody(ctx, &api.ImportStemmingDictionaryParams{ID: id}, "application/octet-stream", body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("typesense: read HTTP %d response: %w", response.StatusCode, err)
		}
		return nil, &HTTPError{Status: response.StatusCode, Body: data}
	}
	return response.Body, nil
}

/* Retrieve lists the IDs of available stemming dictionaries. */
func (s *StemmingDictionaries) Retrieve(ctx context.Context) ([]string, error) {
	response, err := s.apiClient.ListStemmingDictionariesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.Dictionaries, nil
}
