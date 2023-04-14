package mgoc

import (
	"encoding/hex"
	"fmt"
	"github.com/civet148/log"
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
			oid, err := Str2ObjectID(v.(string))
			if err != nil {
				log.Errorf(err.Error())
				return v
			}
			id = oid
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
			oid, err := Str2ObjectID(vid.Hex())
			if err != nil {
				log.Errorf(err.Error())
				return v
			}
			id = oid
		}
	default:
		id = v
	}
	return
}

func Str2ObjectID(strId string) (id interface{}, err error) {
	//log.Debugf("id [%v]", strId)
	if len(strId) == 0 {
		return nil, nil
	}
	if len(strId) == OfficalObjectIdSize { //mongo-driver ObjectID
		oid, err := primitive.ObjectIDFromHex(strId)
		if err != nil {
			return nil, fmt.Errorf("parse object id from string %s error %s", strId, err)
		}
		//log.Debugf("mongo-driver id [%v]", oid)
		return oid, nil
	} else if len(strId) == MgoV2ObjectIdSize { //mgo.v2 ObjectId
		hexid, err := hex.DecodeString(strId)
		if err != nil {
			return nil, fmt.Errorf("decode mgo.v2 object id from string [%s] error [%s]", strId, err)
		}
		oid, err := primitive.ObjectIDFromHex(string(hexid))
		if err != nil {
			return nil, fmt.Errorf("parse object id from string %s error %s", strId, err)
		}
		//log.Debugf("mgo.v2 id [%v]", oid)
		return oid, nil
	}
	return strId, nil
}
