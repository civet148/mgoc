package mgoc

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/civet148/log"
	"reflect"
	"strconv"
	"strings"
)

const (
	TAG_NAME_BSON = "bson"
)

const (
	TAG_VALUE_NULL   = ""
	TAG_VALUE_IGNORE = "-" //ignore
)

type ModelReflector struct {
	value  interface{}            //model value
	engine *Engine                //database engine
	dict   map[string]interface{} //dictionary of structure tag and value
}

type Fetcher struct {
	count     int               //column count in db table
	cols      []string          //column names in db table
	types     []*sql.ColumnType //column types in db table
	arrValues [][]byte          //value slice
	mapValues map[string]string //value map
	arrIndex  int               //fetch index
}

func newReflector(e *Engine, v interface{}) *ModelReflector {

	return &ModelReflector{
		value:  v,
		engine: e,
		dict:   make(map[string]interface{}),
	}
}

// parse struct tag and value to map
func (s *ModelReflector) ToMap() map[string]interface{} {
	models := s.value.([]interface{})
	for _, model := range models {
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
				s.parseStructField(typ, val, TAG_VALUE_NULL)
			}
		case reflect.Slice:
			{
				typ = val.Type().Elem()
				val = reflect.New(typ).Elem()
				s.parseStructField(typ, val, TAG_VALUE_NULL)
			}
		case reflect.Map:
			{
				if v, ok := s.value.(*map[string]interface{}); ok {
					s.dict = *v
					break
				}
				if v, ok := s.value.(map[string]interface{}); ok {
					s.dict = v
					break
				}
				if v, ok := s.value.(*map[string]string); ok {
					s.dict = s.convertMapString(*v)
					break
				}
				if v, ok := s.value.(map[string]string); ok {
					s.dict = s.convertMapString(v)
					break
				}
			}
		default:
			log.Warnf("kind [%v] not support yet", typ.Kind())
		}
	}
	log.Json("dictionary", s.dict)
	return s.dict
}

func (s *ModelReflector) convertMapString(ms map[string]string) (mi map[string]interface{}) {
	mi = make(map[string]interface{}, 10)
	for k, v := range ms {
		mi[k] = v
	}
	return
}

// get struct field's tag value
func (s *ModelReflector) getTag(sf reflect.StructField, tagName string) (strValue string, ignore bool) {

	strValue = sf.Tag.Get(tagName)
	if strValue == TAG_VALUE_IGNORE {
		return "", true
	}
	return
}

// parse struct fields
func (s *ModelReflector) parseStructField(typ reflect.Type, val reflect.Value, tagParent string) {
	kind := typ.Kind()
	if kind == reflect.Struct {
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
			tagVal, ignore := s.getTag(typField, TAG_NAME_BSON)
			if ignore {
				continue
			}
			if tagParent != "" {
				tagVal = fmt.Sprintf("%s.%s", tagParent, tagVal)
			}
			if typField.Type.Kind() == reflect.Struct || typField.Type.Kind() == reflect.Slice || typField.Type.Kind() == reflect.Map {
				s.dict[tagVal] = valField.Interface()
				if typField.Type.Kind() == reflect.Struct {
					s.parseStructField(typField.Type, valField, tagVal)
				}
			} else {
				s.setValueByField(typField, valField, tagVal) // save field tag value and field value to map
			}
		}
	}
}

//trim the field value's first and last blank character and save to map
func (s *ModelReflector) setValueByField(field reflect.StructField, val reflect.Value, tagVal string) {
	_, ignore := s.getTag(field, TAG_NAME_BSON)
	if ignore {
		return
	}
	//parse db、json、protobuf tag
	tagVal = handleTagValue(TAG_NAME_BSON, tagVal)
	if tagVal != "" {
		if d, ok := val.Interface().(driver.Valuer); ok {
			s.dict[tagVal], _ = d.Value()
		} else {
			s.dict[tagVal] = val.Interface()
		}
	}

}

//fetch row data to map
func (e *Engine) fetchToMap(fetcher *Fetcher, arg interface{}) (err error) {

	typ := reflect.TypeOf(arg)

	if typ.Kind() == reflect.Ptr {

		for k, v := range fetcher.mapValues {
			m := *arg.(*map[string]string) //just support map[string]string type
			m[k] = v
		}
	}

	return
}

