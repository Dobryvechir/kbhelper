package main

import (
	"errors"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"strings"
)

type torchProcessor struct {
	t7Reader   func(lf *LuaFileReader, obj *LuaObject, info *torchClassInfo) *dvjson.OrderedMap
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

func torchTensorT7Reader(lf *LuaFileReader, obj *LuaObject, info *torchClassInfo) *dvjson.OrderedMap {
	checkCudaRemoval(lf, obj, info)
	dimAmount := lf.ReadInt()
	dimensions := dvjson.CreateOrderedMap(dimAmount)
	for i := 0; i < dimAmount; i++ {
		dimensions.Put(i+1, lf.ReadLong())
	}
	subArrayDimensions := dvjson.CreateOrderedMap(dimAmount+1)
	if dimAmount == 0 {
		subArrayDimensions.Put(1, int64(lf.ReadInt()))
	} else {
		subArrayAmount := dimAmount + 1
		for i := 0; i < subArrayAmount; i++ {
			subArrayDimensions.Put(i+1, lf.ReadLong())
		}
	}
	return dvjson.CreateOrderedMapByMap(map[interface{}]interface{}{
		"dimAmount":  dimAmount,
		"dimensions": dimensions,
		"subAmounts": subArrayDimensions,
	})
}

func torchStorageT7Reader(lf *LuaFileReader, obj *LuaObject, info *torchClassInfo) *dvjson.OrderedMap {
	checkCudaRemoval(lf, obj, info)
	amount := lf.ReadLong()
	data := make([]interface,amount)
	var i int64
	switch info.kind {
	case NumberIsInt:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i]= int32(lf.ReadUInt(size)
		}
	case NumberIsLong:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i] = int64(lf.ReadUInt(size))
		}
	case NumberIsShort:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i] = int16(lf.ReadUInt(size))
		}
	case NumberIsChar:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i] = int8(lf.ReadUInt(size))
		}
	case NumberIsByte:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			data[i] = uint8(lf.ReadUInt(size))
		}
	case NumberIsDouble:
		for i = 0; i < amount; i++ {
			data[i] = lf.ReadDouble()
		}
	case NumberIsFloat:
		for i = 0; i < amount; i++ {
			data[i] = lf.ReadFloat()
		}
	case NumberIsHalf:
		for i = 0; i < amount; i++ {
			data[i] = lf.ReadHalfFloat()
		}
	}
	res := dvjson.CreateOrderedMapByMap(map[interface{}]interface{}{
		"amount": amount,
		"data":   dvjson.CreateOrderedMapByArray(data, 1),
	})
	return res
}

func checkNonStandardSubAmounts(dimensions interface{}, subAmounts interface{}) bool {
	dims, ok := dimensions.(*dvjson.OrderedMap)
	if !ok {
		return true
	}
	subs, ok1 := subAmounts.(*dvjson.OrderedMap)
	if !ok1 {
		return true
	}
	n := dims.Size()
	m := subs.Size()
	if m != n+1 {
		return true
	}
	if n == 0 {
                r,ok:= subs.GetInt64ByInt64Key(1)
		return !ok || r!= 0
	}
	val, ok := subs.GetInt64ByInt64Key(m)
	if !ok || val!=int64(1) {
		return true		
	}
	m--
	val, ok = subs.GetInt64ByInt64Key(m)
	if !ok || val!=int64(1) {
		return true
	}
	res := int64(1)
	for m--; n > 1; n-- {
		val, ok = subs.GetInt64ByInt64Key(m)
                if !ok {
                        return true
                }
		multiUnit, ok := dims.GetInt64ByInt64Key(n)
		res *= multiUnit
		if resUnit != res || !ok {
			return true
		}
		m--
	}
	return false
}

