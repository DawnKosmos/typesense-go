package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"typesense/optional"

	"github.com/oapi-codegen/runtime"
)

// styleParamWithOptions preserves JSON omission and custom scalar encoders for
// the spec's object-valued query parameters before applying OpenAPI query styling.
func styleParamWithOptions(style string, explode bool, name string, value any, options runtime.StyleParamOptions) (string, error) {
	if boolean, ok := value.(optional.Bool); ok {
		if boolean == optional.None {
			return "", nil
		}
		if boolean != optional.False && boolean != optional.True {
			return "", fmt.Errorf("query parameter %s: invalid optional.Bool state %d", name, boolean)
		}
		value = boolean == optional.True
	}
	if options.ParamLocation == runtime.ParamLocationQuery && options.Type == "object" {
		data, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("encode query parameter %s: %w", name, err)
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		var normalized any
		if err := decoder.Decode(&normalized); err != nil {
			return "", fmt.Errorf("decode query parameter %s: %w", name, err)
		}
		// nil optional objects should not produce a query parameter.
		if normalized == nil {
			return "", nil
		}
		value = normalized
	}
	return runtime.StyleParamWithOptions(style, explode, name, value, options)
}
