package typesense

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"typesense/api"
	"typesense/optional"
)

// Prefix applies the prefix configured with SetPrefix once.
// Configure the prefix before constructing clients or schemas.
func Prefix(name string) string {
	if _prefix == "" || strings.HasPrefix(name, _prefix) {
		return name
	}
	return _prefix + name
}

// ToField builds fields from JSON names and tyse options and panics on invalid input.
// Untagged fields are skipped; untagged embedded structs are flattened.
func ToField(in any) []api.Field {
	_, fields, err := ToFieldWithName(in)
	if err != nil {
		panic(err)
	}
	return fields
}

// ToFieldWithName returns a prefixed, lowercase struct name and its tagged fields.
// Struct values, pointers (including nil pointers), and reflect.Type are accepted.
func ToFieldWithName(in any) (string, []api.Field, error) {
	typ, ok := in.(reflect.Type)
	if !ok {
		typ = reflect.TypeOf(in)
	}
	if typ == nil {
		return "", nil, fmt.Errorf("typesense: input must be a struct")
	}
	typ = dereferenceType(typ)
	if typ.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("typesense: input must be a struct, got %s", typ)
	}
	fields, err := lexField(typ, make(map[reflect.Type]bool))
	if err != nil {
		return "", nil, err
	}
	seen := make(map[string]bool)
	for _, field := range fields {
		if seen[field.Name] {
			return "", nil, fmt.Errorf("typesense: duplicate field %q", field.Name)
		}
		seen[field.Name] = true
	}
	name := ""
	if typ.Name() != "" {
		name = Prefix(strings.ToLower(typ.Name()))
	}
	return name, fields, nil
}

// ToSchema builds a collection schema, enabling nested fields for object fields.
// Anonymous struct types require the caller to assign a collection name.
func ToSchema(in any) (*api.CollectionSchema, error) {
	name, fields, err := ToFieldWithName(in)
	if err != nil {
		return nil, err
	}
	return &api.CollectionSchema{Name: name, Fields: fields, EnableNestedFields: nestedFields(fields)}, nil
}

func lexField(typ reflect.Type, active map[reflect.Type]bool) ([]api.Field, error) {
	if active[typ] {
		return nil, fmt.Errorf("typesense: recursive embedded struct %s", typ)
	}
	active[typ] = true
	defer delete(active, typ)
	var fields []api.Field
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tag, tagged := field.Tag.Lookup("tyse")
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if tag == "-" || name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		fieldType := dereferenceType(field.Type)
		if field.Anonymous && !tagged && fieldType.Kind() == reflect.Struct {
			embedded, err := lexField(fieldType, active)
			if err != nil {
				return nil, err
			}
			fields = append(fields, embedded...)
			continue
		}
		if !tagged {
			continue
		}
		parsed, err := parseField(field.Type, name, tag)
		if err != nil {
			return nil, fmt.Errorf("typesense: field %s: %w", field.Name, err)
		}
		fields = append(fields, parsed)
	}
	return fields, nil
}

func parseField(typ reflect.Type, name, tag string) (api.Field, error) {
	field := api.Field{Name: name, Index: optional.False}
	for _, value := range strings.Split(tag, ",") {
		option := strings.TrimSpace(value)
		switch option {
		case "":
			continue
		case "optional":
			field.Optional = optional.True
		case "facet":
			field.Facet, field.Index = optional.True, optional.True
		case "index":
			field.Index = optional.True
		case "sort":
			field.Sort = optional.True
		case "infix":
			field.Infix = optional.True
		case "nostore":
			field.Store = optional.False
		default:
			key, value, found := strings.Cut(option, ":")
			if !found || value == "" {
				return api.Field{}, fmt.Errorf("unknown or empty tyse option %q", option)
			}
			switch key {
			case "dim":
				dim, err := strconv.Atoi(value)
				if err != nil || dim <= 0 {
					return api.Field{}, fmt.Errorf("dim must be a positive integer, got %q", value)
				}
				field.NumDim = dim
			case "join", "reference":
				field.Reference = prefixReference(value)
			case "type":
				if !validFieldType(api.Type(value)) {
					return api.Field{}, fmt.Errorf("unsupported field type %q", value)
				}
				field.Type = value
			default:
				return api.Field{}, fmt.Errorf("unknown tyse option %q", option)
			}
		}
	}
	if field.Type == "" {
		inferred, err := typeAllowed(typ)
		if err != nil {
			return api.Field{}, err
		}
		field.Type = string(inferred)
	}
	return field, nil
}

func prefixReference(reference string) string {
	collection, field, found := strings.Cut(reference, ".")
	if !found {
		return Prefix(reference)
	}
	return Prefix(collection) + "." + field
}

func typeAllowed(typ reflect.Type) (api.Type, error) {
	typ = dereferenceType(typ)
	switch typ {
	case reflect.TypeFor[optional.Bool]():
		return api.BOOL, nil
	case reflect.TypeFor[api.GeoPoint]():
		return api.GEOPOINT, nil
	case reflect.TypeFor[api.Boundary]():
		return api.GEOPOLYGON, nil
	case reflect.TypeFor[api.Image]():
		return api.IMAGE, nil
	case reflect.TypeFor[api.Object](), reflect.TypeFor[map[string]any]():
		return api.OBJECT, nil
	}
	switch typ.Kind() {
	case reflect.String:
		return api.STRING, nil
	case reflect.Int, reflect.Int32:
		return api.INT32, nil
	case reflect.Int64:
		return api.INT64, nil
	case reflect.Float32, reflect.Float64:
		return api.FLOAT, nil
	case reflect.Bool:
		return api.BOOL, nil
	case reflect.Struct:
		return api.OBJECT, nil
	case reflect.Slice, reflect.Array:
		element, err := typeAllowed(typ.Elem())
		if err != nil {
			return "", err
		}
		array := element + "[]"
		if validFieldType(array) {
			return array, nil
		}
	}
	return "", fmt.Errorf("unsupported Go type %s; provide a type: override", typ)
}

func dereferenceType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

func validFieldType(typ api.Type) bool {
	switch typ {
	case api.STRING, api.STRINGARRAY, api.INT32, api.INT32ARRAY, api.INT64, api.INT64ARRAY, api.FLOAT, api.FLOATARRAY, api.BOOL, api.BOOLARRAY, api.GEOPOINT, api.GEOPOINTARRAY, api.GEOPOLYGON, api.IMAGE, api.OBJECT, api.OBJECTARRAY, api.STRINGPTR:
		return true
	}
	return false
}

func nestedFields(fields []api.Field) optional.Bool {
	for _, field := range fields {
		if field.Type == string(api.OBJECT) || field.Type == string(api.OBJECTARRAY) {
			return optional.True
		}
	}
	return optional.None
}
