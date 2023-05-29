package mgoc

import (
	"encoding/hex"
	"fmt"
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	bson2 "gopkg.in/mgo.v2/bson"
	"time"
)

const (
	radianBase          = float64(6359.0)
	OfficalObjectIdSize = 24
	MgoV2ObjectIdSize   = 48
)

func NewObjectID() ObjectID {
	return primitive.NewObjectID()
}

func NewObjectIDFromString(v string) (ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(v)
	if err != nil {
		return oid, log.Errorf("new object id from string %s error [%s]", v, err)
	}
	return oid, nil
}

func NewObjectIDFromTimestamp(t time.Time) ObjectID {
	return primitive.NewObjectIDFromTimestamp(t)
}

func NewDecimal128(high, low uint64) Decimal128 {
	return primitive.NewDecimal128(high, low)
}

func NewDateTimeFromTime(t time.Time) DateTime {
	return primitive.NewDateTimeFromTime(t)
}

func ConvertValue(column string, value interface{}) (v interface{}) {
	if column == defaultPrimaryKeyName {
		v = MakeObjectID(value)
	} else {
		v = value
	}
	return v
}

func MakeObjectID(v interface{}) (id interface{}) {
	//log.Debugf("value type [%v]", reflect.TypeOf(v))
	switch v.(type) {
	case string:
		{
			id = Str2ObjectID(v.(string))
		}
	case primitive.ObjectID:
		{
			id = v
		}
	case bson2.ObjectId:
		{
			vid := v.(bson2.ObjectId)
			hexid := vid.Hex()
			if len(hexid) == 0 {
				return nil
			}
			id = Str2ObjectID(vid.Hex())
		}
	default:
		id = v
	}
	return
}

func Str2ObjectID(strId string) (id interface{}) {
	//log.Debugf("id [%v]", strId)
	if len(strId) == 0 {
		return nil
	}
	if len(strId) == OfficalObjectIdSize { //mongo-driver MakeObjectID
		oid, err := primitive.ObjectIDFromHex(strId)
		if err != nil {
			return strId
		}
		//log.Debugf("mongo-driver id [%v]", oid)
		return oid
	} else if len(strId) == MgoV2ObjectIdSize { //mgo.v2 ObjectId
		hexid, err := hex.DecodeString(strId)
		if err != nil {
			return strId
		}
		oid, err := primitive.ObjectIDFromHex(string(hexid))
		if err != nil {
			return strId
		}
		//log.Debugf("mgo.v2 id [%v]", oid)
		return oid
	}
	return strId
}

func All(expr interface{}) bson.M {
	return bson.M{
		KeyAll: expr,
	}
}

func Sum(expr interface{}) bson.M {
	return bson.M{
		KeySum: expr,
	}
}

func ToBool(expr interface{}) bson.M {
	return bson.M{
		toBool: expr,
	}
}

func ToDecimal(expr interface{}) bson.M {
	return bson.M{
		toDecimal: expr,
	}
}

func ToDouble(expr interface{}) bson.M {
	return bson.M{
		toDouble: expr,
	}
}

func ToInt(expr interface{}) bson.M {
	return bson.M{
		toInt: expr,
	}
}

func ToLong(expr interface{}) bson.M {
	return bson.M{
		toLong: expr,
	}
}

func ToDate(expr interface{}) bson.M {
	return bson.M{
		toDate: expr,
	}
}

func ToString(expr interface{}) bson.M {
	return bson.M{
		toString: expr,
	}
}

func ToObjectId(expr interface{}) bson.M {
	return bson.M{
		toObjectId: expr,
	}
}

func ToLower(expr interface{}) bson.M {
	return bson.M{
		toLower: expr,
	}
}

func ToUpper(expr interface{}) bson.M {
	return bson.M{
		toUpper: expr,
	}
}

func RoundColumn(strColumn string, place int) bson.M {
	return bson.M{
		KeyRound: bson.A{
			fmt.Sprintf("$%s", strColumn),
			place,
		},
	}
}

//计算弧度
func Radian(meters uint64) float64 {
	r := float64(meters) / 1000
	return r / radianBase
}

func NewGeoPoint(coord Coordinate) *GeoPoint {
	return &GeoPoint{
		Type:        GeoTypePoint,
		Coordinates: []float64{coord.X, coord.Y},
	}
}

func NewGeoMultiPoint(coords []Coordinate) *GeoMultiPoint {
	var coordinates FloatArray2
	for _, coord := range coords {
		coordinates = append(coordinates, FloatArray{coord.X, coord.Y})
	}
	return &GeoMultiPoint{
		Type:        GeoTypeMultiPoint,
		Coordinates: coordinates,
	}
}

func NewGeoLineString(coords []Coordinate) *GeoLineString {
	var coordinates FloatArray2
	for _, coord := range coords {
		coordinates = append(coordinates, FloatArray{coord.X, coord.Y})
	}
	return &GeoLineString{
		Type:        GeoTypeLineString,
		Coordinates: coordinates,
	}
}

func NewGeoMultiLineString(coords [][]Coordinate) *GeoMultiLineString {
	var coordinates FloatArray3
	for _, coord := range coords {
		var cs FloatArray2
		for _, v := range coord {
			cs = append(cs, FloatArray{v.X, v.Y})
		}
		coordinates = append(coordinates, cs)
	}
	return &GeoMultiLineString{
		Type:        GeoTypeMultiLineString,
		Coordinates: coordinates,
	}
}

func NewGeoPolygon(coords [][]Coordinate) *GeoPolygon {
	var coordinates FloatArray3
	for _, coord := range coords {
		var cs FloatArray2
		for _, v := range coord {
			cs = append(cs, FloatArray{v.X, v.Y})
		}
		coordinates = append(coordinates, cs)
	}
	return &GeoPolygon{
		Type:        GeoTypePolygon,
		Coordinates: coordinates,
	}
}

func NewGeoMultiPolygon(coords [][][]Coordinate) *GeoMultiPolygon {
	var coordinates FloatArray4
	for _, coord := range coords {
		var cs FloatArray3
		for _, v := range coord {
			var cs2 FloatArray2
			for _, v2 := range v {
				cs2 = append(cs2, FloatArray{v2.X, v2.Y})
			}
			cs = append(cs, cs2)
		}
		coordinates = append(coordinates, cs)
	}
	return &GeoMultiPolygon{
		Type:        GeoTypeMultiPolygon,
		Coordinates: coordinates,
	}
}

func NewGeoMetry(typ GeoType, coordinates interface{}) *Geometry {
	return &Geometry{
		Type:        typ,
		Coordinates: coordinates,
	}
}
