package mgoc

import (
	"encoding/hex"
	"go.mongodb.org/mongo-driver/bson/primitive"
	bson2 "gopkg.in/mgo.v2/bson"
)

const (
	OfficalObjectIdSize = 24
	MgoV2ObjectIdSize   = 48
)

func ConvertValue(column string, value interface{}) (v interface{}) {
	if column == defaultPrimaryKeyName {
		v = ObjectID(value)
	} else {
		v = value
	}
	return v
}

func ObjectID(v interface{}) (id interface{}) {
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
	if len(strId) == OfficalObjectIdSize { //mongo-driver ObjectID
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
