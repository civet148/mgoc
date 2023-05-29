package mgoc

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	bson2 "gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
)

type ModelType int

const (
	ModelType_Struct   = 1
	ModelType_Slice    = 2
	ModelType_Map      = 3
	ModelType_BaseType = 4
)

func (m ModelType) GoString() string {
	return m.String()
}

func (m ModelType) String() string {
	switch m {
	case ModelType_Struct:
		return "ModelType_Struct"
	case ModelType_Slice:
		return "ModelType_Slice"
	case ModelType_Map:
		return "ModelType_Map"
	case ModelType_BaseType:
		return "ModelType_BaseType"
	}
	return "ModelType_Unknown"
}

// clone engine
func (e *Engine) clone(strDatabaseName string, models ...interface{}) *Engine {
	var opts []*options.DatabaseOptions
	if e.engineOpt.DatabaseOpt != nil {
		opts = append(opts, e.engineOpt.DatabaseOpt)
	}
	engine := &Engine{
		debug:           e.debug,
		engineOpt:       e.engineOpt,
		client:          e.client,
		strPkName:       e.strPkName,
		models:          make([]interface{}, 0),
		exceptColumns:   make(map[string]bool),
		dict:            make(map[string]interface{}),
		filter:          make(map[string]interface{}),
		updates:         make(map[string]interface{}),
		andConditions:   make(map[string]interface{}),
		orConditions:    make(map[string]interface{}),
		groupConditions: make(map[string]interface{}),
		groupByExprs:    make(map[string]interface{}),
		db:              e.client.Database(strDatabaseName, opts...),
	}
	return engine.setModel(models...)
}
func (e *Engine) debugJson(args ...interface{}) {
	if e.debug {
		log.Json(args...)
	}
}

func (e *Engine) setModel(models ...interface{}) *Engine {
	for _, v := range models {
		if v == nil {
			continue
		}
		var isStructPtrPtr bool
		typ := reflect.TypeOf(v)
		val := reflect.ValueOf(v)
		if typ.Kind() == reflect.Ptr {

			typ = typ.Elem()
			val = val.Elem()
			switch typ.Kind() {
			case reflect.Ptr:
				{
					if typ.Elem().Kind() == reflect.Struct { //struct pointer address (&*StructType)
						if val.IsNil() {
							var typNew = typ.Elem()
							var valNew = reflect.New(typNew)
							val.Set(valNew)
						}
						isStructPtrPtr = true
					}
				}
			}
		}

		if isStructPtrPtr {
			e.models = append(e.models, val.Interface())
			e.setModelType(ModelType_Struct)
		} else {
			switch typ.Kind() {
			case reflect.Struct: // struct
				e.setModelType(ModelType_Struct)
			case reflect.Slice: //  slice
				e.setModelType(ModelType_Slice)
			case reflect.Map: // map
				e.setModelType(ModelType_Map)
			default: //base type
				e.setModelType(ModelType_BaseType)
			}
			if typ.Kind() == reflect.Slice {
				modelVal := reflect.ValueOf(v)
				elemTyp := modelVal.Type().Elem()
				elemVal := reflect.New(elemTyp).Elem()
				if typ.Kind() == reflect.Slice && val.IsNil() {
					val.Set(reflect.MakeSlice(elemVal.Type(), 0, 0))
					if len(e.models) == 0 && len(models) != 0 { //used to fetch records
						e.models = models
					}
				} else {
					for j := 0; j < val.Len(); j++ { //used to insert/update records
						e.models = append(e.models, val.Index(j).Interface()) //append elements to the models
					}
				}
			} else if typ.Kind() == reflect.Struct || typ.Kind() == reflect.Map {
				e.models = append(e.models, v) //map, struct or slice
			} else {
				e.models = models //built-in type int/string/float32...
			}
		}

		//var selectColumns []string
		e.dict = newReflector(e, e.models).ToMap()
		break //only check first argument
	}
	return e
}

func (e *Engine) setSelectColumns(strColumns ...string) {
	if len(strColumns) == 0 {
		return
	}
	e.selectColumns = e.appendStrings(e.selectColumns, strColumns...)
}

func (e *Engine) setExceptColumns(strColumns ...string) {
	for _, col := range strColumns {
		e.exceptColumns[col] = true
	}
}

func (e *Engine) setAscColumns(strColumns ...string) {
	e.ascColumns = e.appendStrings(e.ascColumns, strColumns...)
}

func (e *Engine) setDescColumns(strColumns ...string) {
	e.descColumns = e.appendStrings(e.descColumns, strColumns...)
}

func (e *Engine) appendStrings(src []string, dest ...string) []string {
	//check duplicated elements
	for _, v := range dest {
		if !e.exist(src, v) {
			src = append(src, v)
		}
	}
	return src
}

