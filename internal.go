package mgoc

import (
	"fmt"
	"github.com/civet148/log"
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
	if e.engineOption.DatabaseOpt != nil {
		opts = append(opts, e.engineOption.DatabaseOpt)
	}
	engine := &Engine{
		engineOption:    e.engineOption,
		client:          e.client,
		strPkName:       e.strPkName,
		dbTags:          e.dbTags,
		strDatabaseName: strDatabaseName,
		models:          make([]interface{}, 0),
		dict:            make(map[string]interface{}),
		filter:          make(map[string]interface{}),
		db:              e.client.Database(strDatabaseName, opts...),
	}
	return engine.setModel(models...)
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
					if len(e.models) == 0 && len(models) != 0 {//used to fetch records
						e.models = models
					}
				} else {
					for j:= 0; j < val.Len(); j++ {//used to insert/update records
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
