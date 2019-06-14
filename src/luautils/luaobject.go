package main

import (
	"github.com/Dobryvechir/dvserver/src/dvjson"
)

const (
	TYPE_NIL                   = 0
	TYPE_NUMBER                = 1
	TYPE_STRING                = 2
	TYPE_TABLE                 = 3
	TYPE_TORCH                 = 4
	TYPE_BOOLEAN               = 5
	TYPE_FUNCTION              = 6
	TYPE_RECUR_FUNCTION        = 8
	TYPE_LEGACY_RECUR_FUNCTION = 7
)

const (
	NUMBER_DOUBLE = 0
	NUMBER_INT    = 1
	NUMBER_LONG   = 2
)

var typeInfoMap = map[int]string{
	0: "nil",
	1: "number",
	2: "string",
	3: "table",
	4: "torch",
	5: "boolean",
	6: "function",
	7: "lr-function",
	8: "r-function",
}

type LuaObject struct {
	TypeIdx   int
	RecIndex  int
	Dumped    string
	Version   string
	ClassName string
	UpValues  map[interface{}]interface{}
}

const (
	LuaSpecialKey   = "__Lua__Object__"
	LuaSpecialIndex = "index"
	LuaSpecialType  = "_type"
	LuaClassName    = "_class"
	LuaVersion      = "version"
	LuaDumped       = "dump"
	LuaUpValues     = "values"
)

func (o *LuaObject) PrintToJson(w *dvjson.JsonWriter) {
	w.StartObject()
	w.PrintKey(LuaSpecialKey)
	w.PrintValueInteger(o.TypeIdx)
	w.PrintKey(LuaSpecialIndex)
	w.PrintValueInteger(o.RecIndex)
	w.PrintKey(LuaSpecialType)
	w.PrintValueString(typeInfoMap[o.TypeIdx])
	if o.TypeIdx == TYPE_TORCH {
		w.PrintKey(LuaVersion)
		w.PrintValueSmart(o.Version)
		w.PrintKey(LuaClassName)
		w.PrintValueSmart(o.ClassName)
		if o.Dumped != "" {
			w.PrintKey(LuaDumped)
			w.PrintValueSmart(o.Dumped)
		} else {
			w.PrintKey(LuaUpValues)
			WriteLuaObjectInJson(w, o.UpValues)
		}
	} else {
		w.PrintKey(LuaDumped)
		w.PrintValueSmart(o.Dumped)
		w.PrintKey(LuaUpValues)
		WriteLuaObjectInJson(w, o.UpValues)
	}
	w.EndObject()
}

func ReadFromJsonFields(data map[string]*dvjson.DvFieldInfo) *LuaObject {
	val, ok:=dvjson.GetIntValueFromFieldMap(data, LuaSpecialKey)
	if !ok || (val!=TYPE_TORCH && val!=TYPE_FUNCTION && val!=TYPE_RECUR_FUNCTION && val!=TYPE_LEGACY_RECUR_FUNCTION){
		return nil
	}
	luaObj:=&LuaObject{TypeIdx: val}
	if val, ok = dvjson.GetIntValueFromFieldMap(data,LuaSpecialIndex); ok {
		luaObj.RecIndex = val
	}
	str, ok1:=dvjson.GetStringValueFromFieldMap(data, LuaClassName)
	if ok1 {
		luaObj.ClassName = str
	}
	if str, ok1=dvjson.GetStringValueFromFieldMap(data, LuaVersion); ok1 {
		luaObj.Version = str
	}
	if str, ok1=dvjson.GetStringValueFromFieldMap(data, LuaDumped); ok1 {
		luaObj.Dumped = str
	}
	return luaObj
}

func (o *LuaObject) PrintDumpForTorchType(lf *LuaFileWriter) {
	if o.Dumped == "" {
		WriteLuaObject(lf, o.UpValues)
	} else {
		lf.WriteByteArray([]byte(o.Dumped))
	}
}

func GetLuaObjectType(data interface{}) (int, int) {
	switch data.(type) {
	case string:
		return TYPE_STRING, len(data.(string))
	case int:
		return TYPE_NUMBER, NUMBER_INT
	case int64:
		return TYPE_NUMBER, NUMBER_LONG
	case bool:
		return TYPE_BOOLEAN, 0
	case float64:
		return TYPE_NUMBER, NUMBER_DOUBLE
	case nil:
		return TYPE_NIL, 0
	case map[interface{}]interface{}:
		return TYPE_TABLE, len(data.(map[interface{}]interface{}))
	case *LuaObject:
		return data.(*LuaObject).TypeIdx, 0
	}
	return -1, -1
}

func ReadLuaFunction(lf *LuaFileReader, kind int, index int) *LuaObject {
	res:=&LuaObject{
		TypeIdx: kind,
		RecIndex: index,
		Dumped: lf.ReadString(),
	}
	obj:=ReadLuaObject(lf)
	switch obj.(type) {
	case nil:
	case map[interface{}]interface{}:
		res.UpValues = obj.(map[interface{}]interface{})
	default:
		lf.SetErrorData("corrupted file, expected associated map, but received ", obj)
	}
	if index>0 {
		lf.objects[index] = res
	}
	return res
}

func LuaGoodVersion(version string) bool {
	n:=len(version)
	if n<2 || version[0]!='V' {
		return false
	}
	for i:=1;i<n;i++ {
		c:=version[i]
		if !(c=='.' || c>='0' && c<='9') {
			return false
		}
	}
	return true
}

func ReadLuaTorch(lf *LuaFileReader, index int) *LuaObject {
	res:=&LuaObject{
		TypeIdx: TYPE_TORCH,
		RecIndex: index,
		Version: lf.ReadString(),
	}
	if LuaGoodVersion(res.Version) {
		res.ClassName = lf.ReadString()
	} else {
		res.ClassName = res.Version
		res.Version = ""
	}
	obj:=ReadLuaObject(lf)
	switch obj.(type) {
	case map[interface{}]interface{}:
		res.UpValues = obj.(map[interface{}]interface{})
	default:
		lf.SetErrorData("corrupted file, expected associated map, but received ", obj)
	}
	lf.objects[index] = res
	return res
}

