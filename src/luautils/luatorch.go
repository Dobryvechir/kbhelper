package main

import (
	"errors"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"strings"
)

type torchProcessor struct {
	t7Reader   func(lf *LuaFileReader, obj *LuaObject, info *torchClassInfo) map[interface{}]interface{}
	t7Writer   func(o *LuaObject, lf *LuaFileWriter, info *torchClassInfo)
	jsonReader func(info *torchClassInfo, luaObj *LuaObject, keyMap map[string]*dvjson.DvFieldInfo) error
	jsonWriter func(o *LuaObject, w *dvjson.JsonWriter, fullInfo bool) bool
}

func checkCudaRemoval(lf *LuaFileReader, obj *LuaObject, info *torchClassInfo) {
	if info.isCuda && lf.noGpu {
		if info.equivalent != "" {
			obj.ClassName = info.equivalent
		} else {
			obj.ClassName = strings.Replace(obj.ClassName, "torch.cuda.", "torch.", 0)
		}
	}
}

func torchTensorT7Reader(lf *LuaFileReader, obj *LuaObject, info *torchClassInfo) map[interface{}]interface{} {
	checkCudaRemoval(lf, obj, info)
	dimAmount := lf.ReadInt()
	dimensions := make(map[interface{}]interface{}, dimAmount)
	for i := 0; i < dimAmount; i++ {
		dimensions[i+1] = lf.ReadLong()
	}
	subArrayDimensions := make(map[interface{}]interface{}, dimAmount+1)
	if dimAmount == 0 {
		subArrayDimensions[1] = int64(lf.ReadInt())
	} else {
		subArrayAmount := dimAmount + 1
		for i := 0; i < subArrayAmount; i++ {
			subArrayDimensions[i+1] = lf.ReadLong()
		}
	}
	return map[interface{}]interface{}{
		"dimAmount":  dimAmount,
		"dimensions": dimensions,
		"subAmounts": subArrayDimensions,
	}
}

func torchStorageT7Reader(lf *LuaFileReader, obj *LuaObject, info *torchClassInfo) map[interface{}]interface{} {
	checkCudaRemoval(lf, obj, info)
	amount := lf.ReadLong()
	data := make(map[interface{}]interface{}, amount)
	res := map[interface{}]interface{}{
		"amount": amount,
		"data":   data,
	}
	var i int64
	switch info.kind {
	case NumberIsInt:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i+1] = int32(lf.ReadUInt(size))
		}
	case NumberIsLong:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i+1] = int64(lf.ReadUInt(size))
		}
	case NumberIsShort:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i+1] = int16(lf.ReadUInt(size))
		}
	case NumberIsChar:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i+1] = int8(lf.ReadUInt(size))
		}
	case NumberIsByte:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i+1] = uint8(lf.ReadUInt(size))
		}
	case NumberIsDouble:
		for i = 0; i < amount; i++ {
			data[i+1] = lf.ReadDouble()
		}
	case NumberIsFloat:
		for i = 0; i < amount; i++ {
			data[i+1] = lf.ReadFloat()
		}
	case NumberIsHalf:
		for i = 0; i < amount; i++ {
			data[i+1] = lf.ReadHalfFloat()
		}
	}
	return res
}

func checkNonStandardSubAmounts(dimensions interface{}, subAmounts interface{}) bool {
	dims, ok := dimensions.(map[interface{}]interface{})
	if !ok {
		return true
	}
	subs, ok1 := subAmounts.(map[interface{}]interface{})
	if !ok1 {
		return true
	}
	n := len(dims)
	m := len(subs)
	if m != n+1 {
		return true
	}
	if n == 0 {
		return subs[1] != 0
	}
	val, ok := subs[m]
	if !ok {
		val = subs[int64(m)]
	}
	if val != int64(1) && val != int(1) {
		return false
	}
	m--
	val, ok = subs[m]
	if !ok {
		val = subs[int64(m)]
	}
	if val != int64(1) && val != int(1) {
		return false
	}
	res := int64(1)
	for m--; n > 1; n-- {
		val, ok = subs[m]
		if !ok {
			val = subs[int64(m)]
		}
		resUnit, nok := val.(int64)
		if !nok {
			resUnit = int64(val.(int))
		}
		val, ok = dims[n]
		if !ok {
			val = dims[int64(n)]
		}
		multiUnit, ok := val.(int64)
		if !ok {
			multiUnit = int64(val.(int))
		}
		res *= multiUnit
		if resUnit != res {
			return true
		}
		m--
	}
	return false
}