//fetch row data to struct
func (e *Engine) fetchToStruct(fetcher *Fetcher, typ reflect.Type, val reflect.Value) (err error) {

	if typ.Kind() == reflect.Ptr {

		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() == reflect.Struct {
		NumField := val.NumField()
		for i := 0; i < NumField; i++ {
			typField := typ.Field(i)
			valField := val.Field(i)
			e.fetchToStructField(fetcher, typField.Type, typField, valField)
		}
	}
	return
}

func (e *Engine) fetchToStructField(fetcher *Fetcher, typ reflect.Type, field reflect.StructField, val reflect.Value) {

	//	log.Debugf("typField name [%s] type [%s] valField can addr [%v]", field.Name, field.Type.Kind(), val.CanAddr())
	switch typ.Kind() {
	case reflect.Struct:
		{
			e.fetchToStructAny(fetcher, field, val)
		}
	case reflect.Slice:
		if e.getTagValue(field) != "" {
			_ = e.fetchToJsonObject(fetcher, field, val)
		}
	case reflect.Map: //ignore...
	case reflect.Ptr:
		{
			typElem := field.Type.Elem()
			if val.IsNil() {
				valNew := reflect.New(typElem)
				val.Set(valNew)
			}
			e.fetchToStructField(fetcher, typElem, field, val.Elem())
		}
	default:
		{
			_ = e.setValueByField(fetcher, field, val) //assign value to struct field
		}
	}
}

func (e *Engine) fetchToStructAny(fetcher *Fetcher, field reflect.StructField, val reflect.Value) {
	if _, ok := val.Addr().Interface().(sql.Scanner); ok {
		e.fetchToScanner(fetcher, field, val)
	} else {
		if e.getTagValue(field) != "" {
			_ = e.fetchToJsonObject(fetcher, field, val)
		} else {
			_ = e.fetchToStruct(fetcher, field.Type, val)
		}
	}
}

//json string unmarshal to struct/slice
func (e *Engine) fetchToJsonObject(fetcher *Fetcher, field reflect.StructField, val reflect.Value) (err error) {
	//优先给有db标签的成员变量赋值
	strDbTagVal := e.getTagValue(field)
	if strDbTagVal == TAG_VALUE_IGNORE {
		return
	}

	if v, ok := fetcher.mapValues[strDbTagVal]; ok {
		vp := val.Addr()
		if strings.TrimSpace(v) != "" {
			if err = json.Unmarshal([]byte(v), vp.Interface()); err != nil {
				return log.Errorf("json.Unmarshal [%s] error [%s]", v, err)
			}
		} else {
			//if struct field is a slice type and content is nil make space for it
			if field.Type.Kind() == reflect.Slice {
				val.Set(reflect.MakeSlice(field.Type, 0, 0))
			}
		}
	}
	return
}

//fetch to struct object by customize scanner
func (e *Engine) fetchToScanner(fetcher *Fetcher, field reflect.StructField, val reflect.Value) {
	//优先给有db标签的成员变量赋值
	strDbTagVal := e.getTagValue(field)
	if strDbTagVal == TAG_VALUE_IGNORE {
		return
	}
	if v, ok := fetcher.mapValues[strDbTagVal]; ok {
		vp := val.Addr()
		d := vp.Interface().(sql.Scanner)
		if v == "" {
			return
		}
		if err := d.Scan(v); err != nil {
			log.Errorf("scan '%v' to scanner [%+v] error [%+v]", v, vp.Interface(), err.Error())
		}
	}
}

func (e *Engine) fetchToBaseType(fetcher *Fetcher, typ reflect.Type, val reflect.Value) (err error) {
	v := fetcher.arrValues[fetcher.arrIndex]
	e.setValue(typ, val, string(v))
	fetcher.arrIndex++
	return
}

func handleTagValue(strTagName, strTagValue string) string {

	if strTagValue == "" {
		return ""
	}

	if strTagName == TAG_NAME_BSON {
		vs := strings.Split(strTagValue, ",")
		strTagValue = vs[0]
	}
	return strTagValue
}

func (e *Engine) getTagValue(sf reflect.StructField) (strValue string) {
	strValue = handleTagValue(TAG_NAME_BSON, sf.Tag.Get(TAG_NAME_BSON))
	if strValue != "" {
		return
	}
	return
}

//按结构体字段标签赋值
func (e *Engine) setValueByField(fetcher *Fetcher, field reflect.StructField, val reflect.Value) (err error) {
	//优先给有db标签的成员变量赋值
	strDbTagVal := e.getTagValue(field)
	if strDbTagVal == TAG_VALUE_IGNORE {
		return
	}
	if v, ok := fetcher.mapValues[strDbTagVal]; ok {
		e.setValue(field.Type, val, v)
	}
	return
}

//将string存储的值赋值到变量
func (e *Engine) setValue(typ reflect.Type, val reflect.Value, v string) {
	switch typ.Kind() {
	case reflect.Struct:
		s, ok := val.Addr().Interface().(sql.Scanner)
		if !ok {
			log.Warnf("struct type %s not implement sql.Scanner interface", typ.Name())
			return
		}
		if err := s.Scan(v); err != nil {
			log.Fatalf("scan value %s to sql.Scanner implement object error [%s]", v, err)
		}
	case reflect.String:
		val.SetString(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, _ := strconv.ParseInt(v, 10, 64)
		val.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, _ := strconv.ParseUint(v, 10, 64)
		val.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, _ := strconv.ParseFloat(v, 64)
		val.SetFloat(i)
	case reflect.Bool:
		i, _ := strconv.ParseUint(v, 10, 64)
		val.SetBool(true)
		if i == 0 {
			val.SetBool(false)
		}
	case reflect.Ptr:
		typ = typ.Elem()
		//val = val.Elem()
		e.setValue(typ, val, v)
	default:
		log.Fatalf("can't assign value [%v] to variant type [%v]", v, typ.Kind())
		return
	}
}
