package main

import (
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

func ReadLuaResultFromJson(fileName string, context *LuaContext) (*LuaResult, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	l := &LuaResult{data: make([]interface{}, 0, 1)}
	parsed, err1:=dvjson.JsonFullParser(b)
	if err1!=nil {
		return nil, err1
	}
	//TODO!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
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
	if n == 1 {
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