func torchTensorJsonWriter(o *LuaObject, w *dvjson.JsonWriter, fullInfo bool) bool {
	dimensions := o.UpValues["dimensions"].(map[interface{}]interface{})
	if fullInfo || checkNonStandardSubAmounts(dimensions, o.UpValues["subAmounts"]) {
		return true
	}
	w.PrintKey("dimensions")
	n := len(dimensions)
	w.StartArray()
	for i := 1; i <= n; i++ {
		w.PrintValueSmart(dimensions[i])
	}
	w.EndArray()
	return false
}

func torchStorageJsonWriter(o *LuaObject, w *dvjson.JsonWriter, fullInfo bool) bool {
	if fullInfo {
		return true
	}
	w.PrintKey("amount")
	w.PrintValueSmart(o.UpValues["amount"])
	w.PrintKey("data")
	w.StartArray()
	data := o.UpValues["data"].(map[interface{}]interface{})
	n := len(data)
	for i := 1; i <= n; i++ {
		v, ok := data[i]
		if !ok {
			v = data[int64(i)]
		}
		w.PrintValueSmart(v)
	}
	w.EndArray()
	return false
}

func torchTensorJsonReader(info *torchClassInfo, luaObj *LuaObject, keyMap map[string]*dvjson.DvFieldInfo) error {
	var dimensions map[interface{}]interface{}
	dimInfo, ok := keyMap["dimensions"]
	if ok {
		res, ok1 := dimInfo.ConvertValueToInterface()
		if !ok1 {
			return errors.New("dimensions field is corrupted")
		}
		result, ok := res.([]interface{})
		if !ok {
			return errors.New("dimensions field must be an array")
		}
		n := len(result)
		dimensions = make(map[interface{}]interface{})
		for i := 0; i < n; i++ {
			dimensions[i+1] = result[i]
		}
		luaObj.UpValues["dimensions"] = dimensions
	} else {
		res, ok := luaObj.UpValues["dimensions"]
		if !ok {
			return errors.New("dimensions are not defined at all")
		}
		dimensions, ok = res.(map[interface{}]interface{})
		if !ok {
			return errors.New("dimensions are corrupted")
		}
	}
	n := len(dimensions)
	luaObj.UpValues["amount"] = n
	_, ok = luaObj.UpValues["subAmounts"]
	if !ok {
		subAmounts := make(map[interface{}]interface{}, n+1)
		if n == 0 {
			subAmounts[1] = 0
		} else {
			subAmounts[n+1] = int64(1)
			subAmounts[n] = int64(1)
			res := int64(1)
			for n > 1 {
				v, ok := dimensions[n].(int)
				vf := int64(v)
				if !ok {
					vf, ok = dimensions[n].(int64)
				}
				if !ok || vf <= 0 {
					return errors.New("incorrect dimensions")
				}
				res *= vf
				n--
				subAmounts[n] = res
			}
		}
		luaObj.UpValues["subAmounts"] = subAmounts
	}
	return nil
}

func torchStorageJsonReader(info *torchClassInfo, luaObj *LuaObject, keyMap map[string]*dvjson.DvFieldInfo) error {
	var data map[interface{}]interface{}
	datInfo, ok := keyMap["data"]
	if ok {
		res, ok1 := datInfo.ConvertValueToInterface()
		if !ok1 {
			return errors.New("data field is corrupted")
		}
		result, ok := res.([]interface{})
		if !ok {
			return errors.New("data field must be an array")
		}
		n := len(result)
		data = make(map[interface{}]interface{})
		for i := 0; i < n; i++ {
			data[i+1] = result[i]
		}
		luaObj.UpValues["data"] = data
	} else {
		res, ok := luaObj.UpValues["data"]
		if !ok {
			return errors.New("data field is not defined at all")
		}
		data, ok = res.(map[interface{}]interface{})
		if !ok {
			return errors.New("data field is corrupted")
		}
	}
	n := len(data)
	luaObj.UpValues["amount"] = n
	return nil
}

