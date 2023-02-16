package mgoc

import (
	"fmt"
	"github.com/civet148/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		debug:     e.debug,
		engineOpt: e.engineOpt,
		client:    e.client,
		strPkName: e.strPkName,
		dbTags:    e.dbTags,
		models:    make([]interface{}, 0),
		dict:      make(map[string]interface{}),
		filter:    make(map[string]interface{}),
		updates:   make(map[string]interface{}),
		db:        e.client.Database(strDatabaseName, opts...),
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

		var selectColumns []string
		e.dict = newReflector(e, e.models).ToMap(e.dbTags...)
		for k := range e.dict {
			selectColumns = append(selectColumns, k)
		}
		if len(selectColumns) > 0 {
			e.setSelectColumns(selectColumns...)
		}
		break //only check first argument
	}
	return e
}

func (e *Engine) setSelectColumns(strColumns ...string) (ok bool) {
	if len(strColumns) == 0 {
		return false
	}
	if e.selected {
		e.selectColumns = e.appendStrings(e.selectColumns, strColumns...)
	} else {
		e.selectColumns = strColumns
	}
	return true
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
	e.filter = make(map[string]interface{})
	e.options = nil
	e.models = nil
	e.modelType = 0
	e.selected = false
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
		var ok bool
		var opt = &options.FindOptions{}
		if e.limit != 0 {
			opt.SetLimit(e.limit)
			opt.SetSkip(e.skip)
			opt.SetProjection(e.makeProjection())
			ok = true
		}
		if len(e.ascColumns) != 0 || len(e.descColumns) != 0 {
			ok = true
			opt.SetSort(e.makeSort())
		}
		if ok {
			opts = append(opts, opt)
		}
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

//replaceObjectID replace filter _id string to ObjectID
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
