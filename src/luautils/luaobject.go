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
	NumberIsDouble = 0
	NumberIsInt    = 1
	NumberIsLong   = 2
	NumberIsFloat  = 3
	NumberIsHalf   = 4
	NumberIsByte   = 5
	NumberIsShort  = 6
	NumberIsChar   = 7
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
	LuaClassName    = "_class"
	LuaVersion      = "version"
	LuaDumped       = "dump"
	LuaUpValues     = "values"
)

func (o *LuaObject) PrintToJson(w *dvjson.JsonWriter) {
	w.StartObject()
	w.PrintKey(LuaSpecialKey)
	if o.TypeIdx == TYPE_TORCH {
		w.PrintValueSmart(o.ClassName)
	} else {
		w.PrintValueInteger(o.TypeIdx)
	}
	fullJson := w.CustomInfo.(*LuaContext).FullJson
	if fullJson {
		w.PrintKey(LuaSpecialIndex)
		w.PrintValueInteger(o.RecIndex)
	}
	if o.TypeIdx == TYPE_TORCH {
		if fullJson || (o.Version != "" && o.Version != "V 1") {
			w.PrintKey(LuaVersion)
			w.PrintValueSmart(o.Version)
		}
		classInfo, ok := torchClassMapper[o.ClassName]
		if ok {
			needFullInfo := classInfo.processor.jsonWriter(o, w, fullJson)
			if needFullInfo {
				w.PrintKey(LuaUpValues)
				WriteLuaObjectInJson(w, o.UpValues)
			}
		} else if o.Dumped != "" {
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
	val, ok := dvjson.GetIntValueFromFieldMap(data, LuaSpecialKey)
	str, ok1 := dvjson.GetStringValueFromFieldMap(data, LuaSpecialKey)
	if !ok && ok1 && str != "" {
		ok = true
		val = TYPE_TORCH
	}
	if !ok || (val != TYPE_TORCH && val != TYPE_FUNCTION && val != TYPE_RECUR_FUNCTION && val != TYPE_LEGACY_RECUR_FUNCTION) {
		return nil
	}
	luaObj := &LuaObject{TypeIdx: val}
	isTorch := val == TYPE_TORCH
	if val, ok = dvjson.GetIntValueFromFieldMap(data, LuaSpecialIndex); ok {
		luaObj.RecIndex = val
	}
	if isTorch {
		if str != "" {
			luaObj.ClassName = str
		} else {
			str, ok1 = dvjson.GetStringValueFromFieldMap(data, LuaClassName)
			if ok1 {
				luaObj.ClassName = str
			}
		}
		if str, ok1 = dvjson.GetStringValueFromFieldMap(data, LuaVersion); ok1 {
			luaObj.Version = str
		} else {
			luaObj.Version = "V 1"
		}
	}
	if str, ok1 = dvjson.GetStringValueFromFieldMap(data, LuaDumped); ok1 {
		luaObj.Dumped = str
	}
	return luaObj
}

func (o *LuaObject) PrintDumpForTorchType(lf *LuaFileWriter) {
	classInfo, ok := torchClassMapper[o.ClassName]
	if o.TypeIdx == TYPE_TORCH && ok {
		classInfo.processor.t7Writer(o, lf, &classInfo)
	} else if o.Dumped == "" {
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
		return TYPE_NUMBER, NumberIsInt
	case int64:
		return TYPE_NUMBER, NumberIsLong
	case bool:
		return TYPE_BOOLEAN, 0
	case float64:
		return TYPE_NUMBER, NumberIsDouble
	case float32:
		return TYPE_NUMBER, NumberIsFloat
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
	res := &LuaObject{
		TypeIdx:  kind,
		RecIndex: index,
		Dumped:   lf.ReadString(),
	}
	obj := ReadLuaObject(lf)
	switch obj.(type) {
	case nil:
	case map[interface{}]interface{}:
		res.UpValues = obj.(map[interface{}]interface{})
	default:
		lf.SetErrorData("corrupted file, expected associated map, but received ", obj, 0)
	}
	if index > 0 {
		lf.objects[index] = res
	}
	return res
}

func LuaGoodVersion(version string) bool {
	n := len(version)
	if n < 2 || version[0] != 'V' {
		return false
	}
	for i := 1; i < n; i++ {
		c := version[i]
		if !(c == ' ' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func ReadLuaTorch(lf *LuaFileReader, index int) *LuaObject {
	res := &LuaObject{
		TypeIdx:  TYPE_TORCH,
		RecIndex: index,
		Version:  lf.ReadString(),
	}
	if LuaGoodVersion(res.Version) {
		res.ClassName = lf.ReadString()
	} else {
		res.ClassName = res.Version
		res.Version = "V 1"
	}
	var obj interface{}
	torchClass, ok := torchClassMapper[res.ClassName]
	if ok {
		obj = torchClass.processor.t7Reader(lf, res, &torchClass)
	} else {
		obj = ReadLuaObject(lf)
	}
	switch obj.(type) {
	case map[interface{}]interface{}:
		res.UpValues = obj.(map[interface{}]interface{})
	default:
		lf.SetErrorData("corrupted file, expected associated map, but received ", obj, 0)
	}
	lf.objects[index] = res
	return res
}