func (e *Engine) exist(src []string, s string) bool {
	for _, v := range src {
		if s == v {
			return true
		}
	}
	return false
}

func (e *Engine) getModelType() ModelType {
	return e.modelType
}

func (e *Engine) setModelType(modelType ModelType) {
	e.modelType = modelType
}

func (e *Engine) setTableName(strNames ...string) {
	e.strTableName = strings.Join(strNames, ",")
}

func (e *Engine) clean() {
	e.options = nil
	e.models = nil
	e.modelType = 0
	e.exceptColumns = make(map[string]bool)
	e.filter = make(map[string]interface{})
}

//assert bool and string/struct/slice/map nil, call panic
func assert(v interface{}, message string) {
	if isNilOrFalse(v) {
		log.Panic(message)
	}
}

// judgement: bool, integer, string, struct, slice, map is nil or false?
func isNilOrFalse(v interface{}) bool {
	switch v.(type) {
	case string:
		if v.(string) == "" {
			return true
		}
	case bool:
		return !v.(bool)
	case int8, int16, int, int32, int64, uint8, uint16, uint, uint32, uint64:
		{
			if fmt.Sprintf("%v", v) == "0" {
				return true
			}
		}
	default:
		{
			val := reflect.ValueOf(v)
			return val.IsNil()
		}
	}
	return false
}

func (e *Engine) makeFindOptions() []*options.FindOptions {
	var opts []*options.FindOptions
	for _, opt := range e.options {
		opts = append(opts, opt.(*options.FindOptions))
	}
	if len(opts) == 0 {
		var opt = &options.FindOptions{}
		if e.limit != 0 {
			opt.SetLimit(e.limit)
			opt.SetSkip(e.skip)
		}
		opt.SetProjection(e.makeProjection())
		if len(e.ascColumns) != 0 || len(e.descColumns) != 0 {
			opt.SetSort(e.makeSort())
		}
		opts = append(opts, opt)
	} else {
		opt := opts[0]
		if opt.Limit == nil {
			opt.SetLimit(e.limit)
		}
		if opt.Skip == nil {
			opt.SetSkip(e.skip)
		}
		if opt.Sort == nil {
			opt.SetSort(e.makeSort())
		}
		if opt.Projection == nil {
			opt.SetProjection(e.makeProjection())
		}
	}
	return opts
}

func (e *Engine) makeProjection() bson.M {
	var projection = bson.M{}
	for _, v := range e.selectColumns {
		projection[v] = 1
	}
	for _, v := range e.roundColumns {
		projection[v.AS] = RoundColumn(v.Column, v.Place)
	}
	return projection
}

func (e *Engine) makeSort() bson.M {
	var sort = bson.M{}
	for _, v := range e.ascColumns {
		sort[v] = 1
	}
	for _, v := range e.descColumns {
		sort[v] = -1
	}
	return sort
}

func (e *Engine) makeFilters() bson.M {
	if e.filter == nil {
		e.makeFilterMap()
	}
	and := e.makeAndCondition()
	if len(and) != 0 {
		e.filter[KeyAnd] = and
	}
	or := e.makeOrCondition()
	if len(or) != 0 {
		e.filter[KeyOr] = or
	}
	e.filter = e.replaceObjectID(e.filter)
	return e.filter
}

func (e *Engine) isPipelineKeyExist(key string) bool {
	for _, pipe := range e.pipeline {
		for k := range pipe.Map() {
			if k == key {
				return true
			}
		}
	}
	return false
}

//replaceObjectID replace filter _id string to Str2ObjectID
func (e *Engine) replaceObjectID(filter bson.M) bson.M {
	assert(filter, "filter cannot be nil")
	for k, v := range filter {
		if k == defaultPrimaryKeyName {
			switch v.(type) {
			case string:
				{
					oid, err := primitive.ObjectIDFromHex(v.(string))
					if err != nil {
						log.Errorf("parse object id from string %s error %s", v.(string), err)
						return filter
					}
					filter[k] = oid
				}
			}
		}
	}
	return filter
}

func (e *Engine) fetchRows(cur *mongo.Cursor) (err error) {
	if e.modelType == ModelType_Struct || e.modelType == ModelType_Map {
		for _, model := range e.models {
			if !cur.Next(context.TODO()) {
				break
			}
			err = cur.Decode(model)
			if err != nil {
				return log.Errorf(err.Error())
			}
		}
	} else if e.modelType == ModelType_Slice {
		err = cur.All(context.TODO(), e.models[0])
		if err != nil {
			return log.Errorf(err.Error())
		}
	} else {
		return log.Errorf("model type %s not support yet", e.modelType)
	}
	e.replaceQueryObjectID()
	return
}

func (e *Engine) makeUpdates() {
	//select columns to update
	e.makeSelectUpdates()
	//make primary column for filter
	e.makePrimaryKeyUpdates()
}

