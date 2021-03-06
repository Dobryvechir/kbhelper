package main

import (
	"fmt"
	"strconv"
)

type LuaContext struct {
	BigEndian bool
	Force     bool
	RemoveGpu bool
	FullJson  bool
	Eol       int
}

type LuaResult struct {
	data []interface{}
}

func ReadLuaResultFromT7(fileName string, context *LuaContext) (*LuaResult, error) {
	lf, err := OpenLuaFile(fileName, context)
	if err != nil {
		return nil, err
	}
	defer lf.CloseLuaFile()
	lr := &LuaResult{data: make([]interface{}, 0, 1)}
	for lf.HasMore() {
		r := ReadLuaObject(lf)
		lr.data = append(lr.data, r)
	}
	err = lf.err
	lf.err = nil
	return lr, err
}

func WriteLuaResultToT7(fileName string, lua *LuaResult, context *LuaContext) error {
	lf, err := CreateLuaFile(fileName, context)
	if err != nil {
		return err
	}
	defer lf.CloseLuaFile()
	n := len(lua.data)
	for i := 0; i < n && lf.err == nil; i++ {
		WriteLuaObject(lf, lua.data[i])
	}
	return lf.err
}

func WriteObjectIndexAndCheckRecurring(lf *LuaFileWriter, obj interface{}) bool {
	recurring := !lf.context.Force
	if recurring {
		_, ok := obj.(*dvjson.OrderedMap)
		if ok {
			recurring = false
		} else {
			index, ok := lf.objects[obj]
			if ok {
				lf.WriteInt(index)
				return true
			}
		}
	}
	lf.nWriteObject++
	index := lf.nWriteObject
	lf.WriteInt(index)
	if recurring {
		defer func() {
			// recover from panic if one occured. Set err to nil otherwise.
			if recover() != nil {
				fmt.Println("Info: index recurring was not fully supported")
			}
		}()
		lf.objects[obj] = index
	}
	return false
}

func WriteLuaFunction(lf *LuaFileWriter, obj *LuaObject) {
	lf.WriteLengthAndString(obj.Dumped)
	WriteLuaObject(lf, obj.UpValues)
}

func WriteLuaObject(lf *LuaFileWriter, obj interface{}) {
	typ, subTyp := GetLuaObjectType(obj)
	lf.WriteInt(typ)
	switch typ {
	case TYPE_TORCH:
		if WriteObjectIndexAndCheckRecurring(lf, obj) {
			return
		}
		luaObj := obj.(*LuaObject)
		if luaObj.Version != "" {
			lf.WriteLengthAndString(luaObj.Version)
		}
		lf.WriteLengthAndString(luaObj.ClassName)
		luaObj.PrintDumpForTorchType(lf)
	case TYPE_RECUR_FUNCTION, TYPE_LEGACY_RECUR_FUNCTION:
		if WriteObjectIndexAndCheckRecurring(lf, obj) {
			return
		}
		WriteLuaFunction(lf, obj.(*LuaObject))
	case TYPE_FUNCTION:
		WriteLuaFunction(lf, obj.(*LuaObject))
	case TYPE_TABLE:
		if WriteObjectIndexAndCheckRecurring(lf, obj) {
			return
		}
		lf.WriteInt(subTyp)
		tbl := obj.(*dvjson.OrderedMap)
                n:=tbl.Size()
		for i:=0;i<n;i++ {
                        k,v:=tbl.GetAt(i)
			WriteLuaObject(lf, k)
			WriteLuaObject(lf, v)
		}
	case TYPE_STRING:
		lf.WriteLengthAndString(obj.(string))
	case TYPE_NUMBER:
		var nmb float64
		switch subTyp {
		case NumberIsLong:
			nmb = float64(obj.(int64))
		case NumberIsInt:
			nmb = float64(obj.(int))
		case NumberIsDouble:
			nmb = obj.(float64)
		case NumberIsFloat:
			nmb = float64(obj.(float32))
		}
		lf.WriteDouble(nmb)
	case TYPE_BOOLEAN:
		lf.WriteBoolean(obj.(bool))
	case TYPE_NIL:
		return
	default:
		lf.SetErrorData("Unsupported data to write", obj)
	}
}

func ReadObjectIndexAndCheckRecurring(lf *LuaFileReader) (int, interface{}) {
	index := lf.ReadInt()
	if index <= 0 {
		lf.SetErrorMessage("corrupted recurrence index: "+strconv.Itoa(index), lf.integerSize)
		return -1, nil
	}
	data, ok := lf.objects[index]
	if ok {
		return 0, data
	}
	return index, nil
}

func ReadLuaObject(lf *LuaFileReader) interface{} {
	typ := lf.ReadInt()
	switch typ {
	case TYPE_TORCH:
		index, data := ReadObjectIndexAndCheckRecurring(lf)
		if index <= 0 {
			return data
		}
		return ReadLuaTorch(lf, index)
	case TYPE_TABLE:
		index, data := ReadObjectIndexAndCheckRecurring(lf)
		if index <= 0 {
			return data
		}
		res := dvjson.CreateOrderedMap(n)
		lf.objects[index] = res
		n := lf.ReadInt()
		for i := 0; i < n; i++ {
			k := ReadLuaObject(lf)
			v := ReadLuaObject(lf)
			res.Put(k, v)
		}
		return res
	case TYPE_RECUR_FUNCTION, TYPE_LEGACY_RECUR_FUNCTION:
		index, data := ReadObjectIndexAndCheckRecurring(lf)
		if index <= 0 {
			return data
		}
		return ReadLuaFunction(lf, typ, index)
	case TYPE_FUNCTION:
		return ReadLuaFunction(lf, typ, -3)
	case TYPE_STRING:
		return lf.ReadString()
	case TYPE_NUMBER:
		v := lf.ReadDouble()
		iv := int(v)
		if float64(iv) == v {
			return iv
		}
		return v
	case TYPE_BOOLEAN:
		return lf.ReadBoolean()
	case TYPE_NIL:
		return nil
	default:
		lf.SetErrorMessage("Corrupted file: type code (expected 0..8)="+strconv.Itoa(typ), -lf.integerSize)
		return nil
	}
}
