package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type LuaFileReader struct {
	f           *os.File
	buf         []byte
	cursor      int
	total       int
	bigEn       bool
	noGpu       bool
	err         error
	amountRead  int64
	rest        int64
	integerSize int
	objects     map[int]interface{}
}

func OpenLuaFile(fileName string, context *LuaContext) (*LuaFileReader, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	fileInfo, err1 := os.Stat(fileName)
	if err1 != nil {
		return nil, err
	}
	rest := fileInfo.Size()
	total := LuaBufSize
	if int64(total) > rest {
		total = int(rest)
	}
	buf := make([]byte, total)
	_, err = f.Read(buf)
	if err != nil {
		return nil, err
	}
	return &LuaFileReader{
		f:           f,
		buf:         buf,
		cursor:      0,
		total:       total,
		bigEn:       context.BigEndian,
		noGpu:       context.RemoveGpu,
		rest:        rest - int64(total),
		amountRead:  int64(total),
		integerSize: 4,
		objects:     make(map[int]interface{}),
	}, nil
}

func (lf *LuaFileReader) CloseLuaFile() {
	if lf.f != nil {
		err := lf.f.Close()
		if err != nil {
			fmt.Printf("Error closing binary file: %s\n", err.Error())
		}
		if lf.err != nil {
			fmt.Printf("Error reading binary file: %s\n", lf.err.Error())
		}
	}
}

func (lf *LuaFileReader) ReplenishLuaFile() {
	if lf.rest == 0 {
		return
	}
	b := lf.buf
	i := lf.cursor
	n := lf.total - lf.cursor
	if n > 0 && i > 0 {
		for j := 0; j < n; j++ {
			b[j] = b[i+j]
		}
	}
	lf.cursor = 0
	i = LuaBufSize - n
	if int64(i) > lf.rest {
		i = int(lf.rest)
	}
	lf.total = n + i
	lf.rest -= int64(i)
	_, err := lf.f.Read(b[n:lf.total])
	if err != nil {
		lf.SetError(err, 0)
	} else {
		lf.amountRead += int64(i)
	}
}

func (lf *LuaFileReader) SetError(err error, offset int) {
	lf.SetErrorMessage(err.Error(), offset)
}

func (lf *LuaFileReader) SetErrorMessage(message string, offset int) {
	if lf.err == nil {
		pos := lf.amountRead + int64(lf.cursor+offset-lf.total)
		posInfo := fmt.Sprintf(" at 0x%x", pos)
		lf.err = errors.New(message + posInfo)
	}
	lf.total = lf.cursor
	lf.rest = 0
}

func (lf *LuaFileReader) SetErrorData(message string, data interface{}, offset int) {
	if lf.err == nil {
		if !strings.Contains(message, "%v") {
			message += " %v"
		}
		message = fmt.Sprintf(message, data)
		lf.SetErrorMessage(message, offset)
	}
}

func (lf *LuaFileReader) ReadByteArrayWithLength(b []byte, length int) {
	if lf.cursor+length > lf.total {
		src := lf.buf[lf.cursor:]
		n := lf.total - lf.cursor
		if n > 0 {
			for i := 0; i < n; i++ {
				b[i] = src[i]
			}
		}
		lf.cursor = 0
		lf.total = 0
		part := length - n
		if int64(part) > lf.rest {
			lf.SetErrorMessage("unexpected end of file, expected to read byte buffer", 0)
			for ; n < length; n++ {
				b[n] = 0
			}
			return
		}
		if length > LuaBufDirect {
			_, err := lf.f.Read(b[n:length])
			if err != nil {
				lf.SetError(err, 0)
				for ; n < length; n++ {
					b[n] = 0
				}
				return
			}
			lf.rest -= int64(part)
			lf.amountRead += int64(part)
			lf.ReplenishLuaFile()
			return
		}
		lf.ReplenishLuaFile()
		b = b[n:]
		length -= n
	}
	if length > 0 {
		d := lf.buf[lf.cursor:]
		for i := 0; i < length; i++ {
			b[i] = d[i]
		}
		lf.cursor += length
	}
}

func (lf *LuaFileReader) ReadString() string {
	n := lf.ReadInt()
	if n < 0 {
		lf.SetErrorMessage("read string corrupted because of negative length "+strconv.Itoa(n), -lf.integerSize)
		return ""
	}
	if n == 0 {
		return ""
	}
	return lf.ReadStringDirect(n)
}