//makePrimaryKeyUpdates make primary column for filter
func (e *Engine) makePrimaryKeyUpdates() {
	for k, v := range e.dict {
		if k == e.PrimaryKey() && v != nil {
			e.Id(v)
		}
	}
}

//makeSelectUpdates make selected columns to update
func (e *Engine) makeSelectUpdates() {
	if len(e.selectColumns) == 0 {
		for k, v := range e.dict {
			if e.isExcepted(k) {
				continue
			}
			e.Set(k, v)
		}
	} else {
		for _, col := range e.selectColumns {
			if e.isExcepted(col) {
				continue
			}
			e.Set(col, e.dict[col])
		}
	}
}

//makeExceptUpdates make except columns to update
func (e *Engine) isExcepted(col string) (ok bool) {
	_, ok = e.exceptColumns[col]
	return ok
}

func (e *Engine) setAndCondition(strColumn string, value interface{}) {
	e.locker.Lock()
	defer e.locker.Unlock()
	e.andConditions[strColumn] = value
}

func (e *Engine) makeFilterMap() {
	e.locker.Lock()
	defer e.locker.Unlock()
	if e.filter == nil {
		e.filter = make(map[string]interface{})
	}
}

func (e *Engine) makeAndCondition() (cond bson.A) {
	e.locker.RLock()
	defer e.locker.RUnlock()
	for k, v := range e.andConditions {
		cond = append(cond, bson.M{k: v})
	}
	return
}

func (e *Engine) setOrCondition(strColumn string, value interface{}) {
	e.locker.Lock()
	defer e.locker.Unlock()
	e.orConditions[strColumn] = value
}

func (e *Engine) makeOrCondition() (cond bson.A) {
	e.locker.RLock()
	defer e.locker.RUnlock()
	for k, v := range e.orConditions {
		cond = append(cond, bson.M{k: v})
	}
	return
}

func (e *Engine) replaceInsertModels() {
	var mms []map[string]interface{}

	for _, model := range e.models {
		typ := reflect.TypeOf(model)
		val := reflect.ValueOf(model)
		for {
			if typ.Kind() != reflect.Ptr { // pointer type
				break
			}
			typ = typ.Elem()
			val = val.Elem()
		}

		kind := typ.Kind()
		switch kind {
		case reflect.Struct:
			{
				var mm = make(map[string]interface{})
				NumField := val.NumField()
				for i := 0; i < NumField; i++ {
					typField := typ.Field(i)
					valField := val.Field(i)

					if typField.Type.Kind() == reflect.Ptr {
						typField.Type = typField.Type.Elem()
						valField = valField.Elem()
					}
					if !valField.IsValid() || !valField.CanInterface() {
						continue
					}
					tagVal, ignore := getTagValue(typField, TAG_NAME_BSON)
					if ignore {
						continue
					}
					if tagVal == defaultPrimaryKeyName {
						id := MakeObjectID(valField.Interface())
						if id != nil {
							if vid, ok := id.(ObjectID); ok {
								if vid.IsZero() {
									vid = primitive.NewObjectID()
								}
								id = vid
							}
							mm[tagVal] = id
						}
					} else {
						mm[tagVal] = valField.Interface()
					}
				}
				mms = append(mms, mm)
			}
		}
	}
	e.models = nil
	for _, m := range mms {
		e.models = append(e.models, m)
	}
}

// replaceQueryObjectID parse struct and replace mgo.v2 ObjectId
func (e *Engine) replaceQueryObjectID() {
	for _, model := range e.models {
		typ := reflect.TypeOf(model)
		val := reflect.ValueOf(model)
		for {
			if typ.Kind() != reflect.Ptr { // pointer type
				break
			}
			typ = typ.Elem()
			val = val.Elem()
		}

		kind := typ.Kind()
		switch kind {
		case reflect.Struct:
			{
				e.replaceStructFiledObjectId(typ, val)
			}
		case reflect.Slice:
			{
				for i := 0; i < val.Len(); i++ {
					childTyp := reflect.TypeOf(val.Index(i).Interface())
					childVal := reflect.ValueOf(val.Interface()).Index(i)
					if childVal.Kind() == reflect.Ptr {
						childTyp = childTyp.Elem()
						childVal = childVal.Elem()
					}
					//log.Debugf("slice[%d] type [%v] %+v", i, childTyp, childVal)
					e.replaceStructFiledObjectId(childTyp, childVal)
				}
			}
		default:
		}
	}
}

