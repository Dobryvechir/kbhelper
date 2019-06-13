package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

const (
	LuaBufSize   = 1024000
	LuaBufDirect = 70000
)

type LuaFileWriter struct {
	f            *os.File
	buf          []byte
	cursor       int
	total        int
	bigEn        bool
	err          error
	context      *LuaContext
	objects      map[interface{}]int
	nWriteObject int
}

func CreateLuaFile(fileName string, context *LuaContext) (*LuaFileWriter, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return &LuaFileWriter{
		f:            f,
		buf:          make([]byte, LuaBufSize),
		cursor:       0,
		total:        LuaBufSize,
		bigEn:        context.BigEndian,
		context:      context,
		objects:      make(map[interface{}]int),
		nWriteObject: 0,
	}, nil
}

func (lf *LuaFileWriter) CloseLuaFile() {
	if lf.f != nil {
		lf.FlushLuaFile()
		err := lf.f.Close()
		if err != nil {
			fmt.Printf("Error closing binary file %s\n", err.Error())
		}
		if lf.err != nil {
			fmt.Printf("Error writing binary file %s\n", lf.err.Error())
		}
	}
}

func (lf *LuaFileWriter) FlushLuaFile() {
	if lf.cursor > 0 {
		_, err := lf.f.Write(lf.buf[:lf.cursor])
		if err != nil {
			lf.SetError(err)
		}
		lf.cursor = 0
	}
}

func (lf *LuaFileWriter) SetError(err error) {
	if lf.err == nil {
		lf.err = err
	}
}

func (lf *LuaFileWriter) SetErrorMessage(message string) {
	if lf.err == nil {
		lf.err = errors.New(message)
	}
}

func (lf *LuaFileWriter) SetErrorData(message string, obj interface{}) {
	if lf.err == nil {
		if !strings.Contains(message, "%v") {
			message = message + " %v"
		}
		message = fmt.Sprintf(message, obj)
		lf.err = errors.New(message)
	}
}

func (lf *LuaFileWriter) WriteByteArrayWithLength(b []byte, length int) {
	if lf.cursor+length > lf.total || length > LuaBufDirect {
		lf.FlushLuaFile()
		if length > LuaBufDirect {
			_, err := lf.f.Write(b[:length])
			if err != nil {
				lf.err = err
			}
			length = 0
		}
	}
	if length > 0 {
		d := lf.buf[lf.cursor:]
		for i := 0; i < length; i++ {
			d[i] = b[i]
		}
		lf.cursor += length
	}
}

func (lf *LuaFileWriter) WriteByteArray(b []byte) {
	lf.WriteByteArrayWithLength(b, len(b))
}

func (lf *LuaFileWriter) WriteInt(v int) {
	if lf.cursor+4 > lf.total {
		lf.FlushLuaFile()
	}
	d := lf.buf[lf.cursor:]
	if !lf.bigEn {
		d[0] = byte(v)
		d[1] = byte(v >> 8)
		d[2] = byte(v >> 16)
		d[3] = byte(v >> 24)
	} else {
		d[3] = byte(v)
		d[2] = byte(v >> 8)
		d[1] = byte(v >> 16)
		d[0] = byte(v >> 24)
	}
	lf.cursor += 4
}

func (lf *LuaFileWriter) WriteBoolean(v bool) {
	k := 0
	if v {
		k = 1
	}
	lf.WriteInt(k)
}

func (lf *LuaFileWriter) WriteLong(v int64) {
	if lf.cursor+8 > lf.total {
		lf.FlushLuaFile()
	}
	d := lf.buf[lf.cursor:]
	if !lf.bigEn {
		d[0] = byte(v)
		d[1] = byte(v >> 8)
		d[2] = byte(v >> 16)
		d[3] = byte(v >> 24)
		d[4] = byte(v >> 32)
		d[5] = byte(v >> 40)
		d[6] = byte(v >> 48)
		d[7] = byte(v >> 56)
	} else {
		d[7] = byte(v)
		d[6] = byte(v >> 8)
		d[5] = byte(v >> 16)
		d[4] = byte(v >> 24)
		d[3] = byte(v >> 32)
		d[2] = byte(v >> 40)
		d[1] = byte(v >> 48)
		d[0] = byte(v >> 56)
	}
	lf.cursor += 8
}

func (lf *LuaFileWriter) WriteULong(v uint64) {
	if lf.cursor+8 > lf.total {
		lf.FlushLuaFile()
	}
	d := lf.buf[lf.cursor:]
	if !lf.bigEn {
		d[0] = byte(v)
		d[1] = byte(v >> 8)
		d[2] = byte(v >> 16)
		d[3] = byte(v >> 24)
		d[4] = byte(v >> 32)
		d[5] = byte(v >> 40)
		d[6] = byte(v >> 48)
		d[7] = byte(v >> 56)
	} else {
		d[7] = byte(v)
		d[6] = byte(v >> 8)
		d[5] = byte(v >> 16)
		d[4] = byte(v >> 24)
		d[3] = byte(v >> 32)
		d[2] = byte(v >> 40)
		d[1] = byte(v >> 48)
		d[0] = byte(v >> 56)
	}
	lf.cursor += 8
}

func (lf *LuaFileWriter) WriteDouble(v float64) {
	lf.WriteULong(math.Float64bits(v))
}

func (lf *LuaFileWriter) WriteLengthAndString(data string) {
	bf:=[]byte(data)
	length:=len(bf)
	lf.WriteInt(length)
	lf.WriteByteArrayWithLength(bf, length)
}