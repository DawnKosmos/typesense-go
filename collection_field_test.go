package typesense_test

import (
	"reflect"
	"testing"

	"typesense"
	"typesense/api"
	"typesense/optional"
	data "typesense/tests_data"
)

func TestCollectionFields(t *testing.T) {
	t.Run("NamesOptionsAndReferences", func(t *testing.T) {
		type Embedded struct {
			Code string `json:"code" tyse:"index"`
		}
		type Item struct {
			Embedded
			Embedding []api.Float `json:"embedding,omitempty" tyse:"index,dim:384"`
			Owner     string      `json:"owner_id" tyse:"index,join:users.id,optional"`
			Secret    string      `json:"secret" tyse:"optional,nostore"`
			Tag       string      `json:"tag" tyse:"facet,infix,sort"`
			Default   string      `json:"default" tyse:""`
			Skipped   string      `json:"skipped"`
			Ignored   string      `json:"-" tyse:"index"`
			Disabled  string      `json:"disabled" tyse:"-"`
		}
		name, fields, err := typesense.ToFieldWithName((*Item)(nil))
		requireNoError(t, err)
		if name != "item" || len(fields) != 6 {
			t.Fatalf("wrong name or fields: %s %+v", name, fields)
		}
		byName := map[string]api.Field{}
		for _, field := range fields {
			byName[field.Name] = field
		}
		if byName["embedding"].Type != string(api.FLOATARRAY) || byName["embedding"].NumDim != 384 {
			t.Fatalf("bad vector field: %+v", byName["embedding"])
		}
		if byName["owner_id"].Reference != "users.id" || byName["owner_id"].Optional != optional.True {
			t.Fatalf("bad reference: %+v", byName["owner_id"])
		}
		if byName["secret"].Store != optional.False || byName["default"].Store != optional.None || byName["default"].Index != optional.False {
			t.Fatal("wrong default or nostore options")
		}
		tag := byName["tag"]
		if tag.Index != optional.True || tag.Facet != optional.True || tag.Infix != optional.True || tag.Sort != optional.True {
			t.Fatalf("bad flag options: %+v", tag)
		}
		fieldsFromType := typesense.ToField(reflect.TypeFor[Item]())
		if !reflect.DeepEqual(fieldsFromType, fields) {
			t.Fatal("reflect.Type and pointer inference differ")
		}
	})
	t.Run("InferredAndExplicitTypes", func(t *testing.T) {
		type Fields struct {
			Point    api.GeoPoint             `json:"point" tyse:"index"`
			Points   api.GeoPointArray        `json:"points" tyse:"index"`
			Polygon  api.Boundary             `json:"polygon" tyse:"index"`
			Image    api.Image                `json:"image" tyse:"optional"`
			Object   struct{ Value string }   `json:"object" tyse:"index"`
			Objects  []struct{ Value string } `json:"objects" tyse:"index"`
			Any      api.Object               `json:"any" tyse:"index"`
			Map      map[string]any           `json:"map" tyse:"index"`
			Bool     optional.Bool            `json:"bool" tyse:"index"`
			Strings  [2]string                `json:"strings" tyse:"index"`
			Pointer  *int64                   `json:"pointer" tyse:"index,optional"`
			Override [][]float64              `json:"override" tyse:"type:geopolygon,index"`
			Mixed    api.StringOrArray        `json:"mixed" tyse:"type:string*,index"`
		}
		expected := []api.Type{api.GEOPOINT, api.GEOPOINTARRAY, api.GEOPOLYGON, api.IMAGE, api.OBJECT, api.OBJECTARRAY, api.OBJECT, api.OBJECT, api.BOOL, api.STRINGARRAY, api.INT64, api.GEOPOLYGON, api.STRINGPTR}
		schema, err := typesense.ToSchema(Fields{})
		requireNoError(t, err)
		if schema.EnableNestedFields != optional.True || len(schema.Fields) != len(expected) {
			t.Fatalf("wrong nested schema: %+v", schema)
		}
		for i, typ := range expected {
			if schema.Fields[i].Type != string(typ) {
				t.Errorf("field %s: got %s, want %s", schema.Fields[i].Name, schema.Fields[i].Type, typ)
			}
		}
		scalar, err := typesense.ToSchema(data.Product{})
		requireNoError(t, err)
		if scalar.EnableNestedFields != optional.None {
			t.Fatal("scalar schema enabled nested fields")
		}
	})
	t.Run("InvalidInput", func(t *testing.T) {
		for _, input := range []any{
			nil, 1, []string{},
			struct {
				Value chan int `tyse:"index"`
			}{},
			struct {
				Value string `tyse:"dim:0"`
			}{},
			struct {
				Value string `tyse:"dim:no"`
			}{},
			struct {
				Value string `tyse:"store"`
			}{},
			struct {
				Value string `tyse:"unknown"`
			}{},
			struct {
				Value string `tyse:"type:invalid"`
			}{},
			struct {
				Value string `tyse:"join:"`
			}{},
			struct {
				A string `json:"same" tyse:"index"`
				B string `json:"same" tyse:"index"`
			}{},
		} {
			if _, _, err := typesense.ToFieldWithName(input); err == nil {
				t.Errorf("invalid input accepted: %T", input)
			}
		}
		type Recursive struct{ *Recursive }
		if _, _, err := typesense.ToFieldWithName(Recursive{}); err == nil {
			t.Fatal("recursive embedding accepted")
		}
	})
}

func TestCollectionFieldEdgeCases(t *testing.T) {
	type Fields struct {
		unexported string  `tyse:"index"`
		Integer    int     `tyse:"index"`
		Integer32  int32   `tyse:"index"`
		Float32    float32 `tyse:"index"`
		Bool       bool    `tyse:"index"`
		Reference  string  `tyse:"reference:users"`
	}
	fields := typesense.ToField(Fields{})
	if len(fields) != 5 || fields[0].Name != "Integer" || fields[0].Type != "int32" || fields[2].Type != "float" || fields[3].Type != "bool" || fields[4].Reference != "users" {
		t.Fatalf("wrong inferred fields: %+v", fields)
	}
	for _, input := range []any{
		struct {
			Value []chan int `tyse:""`
		}{},
		struct {
			Value [][]string `tyse:""`
		}{},
		struct {
			Value string `tyse:"unknown:value"`
		}{},
	} {
		if _, err := typesense.ToSchema(input); err == nil {
			t.Fatalf("invalid schema accepted: %T", input)
		}
	}
	t.Run("PanicsOnInvalidInput", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("ToField did not panic")
			}
		}()
		typesense.ToField(nil)
	})
	name, _, err := typesense.ToFieldWithName(struct{}{})
	requireNoError(t, err)
	if name != "" {
		t.Fatal("anonymous type acquired a collection name")
	}
}