func torchTensorT7Writer(o *LuaObject, lf *LuaFileWriter, info *torchClassInfo) {
	d, ok := o.UpValues["dimensions"]
	if !ok {
		lf.SetErrorMessage("dimensions is not defined")
		return
	}
	dimensions, ok := d.(map[interface{}]interface{})
	if !ok {
		lf.SetErrorMessage("dimensions must be an array")
		return
	}
	n := len(dimensions)
	lf.WriteInt(n)
	for i := 0; i < n; i++ {
		v, ok := dimensions[i+1]
		if !ok {
			v = dimensions[int64(i+1)]
		}
		vl, ok := v.(int64)
		if !ok {
			r, ok := v.(int)
			if !ok {
				lf.SetErrorMessage("incorrect dimension type: must be a long integer")
				return
			}
			vl = int64(r)
		}
		lf.WriteLong(vl)
	}
	sub, ok := o.UpValues["subAmounts"]
	if !ok {
		lf.SetErrorMessage("subAmounts must be defined")
		return
	}
	subAmounts, ok := sub.(map[interface{}]interface{})
	if !ok {
		lf.SetErrorMessage("subAmounts are corrupted")
	}
	m := len(subAmounts)
	for i := 0; i < m; i++ {
		v, ok := subAmounts[i+1]
		if !ok {
			v = subAmounts[int64(i+1)]
		}
		vl, ok := v.(int64)
		if !ok {
			r, ok := v.(int)
			if !ok {
				lf.SetErrorMessage("incorrect subAmount type: must be a long integer")
				return
			}
			vl = int64(r)
		}
		lf.WriteLong(vl)
	}
}

func torchStorageT7Writer(o *LuaObject, lf *LuaFileWriter, info *torchClassInfo) {
	d, ok := o.UpValues["data"]
	if !ok {
		lf.SetErrorMessage("storage must have data field")
		return
	}
	data, ok := d.(map[interface{}]interface{})
	if !ok {
		lf.SetErrorMessage("storage data is not an array")
		return
	}
	amount := int64(len(data))
	lf.WriteLong(int64(amount))
	var i int64
	switch info.kind {
	case NumberIsInt, NumberIsLong, NumberIsShort, NumberIsChar, NumberIsByte:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			v, ok := data[i+1]
			if !ok {
				v = data[int(i+1)]
			}
			lf.WriteIntNumber(v, size)
		}
	case NumberIsDouble:
		for i = 0; i < amount; i++ {
			v, ok := data[i+1]
			if !ok {
				v = data[int(i+1)]
			}
			lf.WriteDoubleNumber(v)
		}
	case NumberIsFloat:
		for i = 0; i < amount; i++ {
			v, ok := data[i+1]
			if !ok {
				v = data[int(i+1)]
			}
			lf.WriteFloatNumber(v)
		}
	case NumberIsHalf:
		for i = 0; i < amount; i++ {
			v, ok := data[i+1]
			if !ok {
				v = data[int(i+1)]
			}
			lf.WriteHalfFloat(v)
		}
	}
}

var torchTensor = &torchProcessor{
	t7Reader:   torchTensorT7Reader,
	t7Writer:   torchTensorT7Writer,
	jsonReader: torchTensorJsonReader,
	jsonWriter: torchTensorJsonWriter,
}

var torchStorage = &torchProcessor{
	t7Reader:   torchStorageT7Reader,
	t7Writer:   torchStorageT7Writer,
	jsonReader: torchStorageJsonReader,
	jsonWriter: torchStorageJsonWriter,
}

type torchClassInfo struct {
	unitSize   int
	kind       int
	isCuda     bool
	processor  *torchProcessor
	equivalent string
}

