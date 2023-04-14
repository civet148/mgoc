package mgoc

import (
	"database/sql"
	"database/sql/driver"
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
	for k, v := range s.dict {
		value := ConvertValue(k, v)
		s.dict[k] = value
	}
	//log.Json("dictionary", s.dict)
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
func getTagValue(sf reflect.StructField, tagName string) (strValue string, ignore bool) {
	strValue = sf.Tag.Get(tagName)
	if strValue == TAG_VALUE_IGNORE {
		return "", true
	}
	if tagName == TAG_NAME_BSON {
		vs := strings.Split(strValue, ",")
		strValue = vs[0]
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
			tagVal, ignore := getTagValue(typField, TAG_NAME_BSON)
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
	_, ignore := getTagValue(field, TAG_NAME_BSON)
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

//将string存储的值赋值到变量
func setValue(typ reflect.Type, val reflect.Value, v string) {
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
		setValue(typ, val, v)
	default:
		log.Fatalf("can't assign value [%v] to variant type [%v]", v, typ.Kind())
		return
	}
}
