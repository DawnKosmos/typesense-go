package typesense

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"typesense/api"
)

const (
	defaultImportBatchSize = 40
)

// Documents provides operations on documents within a collection.
type Documents struct {
	apiClient      api.ClientWithResponsesInterface
	collectionName string
}

func (d *Documents) indexDocument(ctx context.Context, document any, params *api.IndexDocumentParams) (map[string]any, error) {
	response, err := d.apiClient.IndexDocumentWithResponse(ctx,
		d.collectionName, params, document)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 201 || !strings.Contains(response.ContentType(), "json") {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON201, nil
}

func (d *Documents) Create(ctx context.Context, document any, params *api.DocumentIndexParameters) (map[string]any, error) {
	return d.indexDocument(ctx, document, &api.IndexDocumentParams{Action: api.Create, DirtyValues: dirtyValues(params)})
}

func (d *Documents) Update(ctx context.Context, updateFields any, params *api.UpdateDocumentsParams) (int, error) {
	response, err := d.apiClient.UpdateDocumentsWithResponse(ctx,
		d.collectionName, params, updateFields)
	if err != nil {
		return 0, err
	}
	if response.JSON200 == nil {
		return 0, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.NumUpdated, nil
}

func (d *Documents) Upsert(ctx context.Context, document any, params *api.DocumentIndexParameters) (map[string]any, error) {
	return d.indexDocument(ctx, document, &api.IndexDocumentParams{Action: api.Upsert, DirtyValues: dirtyValues(params)})
}

func (d *Documents) Delete(ctx context.Context, filter *api.DeleteDocumentsParams) (int, error) {
	response, err := d.apiClient.DeleteDocumentsWithResponse(ctx,
		d.collectionName, filter)
	if err != nil {
		return 0, err
	}
	if response.JSON200 == nil {
		return 0, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200.NumDeleted, nil
}

func (d *Documents) Search(ctx context.Context, params *api.SearchCollectionParams) (*api.SearchResult, error) {
	if params == nil {
		return nil, fmt.Errorf("typesense: search parameters must not be nil")
	}
	options := *params
	options.SearchParameters = prefixSearch(options.SearchParameters)
	response, err := d.apiClient.SearchCollectionWithResponse(ctx,
		d.collectionName, &options)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, &HTTPError{Status: response.StatusCode(), Body: response.Body}
	}
	return response.JSON200, nil
}

func (d *Documents) Export(ctx context.Context, params *api.ExportDocumentsParams) (io.ReadCloser, error) {
	raw, ok := d.apiClient.(api.ClientInterface)
	if !ok {
		return nil, fmt.Errorf("typesense: API client does not support raw document export")
	}
	response, err := raw.ExportDocuments(ctx, d.collectionName, params)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return nil, fmt.Errorf("typesense: read HTTP %d response: %w", response.StatusCode, readErr)
		}
		return nil, &HTTPError{Status: response.StatusCode, Body: body}
	}
	return response.Body, nil
}

// importParams applies defaults to a copy without modifying caller options.
func importParams(params *api.ImportDocumentsParams_ImportDocumentsParameters) *api.ImportDocumentsParams {
	options := api.ImportDocumentsParams_ImportDocumentsParameters{}
	if params != nil {
		options = *params
	}
	if options.BatchSize == 0 {
		options.BatchSize = defaultImportBatchSize
	}
	if options.Action == "" {
		options.Action = api.Create
	}
	return &api.ImportDocumentsParams{ImportDocumentsParameters: &options}
}

func dirtyValues(params *api.DocumentIndexParameters) api.DirtyValues {
	if params != nil && params.DirtyValues != nil {
		return *params.DirtyValues
	}
	return ""
}

func (d *Documents) Import(ctx context.Context, documents any, params *api.ImportDocumentsParams_ImportDocumentsParameters) ([]*api.ImportDocumentResponse, error) {
	values := reflect.ValueOf(documents)
	if !values.IsValid() || (values.Kind() != reflect.Slice && values.Kind() != reflect.Array) {
		return nil, fmt.Errorf("typesense: documents must be a slice or array")
	}
	if values.Len() == 0 {
		return nil, fmt.Errorf("typesense: documents list is empty")
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	for i := 0; i < values.Len(); i++ {
		if err := encoder.Encode(values.Index(i).Interface()); err != nil {
			return nil, fmt.Errorf("typesense: encode document at index %d: %w", i, err)
		}
	}

	response, err := d.ImportJsonl(ctx, &buf, params)
	if err != nil {
		return nil, err
	}
	defer response.Close()

	var result []*api.ImportDocumentResponse
	decoder := json.NewDecoder(response)
	for {
		var item *api.ImportDocumentResponse
		err := decoder.Decode(&item)
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return result, fmt.Errorf("typesense: decode import result at index %d: %w", len(result), err)
		}
		if item == nil {
			return result, fmt.Errorf("typesense: null import result at index %d", len(result))
		}
		result = append(result, item)
	}

}

func (d *Documents) ImportJsonl(ctx context.Context, body io.Reader, params *api.ImportDocumentsParams_ImportDocumentsParameters) (io.ReadCloser, error) {
	raw, ok := d.apiClient.(api.ClientInterface)
	if !ok {
		return nil, fmt.Errorf("typesense: API client does not support raw document import")
	}
	response, err := raw.ImportDocumentsWithBody(ctx,
		d.collectionName, importParams(params), "application/octet-stream", body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return nil, fmt.Errorf("typesense: read HTTP %d response: %w", response.StatusCode, readErr)
		}
		return nil, &HTTPError{Status: response.StatusCode, Body: body}
	}
	return response.Body, nil
}