// replaceStructFiledObjectId parse struct fields and replace mgo.v2 ObjectId value
func (e *Engine) replaceStructFiledObjectId(typ reflect.Type, val reflect.Value) {
	kind := typ.Kind()
	if kind == reflect.Struct {
		NumField := val.NumField()
		for i := 0; i < NumField; i++ {
			typField := typ.Field(i)
			valField := val.Field(i)
			if !valField.IsValid() || !valField.CanInterface() {
				continue
			}
			tagVal, ignore := getTagValue(typField, TAG_NAME_BSON)
			if ignore {
				continue
			}
			//log.Debugf("type [%v] tag [%s] value [%v]", typField.Type, tagVal, valField.Interface())
			if tagVal == e.PrimaryKey() {
				vid := valField.Interface()
				switch valField.Interface().(type) {
				case bson2.ObjectId:
					{
						id := vid.(bson2.ObjectId)
						hid := id.Hex()
						if len(hid) == MgoV2ObjectIdSize {
							data, err := hex.DecodeString(hid)
							if err != nil {
								log.Errorf(err.Error())
								return
							}
							oid := bson2.ObjectIdHex(string(data))
							valField.Set(reflect.ValueOf(oid))
						}
					}
				}
			}
		}
	}
}

func (e *Engine) addGroupCondition(column, key string, values ...interface{}) *Engine {
	var value interface{}
	if len(values) > 0 {
		value = values[0]
	} else {
		value = fmt.Sprintf("$%s", column)
	}
	e.isAggregate = true
	e.groupConditions[column] = bson.M{key: value}
	return e
}

func (e *Engine) makePipelineUnwind() bson.D {
	if e.isPipelineKeyExist(KeyUnwind) {
		return nil
	}
	if e.unwind == nil {
		return nil
	}
	var value interface{}
	if s, ok := e.unwind.(string); ok {
		value = fmt.Sprintf("$%s", s)
	}
	var sort = bson.D{
		{KeyUnwind, value},
	}
	return sort
}

func (e *Engine) makePipelineSort() bson.D {
	if e.isPipelineKeyExist(KeySort) {
		return nil
	}
	s := e.makeSort()
	if len(s) == 0 {
		return nil
	}
	var sort = bson.D{
		{KeySort, s},
	}
	return sort
}

func (e *Engine) makePipelineSkip() bson.D {
	if e.isPipelineKeyExist(KeySkip) {
		return nil
	}
	if e.skip == 0 {
		return nil
	}
	var skip = bson.D{
		{KeySkip, e.skip},
	}
	return skip
}

func (e *Engine) makePipelineLimit() bson.D {
	if e.isPipelineKeyExist(KeyLimit) {
		return nil
	}
	if e.limit == 0 {
		return nil
	}
	var limit = bson.D{
		{KeyLimit, e.limit},
	}
	return limit
}

func (e *Engine) makePipelineMatch() bson.D {
	if e.isPipelineKeyExist(KeyMatch) {
		return nil
	}
	filters := e.makeFilters()
	if len(filters) == 0 {
		return nil
	}
	var match bson.D
	match = bson.D{
		{KeyMatch, filters},
	}
	return match
}

func (e *Engine) makePipelineGroup() bson.D {
	if e.isPipelineKeyExist(KeyGroup) {
		return nil
	}
	var group bson.D
	if len(e.groupConditions) == 0 {
		return nil
	}

	if _, ok := e.groupConditions[defaultPrimaryKeyName]; !ok {
		if len(e.groupByExprs) != 0 {
			e.groupConditions[defaultPrimaryKeyName] = e.groupByExprs
		} else {
			e.groupConditions[defaultPrimaryKeyName] = nil
		}
	}

	group = bson.D{
		{KeyGroup, e.groupConditions},
	}
	return group
}

func (e *Engine) makePipelineProjection() bson.D {
	if e.isPipelineKeyExist(KeyProject) {
		return nil
	}
	var project bson.D
	projection := e.makeProjection()
	if len(projection) == 0 {
		return nil
	}
	project = bson.D{
		{KeyProject, projection},
	}
	return project
}

func (e *Engine) makeGroupByPipelines() *Engine {
	if len(e.pipeline) != 0 {
		return e
	}
	var pipelines []bson.D
	if p := e.makePipelineMatch(); p != nil {
		pipelines = append(pipelines, p)
	}

	if p := e.makePipelineGroup(); p != nil {
		pipelines = append(pipelines, p)
	}

	if p := e.makePipelineProjection(); p != nil {
		pipelines = append(pipelines, p)
	}

	if p := e.makePipelineSort(); p != nil {
		pipelines = append(pipelines, p)
	}

	if p := e.makePipelineSkip(); p != nil {
		pipelines = append(pipelines, p)
	}

	if p := e.makePipelineLimit(); p != nil {
		pipelines = append(pipelines, p)
	}

	if p := e.makePipelineUnwind(); p != nil {
		pipelines = append(pipelines, p)
	}
	return e.Pipeline(pipelines...)
}