var torchClassMapper = map[string]torchClassInfo{
	"torch.DoubleTensor": {
		unitSize:  8,
		isCuda:    false,
		kind:      NumberIsDouble,
		processor: torchTensor,
	},
	"torch.FloatTensor": {
		unitSize:  4,
		isCuda:    false,
		kind:      NumberIsFloat,
		processor: torchTensor,
	},
	"torch.HalfTensor": {
		unitSize:  2,
		isCuda:    false,
		kind:      NumberIsHalf,
		processor: torchTensor,
	},
	"torch.ByteTensor": {
		unitSize:  1,
		isCuda:    false,
		kind:      NumberIsByte,
		processor: torchTensor,
	},
	"torch.CharTensor": {
		unitSize:  1,
		isCuda:    false,
		kind:      NumberIsChar,
		processor: torchTensor,
	},
	"torch.ShortTensor": {
		unitSize:  2,
		isCuda:    false,
		kind:      NumberIsShort,
		processor: torchTensor,
	},
	"torch.IntTensor": {
		unitSize:  4,
		isCuda:    false,
		kind:      NumberIsInt,
		processor: torchTensor,
	},
	"torch.LongTensor": {
		unitSize:  8,
		isCuda:    false,
		kind:      NumberIsLong,
		processor: torchTensor,
	},
	"torch.DoubleStorage": {
		unitSize:  8,
		isCuda:    false,
		kind:      NumberIsDouble,
		processor: torchStorage,
	},
	"torch.FloatStorage": {
		unitSize:  4,
		isCuda:    false,
		kind:      NumberIsFloat,
		processor: torchStorage,
	},
	"torch.HalfStorage": {
		unitSize:  2,
		isCuda:    false,
		kind:      NumberIsHalf,
		processor: torchStorage,
	},
	"torch.ByteStorage": {
		unitSize:  1,
		isCuda:    false,
		kind:      NumberIsByte,
		processor: torchStorage,
	},
	"torch.CharStorage": {
		unitSize:  1,
		isCuda:    false,
		kind:      NumberIsChar,
		processor: torchStorage,
	},
	"torch.ShortStorage": {
		unitSize:  2,
		isCuda:    false,
		kind:      NumberIsShort,
		processor: torchStorage,
	},
	"torch.IntStorage": {
		unitSize:  4,
		isCuda:    false,
		kind:      NumberIsInt,
		processor: torchStorage,
	},
	"torch.LongStorage": {
		unitSize:  8,
		isCuda:    false,
		kind:      NumberIsLong,
		processor: torchStorage,
	},
	"torch.CudaTensor": {
		unitSize:   4,
		isCuda:     true,
		kind:       NumberIsFloat,
		processor:  torchTensor,
		equivalent: "torch.FloatTensor",
	},
	"torch.cuda.DoubleTensor": {
		unitSize:  8,
		isCuda:    true,
		kind:      NumberIsDouble,
		processor: torchTensor,
	},
	"torch.cuda.FloatTensor": {
		unitSize:  4,
		isCuda:    true,
		kind:      NumberIsFloat,
		processor: torchTensor,
	},
	"torch.cuda.HalfTensor": {
		unitSize:  2,
		isCuda:    true,
		kind:      NumberIsHalf,
		processor: torchTensor,
	},
	"torch.cuda.ByteTensor": {
		unitSize:  1,
		isCuda:    true,
		kind:      NumberIsByte,
		processor: torchTensor,
	},
	"torch.cuda.CharTensor": {
		unitSize:  1,
		isCuda:    true,
		kind:      NumberIsChar,
		processor: torchTensor,
	},
	"torch.cuda.ShortTensor": {
		unitSize:  2,
		isCuda:    true,
		kind:      NumberIsShort,
		processor: torchTensor,
	},
	"torch.cuda.IntTensor": {
		unitSize:  4,
		isCuda:    true,
		kind:      NumberIsInt,
		processor: torchTensor,
	},
	"torch.cuda.LongTensor": {
		unitSize:  8,
		isCuda:    true,
		kind:      NumberIsLong,
		processor: torchTensor,
	},
	"torch.cuda.DoubleStorage": {
		unitSize:  8,
		isCuda:    true,
		kind:      NumberIsDouble,
		processor: torchStorage,
	},
	"torch.CudaStorage": {
		unitSize:   4,
		isCuda:     true,
		kind:       NumberIsFloat,
		processor:  torchStorage,
		equivalent: "torch.FloatStorage",
	},
	"torch.cuda.FloatStorage": {
		unitSize:  4,
		isCuda:    true,
		kind:      NumberIsFloat,
		processor: torchStorage,
	},
	"torch.cuda.HalfStorage": {
		unitSize:  2,
		isCuda:    true,
		kind:      NumberIsHalf,
		processor: torchStorage,
	},
	"torch.cuda.ByteStorage": {
		unitSize:  1,
		isCuda:    true,
		kind:      NumberIsByte,
		processor: torchStorage,
	},
	"torch.cuda.CharStorage": {
		unitSize:  1,
		isCuda:    true,
		kind:      NumberIsChar,
		processor: torchStorage,
	},
	"torch.cuda.ShortStorage": {
		unitSize:  2,
		isCuda:    true,
		kind:      NumberIsShort,
		processor: torchStorage,
	},
	"torch.cuda.IntStorage": {
		unitSize:  4,
		isCuda:    true,
		kind:      NumberIsInt,
		processor: torchStorage,
	},
	"torch.cuda.LongStorage": {
		unitSize:  8,
		isCuda:    true,
		kind:      NumberIsLong,
		processor: torchStorage,
	},
}
