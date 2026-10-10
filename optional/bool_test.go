package optional_test

import (
	"encoding/json"
	"testing"
	"typesense/api"
	"typesense/optional"
)

func TestBool(t *testing.T) {
	for _, tc := range []struct {
		name      string
		value     optional.Bool
		json      string
		set       bool
		boolValue bool
	}{
		{"None", optional.None, "null", false, false},
		{"False", optional.False, "false", true, false},
		{"True", optional.True, "true", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := json.Marshal(tc.value)
			requireNoError(t, err)
			if string(encoded) != tc.json {
				t.Fatalf("wrong encoding: %s", encoded)
			}
			var decoded optional.Bool
			requireNoError(t, json.Unmarshal(encoded, &decoded))
			value, set := decoded.Value()
			if decoded != tc.value || set != tc.set || value != tc.boolValue {
				t.Fatalf("wrong decoded state: %v", decoded)
			}
			field := api.Field{Name: "secret", Type: "string", Index: tc.value, Store: tc.value}
			encoded, err = json.Marshal(field)
			requireNoError(t, err)
			var payload map[string]any
			requireNoError(t, json.Unmarshal(encoded, &payload))
			for _, key := range []string{"index", "store"} {
				value, exists := payload[key]
				if exists != tc.set || (exists && value != tc.boolValue) {
					t.Fatalf("wrong omission/explicit bool: %s", encoded)
				}
			}
		})
	}
	if optional.FromBool(true) != optional.True || optional.FromBool(false) != optional.False {
		t.Fatal("FromBool returned wrong state")
	}
	if _, err := json.Marshal(optional.Bool(255)); err == nil {
		t.Fatal("invalid bool state encoded")
	}
	for _, payload := range []string{`0`, `1`, `"true"`, `{}`} {
		value := optional.True
		if err := json.Unmarshal([]byte(payload), &value); err == nil {
			t.Fatalf("accepted invalid bool %s", payload)
		}
		if value != optional.True {
			t.Fatal("invalid input changed existing state")
		}
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
