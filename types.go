package mgoc

import "go.mongodb.org/mongo-driver/bson/primitive"

type ObjectID = primitive.ObjectID
type Decimal128 = primitive.Decimal128
type DateTime = primitive.DateTime

type GeoType string

const (
	GeoTypePoint           GeoType = "Point"
	GeoTypeMultiPoint      GeoType = "MultiPoint"
	GeoTypeLineString      GeoType = "LineString"
	GeoTypeMultiLineString GeoType = "MultiLineString"
	GeoTypePolygon         GeoType = "Polygon"
	GeoTypeMultiPolygon    GeoType = "MultiPolygon"
)

type Coordinate struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

type FloatArray = []float64
type FloatArray2 = []FloatArray
type FloatArray3 = []FloatArray2
type FloatArray4 = []FloatArray3

type Geometry struct {
	Type        GeoType     `json:"type" bson:"type"`
	Coordinates interface{} `json:"coordinates" bson:"coordinates"`
}

type GeoPoint struct {
	Type        GeoType    `json:"type" bson:"type"`
	Coordinates FloatArray `json:"coordinates" bson:"coordinates"`
}

type GeoMultiPoint struct {
	Type        GeoType     `json:"type" bson:"type"`
	Coordinates FloatArray2 `json:"coordinates" bson:"coordinates"`
}

type GeoLineString struct {
	Type        GeoType     `json:"type" bson:"type"`
	Coordinates FloatArray2 `json:"coordinates" bson:"coordinates"`
}

type GeoMultiLineString struct {
	Type        GeoType     `json:"type" bson:"type"`
	Coordinates FloatArray3 `json:"coordinates" bson:"coordinates"`
}

type GeoPolygon struct {
	Type        GeoType     `json:"type" bson:"type"`
	Coordinates FloatArray3 `json:"coordinates" bson:"coordinates"`
}

type GeoMultiPolygon struct {
	Type        GeoType     `json:"type" bson:"type"`
	Coordinates FloatArray4 `json:"coordinates" bson:"coordinates"`
}