func torchTensorJsonWriter(o *LuaObject, w *dvjson.JsonWriter, fullInfo bool) bool {
	dimensions := o.UpValues["dimensions"].(*dvjson.OrderedMap)
	if fullInfo || checkNonStandardSubAmounts(dimensions, o.UpValues["subAmounts"]) {
		return true
	}
	w.PrintKey("dimensions")
	n := dimensions.Size()
	w.StartArray()
	for i := 1; i <= n; i++ {
		w.PrintValueSmart(dimensions.GetByInt64Key(int64(i)))
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
	data := o.UpValues["data"].(*dvjson.OrderedMap)
	n := len(data)
	for i := 1; i <= n; i++ {
		v, ok := data.GetByInt64Key(int64(i))
		w.PrintValueSmart(v)
	}
	w.EndArray()
	return false
}

func torchTensorJsonReader(info *torchClassInfo, luaObj *LuaObject, keyMap map[string]*dvjson.DvFieldInfo) error {
	var dimensions *dvjson.OrderedMap
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
		dimensions = dvjson.CreateOrderedMap(n)
		for i := 0; i < n; i++ {
			dimensions.Put(i+1, result[i])
		}
		luaObj.UpValues.Put("dimensions", dimensions)
	} else {
		res, ok := luaObj.UpValues.Get("dimensions")
		if !ok {
			return errors.New("dimensions are not defined at all")
		}
		dimensions, ok = res.(*dvjson.OrderedMap)
		if !ok {
			return errors.New("dimensions are corrupted")
		}
	}
	n := dimensions.Size()
	luaObj.UpValues.Put("amount", n)
	_, ok = luaObj.UpValues.Get("subAmounts")
	if !ok {
		subAmounts := dvjson.CreateOrderedMap(n+1)
		if n == 0 {
			subAmounts.Put(1, 0)
		} else {
			subAmounts.Put(n+1, int64(1))
			subAmounts.Put(n, int64(1))
			res := int64(1)
			for n > 1 {
				vf, ok := dimensions.GetInt64ByInt64Key(n)
				if !ok || vf <= 0 {
					return errors.New("incorrect dimensions")
				}
				res *= vf
				n--
				subAmounts.Put(n, res)
			}
		}
		luaObj.UpValues.Put("subAmounts", subAmounts)
	}
	return nil
}

func torchStorageJsonReader(info *torchClassInfo, luaObj *LuaObject, keyMap map[string]*dvjson.DvFieldInfo) error {
	var data *dvjson.OrderedMap
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
		data = make([]interface{})
		for i := 0; i < n; i++ {
			data[i] = result[i]
		}
		luaObj.UpValues.Put("data", dvjson.CreateOrderedMapByArray(data,1))
	} else {
		res, ok := luaObj.UpValues.Get("data")
		if !ok {
			return errors.New("data field is not defined at all")
		}
		data, ok = res.(*dvjson.OrderedMap)
		if !ok {
			return errors.New("data field is corrupted")
		}
	}
	n := data.Size()
	luaObj.UpValues.Put("amount", n)
	return nil
}

func torchTensorT7Writer(o *LuaObject, lf *LuaFileWriter, info *torchClassInfo) {
	d, ok := o.UpValues.Get("dimensions")
	if !ok {
		lf.SetErrorMessage("dimensions is not defined")
		return
	}
	dimensions, ok := d.(*dvjson.OrderedMap)
	if !ok {
		lf.SetErrorMessage("dimensions must be an array")
		return
	}
	n := dimensions.Size()
	lf.WriteInt(n)
	for i := 0; i < n; i++ {
		v, ok := dimensions.GetInt64ByInt64Key(i+1)
		if !ok {
				lf.SetErrorMessage("incorrect dimension type: must be a long integer")
				return
		}
		lf.WriteLong(v)
	}
	sub, ok := o.UpValues.Get("subAmounts")
	if !ok {
		lf.SetErrorMessage("subAmounts must be defined")
		return
	}
	subAmounts, ok := sub.(*dvjson.OrderedMap)
	if !ok {
		lf.SetErrorMessage("subAmounts are corrupted")
                return
	}
	m := subAmounts.Size()
	for i := 0; i < m; i++ {
		v, ok := subAmounts.GetInt64ByInt64Key(i+1)
		if !ok {
				lf.SetErrorMessage("incorrect subAmount type: must be a long integer")
				return
		}
		lf.WriteLong(v)
	}
}

func torchStorageT7Writer(o *LuaObject, lf *LuaFileWriter, info *torchClassInfo) {
	d, ok := o.UpValues.Get("data")
	if !ok {
		lf.SetErrorMessage("storage must have data field")
		return
	}
	data, ok := d.(*dvjson.OrderedMap)
	if !ok {
		lf.SetErrorMessage("storage data is not an array")
		return
	}
	amount := int64(data.Size())
	lf.WriteLong(amount)
	var i int64
	switch info.kind {
	case NumberIsInt, NumberIsLong, NumberIsShort, NumberIsChar, NumberIsByte:
		size := info.unitSize
		for i = 0; i < amount; i++ {
			v, ok := data.GetByInt64Key(i+1)
			lf.WriteIntNumber(v, size)
		}
	case NumberIsDouble:
		for i = 0; i < amount; i++ {
			v, ok := data.GetByInt64Key(i+1)
			if !ok {
			    lf.SetErrorMessage("data is corrupted")
			}
			lf.WriteDoubleNumber(v)
		}
	case NumberIsFloat:
		for i = 0; i < amount; i++ {
			v, ok := data.GetByInt64Key(i+1)
			if !ok {
			    lf.SetErrorMessage("data is corrupted")
			}
			lf.WriteFloatNumber(v)
		}
	case NumberIsHalf:
		for i = 0; i < amount; i++ {
			v, ok := data.GetByInt64Key(i+1)
			if !ok {
			    lf.SetErrorMessage("data is corrupted")
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
