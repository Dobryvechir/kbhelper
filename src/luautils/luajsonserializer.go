package main

import (
	"errors"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"io/ioutil"
)

const (
	MapMixed      = 0
	MapPureArray  = 1
	MapPureObject = 2
	KeyName       = "__lua__key__"
	ValueName     = "__value__"
)

func ReadLuaResultJsonItemGeneral(item *dvjson.DvFieldInfo, context *LuaContext) (interface{}, error) {
	simple, ok := item.ConvertSimpleValueToInterface()
	if ok {
		return simple, nil
	}
	switch item.Kind {
	case dvjson.FIELD_OBJECT:
		return ReadLuaResultJsonObject(item.Fields, context)
	case dvjson.FIELD_ARRAY:
		return ReadLuaResultJsonArray(item.Fields, context)
	}
	return nil, errors.New("error parsing the json array")
}

func ReadLuaResultJsonObject(fields []*dvjson.DvFieldInfo, context *LuaContext) (interface{}, error) {
	keyMap := dvjson.ConvertDvFieldInfoArrayIntoMap(fields)
	luaObj := ReadFromJsonFields(keyMap)
	if luaObj != nil {
		res, err := ReadLuaResultJsonItemGeneral(keyMap[LuaUpValues], context)
		if err != nil {
			return nil, err
		}
		resMap, ok := res.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("expected map, but found simple values for 'values' field ")
		}
		luaObj.UpValues = resMap
		return luaObj, nil
	}
	res := make(map[interface{}]interface{})
	for k, v := range keyMap {
		vf, err := ReadLuaResultJsonItemGeneral(v, context)
		if err != nil {
			return nil, err
		}
		res[k] = vf
	}
	return res, nil
}

func ReadLuaResultJsonArray(fields []*dvjson.DvFieldInfo, context *LuaContext) (interface{}, error) {
	n := len(fields)
	assumeComplex := n != 0
	for i := 0; i < n; i++ {
		d := fields[i]
		if d.Kind != dvjson.FIELD_OBJECT || len(d.Fields) != 2 || string(d.Fields[0].Name) != KeyName || string(d.Fields[1].Name) != ValueName {
			assumeComplex = false
			break
		}
	}
	res := make(map[interface{}]interface{})
	if assumeComplex {
		for i := 0; i < n; i++ {
			d := fields[i].Fields
			k, err := ReadLuaResultJsonItemGeneral(d[0], context)
			if err != nil {
				return nil, err
			}
			v, err1 := ReadLuaResultJsonItemGeneral(d[1], context)
			if err1 != nil {
				return nil, err
			}
			res[k] = v
		}
	} else {
		for i := 0; i < n; i++ {
			v, err := ReadLuaResultJsonItemGeneral(fields[i], context)
			if err != nil {
				return nil, err
			}
			res[i+1] = v
		}
	}
	return res, nil
}

func ReadLuaResultFromJson(fileName string, context *LuaContext) (*LuaResult, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	l := &LuaResult{data: make([]interface{}, 0, 1)}
	parsed, err1 := dvjson.JsonFullParser(b)
	if err1 != nil {
		return nil, err1
	}
	switch parsed.Kind {
	case dvjson.FIELD_ARRAY:
		n := len(parsed.Items)
		for i := 0; i < n; i++ {
			val1, err2 := ReadLuaResultJsonItemGeneral(&parsed.Items[i].DvFieldInfo, context)
			if err2 != nil {
				return nil, err2
			}
			l.data = append(l.data, val1)
		}
	case dvjson.FIELD_OBJECT:
		res, err3 := ReadLuaResultJsonObject(parsed.GetDvFieldInfoHierarchy(), context)
		if err3 != nil {
			return nil, err3
		}
		l.data = append(l.data, res)
	default:
		val, ok := parsed.ConvertSimpleValueToInterface()
		if !ok {
			return nil, errors.New("parsing internal error")
		}
		l.data = append(l.data, val)
	}
	return l, nil
}

func WriteLuaObjectInJson(w *dvjson.JsonWriter, data interface{}) {
	simple := w.PrintValueSmart(data)
	if !simple {
		switch data.(type) {
		case *LuaObject:
			data.(*LuaObject).PrintToJson(w)
		case map[interface{}]interface{}:
			WriteMapInJson(w, data.(map[interface{}]interface{}))
		}
	}
}

func GetInterfaceKind(m map[interface{}]interface{}) int {
	n := len(m)
	if n == 0 {
		return MapPureArray
	}
	res := MapMixed
	for k := range m {
		switch k.(type) {
		case int:
			if res == MapPureObject || k.(int) > n || k.(int) <= 0 {
				return MapMixed
			}
			res = MapPureArray
		case string:
			if res == MapPureArray {
				return MapMixed
			}
			res = MapPureObject
		default:
			return MapMixed
		}
	}
	return res
}

func WriteMapInJson(w *dvjson.JsonWriter, m map[interface{}]interface{}) {
	kind := GetInterfaceKind(m)
	n := len(m)
	switch kind {
	case MapMixed:
		w.StartArray()
		for k, v := range m {
			w.StartObject()
			w.PrintKey(KeyName)
			WriteLuaObjectInJson(w, k)
			w.PrintKey(ValueName)
			WriteLuaObjectInJson(w, v)
			w.EndObject()
		}
		w.EndArray()
	case MapPureArray:
		w.StartArray()
		for i := 1; i <= n; i++ {
			WriteLuaObjectInJson(w, m[i])
		}
		w.EndArray()
	case MapPureObject:
		w.StartObject()
		for k, v := range m {
			w.PrintKey(k.(string))
			WriteLuaObjectInJson(w, v)
		}
		w.EndObject()
	}
}

func WriteLuaResultToJson(fileName string, lua *LuaResult, context *LuaContext) error {
	w, err := dvjson.CreateJsonWriter(fileName, 2, LuaBufSize, context.Eol)
	if err != nil {
		return err
	}
	n := len(lua.data)
	if n == 1 && !dvjson.IsValueSimple(lua.data[0]) {
		WriteLuaObjectInJson(w, lua.data[0])
	} else {
		w.StartArray()
		for i := 0; i < n; i++ {
			WriteLuaObjectInJson(w, lua.data[i])
		}
		w.EndArray()
	}
	w.Close()
	return w.GetError()
}
