package api

// Type is a Typesense collection field type.
type Type string

const (
	STRING        Type = "string"
	STRINGARRAY   Type = "string[]"
	INT32         Type = "int32"
	INT32ARRAY    Type = "int32[]"
	INT64         Type = "int64"
	INT64ARRAY    Type = "int64[]"
	FLOAT         Type = "float"
	FLOATARRAY    Type = "float[]"
	BOOL          Type = "bool"
	BOOLARRAY     Type = "bool[]"
	GEOPOINT      Type = "geopoint"
	GEOPOINTARRAY Type = "geopoint[]"
	GEOPOLYGON    Type = "geopolygon"
	OBJECT        Type = "object"
	OBJECTARRAY   Type = "object[]"
	STRINGPTR     Type = "string*"
	IMAGE         Type = "image"
)

type (
	String = string
	Int32  = int32
	Int64  = int64
	Float  = float64
	Bool   = bool
	// Object can hold a struct or another JSON object value.
	Object      = any
	ObjectArray = []Object

	// GeoPoint contains latitude followed by longitude.
	GeoPoint      [2]float64
	GeoPointArray = []GeoPoint
	// Boundary contains alternating latitude and longitude values for a polygon.
	Boundary []float64
	// Image contains a base64-encoded image.
	Image string
	// StringOrArray accepts string or []string values. Use tyse:"type:string*"
	// to select Typesense's string* type; the alias cannot be identified by reflection.
	StringOrArray = any
)