func (lf *LuaFileReader) ReadStringDirect(length int) string {
	if lf.cursor+length > lf.total {
		lf.ReplenishLuaFile()
	}
	if lf.cursor+length > lf.total {
		bf := make([]byte, length)
		lf.ReadByteArrayWithLength(bf, length)
		if lf.err != nil {
			return ""
		}
		return string(bf)
	}
	res := string(lf.buf[lf.cursor : lf.cursor+length])
	lf.cursor += length
	return res
}

func (lf *LuaFileReader) ReadByteArray(b []byte) {
	lf.ReadByteArrayWithLength(b, len(b))
}

func (lf *LuaFileReader) ReadUInt(size int) uint64 {
	if lf.cursor+size > lf.total {
		lf.ReplenishLuaFile()
	}
	if lf.cursor+size > lf.total {
		lf.SetErrorMessage("unexpected end of data, expected to read integer("+strconv.Itoa(lf.integerSize)+" bytes)", 0)
		return 0
	}
	d := lf.buf[lf.cursor:]
	lf.cursor += size
	switch size {
	case 1:
		return uint64(d[0])
	case 2:
		if !lf.bigEn {
			return uint64(d[0]) | uint64(d[1])<<8
		}
		return uint64(d[1]) | uint64(d[0])<<8
	case 4:
		if !lf.bigEn {
			return uint64(d[0]) | uint64(d[1])<<8 | uint64(d[2])<<16 | uint64(d[3])<<24
		}
		return uint64(d[3]) | uint64(d[2])<<8 | uint64(d[1])<<16 | uint64(d[0])<<24
	}
	if !lf.bigEn {
		return uint64(d[0]) | uint64(d[1])<<8 | uint64(d[2])<<16 | uint64(d[3])<<24 | uint64(d[4])<<32 | uint64(d[5])<<40 | uint64(d[6])<<48 | uint64(d[7])<<56
	}
	return uint64(d[7]) | uint64(d[6])<<8 | uint64(d[5])<<16 | uint64(d[4])<<24 | uint64(d[3])<<32 | uint64(d[2])<<40 | uint64(d[1])<<48 | uint64(d[0])<<56
}

func (lf *LuaFileReader) ReadInt() int {
	return int(lf.ReadUInt(lf.integerSize))
}

func (lf *LuaFileReader) ReadBoolean() bool {
	k := lf.ReadInt()
	return k == 1
}

func (lf *LuaFileReader) ReadULong() uint64 {
	if lf.cursor+8 > lf.total {
		lf.ReplenishLuaFile()
	}
	if lf.cursor+8 > lf.total {
		lf.SetErrorMessage("unexpected end of data, expected to read long(8 bytes)", 0)
		return 0
	}
	d := lf.buf[lf.cursor:]
	lf.cursor += 8
	if !lf.bigEn {
		return uint64(d[0]) | uint64(d[1])<<8 | uint64(d[2])<<16 | uint64(d[3])<<24 | uint64(d[4])<<32 | uint64(d[5])<<40 | uint64(d[6])<<48 | uint64(d[7])<<56
	}
	return uint64(d[7]) | uint64(d[6])<<8 | uint64(d[5])<<16 | uint64(d[4])<<24 | uint64(d[3])<<32 | uint64(d[2])<<40 | uint64(d[1])<<48 | uint64(d[0])<<56
}

func (lf *LuaFileReader) ReadLong() int64 {
	return int64(lf.ReadULong())
}

func (lf *LuaFileReader) ReadDouble() float64 {
	return math.Float64frombits(lf.ReadULong())
}

func (lf *LuaFileReader) ReadFloat() float32 {
	return math.Float32frombits(uint32(lf.ReadUInt(4)))
}

func (lf *LuaFileReader) ReadHalfFloat() float32 {
	v := int32(lf.ReadUInt(2))
	var f float32 = 0
	if v != 0 {
		var sign = (v & 0x8000) != 0
		var degreeMinus10 = (v&0x7c00)>>10 - 0xf - 10
		var mantissa = (v & 0x3ff) | 0x400
		f = float32(mantissa)
		if degreeMinus10 > 0 {
			for ; degreeMinus10 > 0; degreeMinus10-- {
				f = f * 2
			}
		} else if degreeMinus10 < 0 {
			for ; degreeMinus10 < 0; degreeMinus10++ {
				f = f / 2
			}
		}
		if sign {
			f = -f
		}
	}
	return f
}

func (lf *LuaFileReader) HasMore() bool {
	return lf.err == nil && (lf.cursor < lf.total || lf.rest > 0)
}
