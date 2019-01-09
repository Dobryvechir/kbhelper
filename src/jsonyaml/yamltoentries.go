// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
)

func readYamlValue(data []byte, pos int) (res string, nextPos int, err error, tp DvEntryType) {
	n := len(data)
	b := data[pos]
	if b == '\'' {
		for nextPos = pos + 1; nextPos < n && data[nextPos] != '\''; nextPos++ {
		}
		if nextPos >= n {
			err = fmt.Errorf("Unclosed single quote string %s", getPositionErrorInfo(data, pos))
			return
		}
		res = string(data[pos+1 : nextPos])
		tp = DV_ENTRY_STRING
		nextPos++
		return
	}
	if b == '"' {
		buf := make([]byte, 0, 200)
		for nextPos = pos + 1; nextPos < n && data[nextPos] != '"'; nextPos++ {
			b := data[nextPos]
			if b == '\\' {
				nextPos++
				if nextPos < n {
					buf = append(buf, data[nextPos])
				}
			} else {
				buf = append(buf, b)
			}
		}
		if nextPos >= n {
			err = fmt.Errorf("Unclosed double quote string %s", getPositionErrorInfo(data, pos))
			return
		}
		res = string(buf)
		tp = DV_ENTRY_STRING
		nextPos++
		return
	}
	isNumber := b >= '0' && b <= '9' || b == '-' || b == '+'
	for nextPos = pos; nextPos < n; nextPos++ {
		b := data[nextPos]
		if isNumber && !(b >= '0' && b <= '9' || b == '-' || b == '+' || b == 'e' || b == 'E' || b == '.') {
			isNumber = false
		}
		if b == 10 || b == 13 || b == ':' || b == '|' || b == '"' || b == '[' || b == ']' || b == '{' || b == '}' {
			break
		}
	}
	for ; nextPos > pos && data[nextPos-1] <= 32; nextPos-- {
	}
	res = string(data[pos:nextPos])
	tp = DV_ENTRY_STRING
	if isNumber || res == "false" || res == "true" || res == "null" {
		tp = DV_ENTRY_SIMPLE
	}
	return
}

func readYamlNonEmptyLine(data []byte, pos int) (nxtPos int, indent int) {
	indent = 0
	for prevPos := pos; prevPos > 0 && data[prevPos-1] != 10 && data[prevPos-1] != 13; prevPos-- {
		indent++
	}
	n := len(data)
	for nxtPos = pos; nxtPos < n; nxtPos++ {
		b := data[nxtPos]
		if b == 13 || b == 10 {
			indent = 0
			continue
		}
		if b <= 32 {
			indent++
			continue
		}
		if b == '%' || b == '#' {
			for ; nxtPos < n; nxtPos++ {
				if data[nxtPos] == 13 || data[nxtPos] == 10 {
					break
				}
			}
			indent = 0
			continue
		}
		break
	}
	return
}

func readYamlPartMap(data []byte, pos int, indent int, endChar byte, currentKey string) (*DvEntry, int, error) {
	mapValue := make(map[string]*DvEntry)
	n := len(data)
	for pos < n {
		dvEntry, nxtPos, err := readYamlPart(data, pos, indent+1, endChar)
		if err != nil {
			return nil, n, err
		}
		dvEntry.Order = len(mapValue)
		mapValue[currentKey] = dvEntry
		pos = nxtPos
		nextPos, newIndent := readYamlNonEmptyLine(data, nxtPos)
		if newIndent != indent || nextPos >= n || data[nextPos] == '-' || data[nextPos] == endChar {
			break
		}
		currentKey, nextPos, err, _ = readYamlValue(data, nextPos)
		if err != nil {
			return nil, n, err
		}
		pos, newIndent = readYamlNonEmptyLine(data, nextPos)
		if pos >= n || data[pos] != ':' || newIndent <= indent {
			return nil, n, fmt.Errorf("Expected colon  %s", getPositionErrorInfo(data, pos))
		}
		pos++

	}
	return &DvEntry{Type: DV_ENTRY_MAP, MapValue: mapValue}, pos, nil
}

func readYamlPartArray(data []byte, pos int, indent int, endChar byte) (*DvEntry, int, error) {
	arrayValue := make([]*DvEntry, 0, 20)
	n := len(data)
	for pos < n && data[pos] == '-' {
		dvEntry, nxtPos, err := readYamlPart(data, pos+1, indent+1, endChar)
		if err != nil {
			return nil, n, err
		}
		dvEntry.Order = len(arrayValue)
		arrayValue = append(arrayValue, dvEntry)
		pos = nxtPos
		nextPos, newIndent := readYamlNonEmptyLine(data, nxtPos)
		if newIndent != indent {
			break
		}
		pos = nextPos
	}
	return &DvEntry{Type: DV_ENTRY_ARRAY, ArrayValue: arrayValue}, pos, nil
}

func readYamlPart(data []byte, pos int, indent int, endChar byte) (*DvEntry, int, error) {
	n := len(data)
	for pos < n {
		nxtPos, newIndent := readYamlNonEmptyLine(data, pos)
		if newIndent < indent || nxtPos >= n {
			return nil, n, fmt.Errorf("Uncompleted yaml %s (new indent=%d old indent=%d next position=%d) ", getPositionErrorInfo(data, pos), newIndent, indent, nxtPos)
		}
		b := data[nxtPos]
		switch b {
		case '-':
			return readYamlPartArray(data, nxtPos, newIndent, endChar)
		case '|', '[', ']', '{', '}', '?', ':':
			panic(string(data[nxtPos:nxtPos+1]) + " is unimplemented for yaml yet")
		default:
			str, nextPos, err, tp := readYamlValue(data, nxtPos)
			if err != nil {
				return nil, n, err
			}
			nexterPos, newerIndent := readYamlNonEmptyLine(data, nextPos)
			if newerIndent <= newIndent || nexterPos >= n || data[nexterPos] != ':' {
				return &DvEntry{Type: tp, StringValue: str}, nextPos, nil
			}
			if data[nexterPos] == ':' {
				return readYamlPartMap(data, nexterPos+1, newIndent, endChar, str)
			}
			return nil, n, fmt.Errorf("Unexpected character at %s", getPositionErrorInfo(data, nexterPos))

		}
	}
	return nil, n, fmt.Errorf("Empty definition found at the end %s", getPositionErrorInfo(data, n))
}

func readYamlAsEntries(data []byte) (*DvEntry, error) {
	dvEntry, pos, err := readYamlPart(data, 0, 0, byte(0))
	if err != nil {
		return nil, err
	}
	l := len(data)
	for ; pos < l; pos++ {
		if data[pos] > 32 {
			return nil, fmt.Errorf("Unexpected extra characters %s", getPositionErrorInfo(data, pos))
		}
	}
	return dvEntry, nil
}
