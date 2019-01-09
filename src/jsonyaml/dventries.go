// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"sort"
)

type DvEntryType int

const (
	DV_ENTRY_ARRAY DvEntryType = iota
	DV_ENTRY_MAP
	DV_ENTRY_STRING
	DV_ENTRY_SIMPLE
)

type DvEntry struct {
	Type        DvEntryType
	StringValue string
	MapValue    map[string]*DvEntry
	ArrayValue  []*DvEntry
	Order       int
	keyValue    string
}

type dvEntrySorter struct {
	dvEntries []*DvEntry
}

func (d *dvEntrySorter) Len() int {
	return len(d.dvEntries)
}

func (d *dvEntrySorter) Swap(i, j int) {
	d.dvEntries[i], d.dvEntries[j] = d.dvEntries[j], d.dvEntries[i]
}

func (d *dvEntrySorter) Less(i, j int) bool {
	orderI := -1
	orderJ := -1
	if d.dvEntries[i] != nil {
		orderI = d.dvEntries[i].Order
	}
	if d.dvEntries[j] != nil {
		orderJ = d.dvEntries[j].Order
	}
	return orderI < orderJ
}

func getSortedDvEntryByOrder(mapValue map[string]*DvEntry) []*DvEntry {
	n := len(mapValue)
	entryList := make([]*DvEntry, n)
	count := 0
	for k, v := range mapValue {
		v.keyValue = k
		entryList[count] = v
		count++
	}
	sort.Stable(&dvEntrySorter{dvEntries: entryList})
	return entryList
}

func (dvEntry *DvEntry) PrintToJsonAtLevel(res []byte, level int, indent int, noIndentAtFirst bool) []byte {
	if dvEntry == nil {
		return res
	}
	n := indent * level
	nextN := n + indent
	nextLevel := level + 1
	var indentBuf, nextIndentBuf []byte
	if indent > 0 {
		indentBuf = make([]byte, n)
		for i := 0; i < n; i++ {
			indentBuf[i] = ' '
		}

		if !noIndentAtFirst {
			res = append(res, indentBuf...)
		}
	}
	switch dvEntry.Type {
	case DV_ENTRY_STRING:
		res = appendJsonEscapedString(res, []byte(dvEntry.StringValue))
	case DV_ENTRY_ARRAY:
		arrayAmount := len(dvEntry.ArrayValue)
		res = append(res, '[')
		if arrayAmount > 0 {
			if indent > 0 {
				res = append(res, byte(10))
				nextIndentBuf = make([]byte, nextN)
				for i := 0; i < nextN; i++ {
					nextIndentBuf[i] = ' '
				}
			}

			for i := 0; i < arrayAmount; i++ {
				if i > 0 {
					res = append(res, ',', byte(10))
				}
				if indent > 0 {
					res = append(res, nextIndentBuf...)
				}
				res = dvEntry.ArrayValue[i].PrintToJsonAtLevel(res, nextLevel, indent, true)
			}
			if indent > 0 {
				res = append(res, byte(10))
				res = append(res, indentBuf...)
			}
		}
		res = append(res, ']')

	case DV_ENTRY_MAP:
		mapAmount := len(dvEntry.MapValue)
		res = append(res, '{')
		if mapAmount > 0 {
			if indent > 0 {
				res = append(res, byte(10))
				nextIndentBuf = make([]byte, nextN)
				for i := 0; i < nextN; i++ {
					nextIndentBuf[i] = ' '
				}
			}
			isNext := false
			entryList := getSortedDvEntryByOrder(dvEntry.MapValue)
			for _, v := range entryList {
				if isNext {
					res = append(res, ',', byte(10))
				} else {
					isNext = true
				}
				if indent > 0 {
					res = append(res, nextIndentBuf...)
				}
				res = appendJsonEscapedString(res, []byte(v.keyValue))
				res = append(res, ':', ' ')
				res = v.PrintToJsonAtLevel(res, nextLevel, indent, true)
			}
			if indent > 0 {
				res = append(res, byte(10))
				res = append(res, indentBuf...)
			}
		}
		res = append(res, '}')

	default:
		res = append(res, []byte(dvEntry.StringValue)...)
	}
	return res
}

func appendJsonEscapedString(res []byte, add []byte) []byte {
	n := len(add)
	res = append(res, '"')
	for i := 0; i < n; i++ {
		b := add[i]
		if b == '"' || b == '\\' {
			res = append(res, '\\')
		}
		res = append(res, b)
	}
	res = append(res, '"')
	return res
}

func isReservedWord(data []byte) bool {
	word := string(data)
	return word == "null" || word == "false" || word == "true"
}

func appendYamlEscapedString(res []byte, add []byte) []byte {
	n := len(add)
	isNumber := true
	isSimple := true
	for i := 0; i < n; i++ {
		b := add[i]
		if !(b >= '0' && b <= '9') {
			if b != '.' && b != '+' && b != 'e' && b != '-' {
				isNumber = false
			}
			if !(b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b > 127 || (i > 0 && i < n-1 && (b == '.' || b == '/' || b == '-'))) {
				isSimple = false
				if !isNumber {
					break
				}
			}
		}
	}
	if isNumber {
		res = append(res, '\'')
		res = append(res, add...)
		res = append(res, '\'')
		return res
	}
	if isSimple && isReservedWord(add) {
		isSimple = false
	}
	if isSimple {
		res = append(res, add...)
		return res
	}
	res = append(res, '"')
	for i := 0; i < n; i++ {
		b := add[i]
		if b == '"' || b == '\\' {
			res = append(res, '\\')
		}
		res = append(res, b)
	}
	res = append(res, '"')
	return res
}

func (dvEntry *DvEntry) PrintToYamlAtLevel(res []byte, level int, indent int, noIndentAtFirst bool) []byte {
	if dvEntry == nil {
		return res
	}
	n := indent * level
	nextLevel := level + 1
	var indentBuf []byte
	indentBuf = make([]byte, n+1)
	indentBuf[0] = byte(10)
	for i := 1; i <= n; i++ {
		indentBuf[i] = ' '
	}
	switch dvEntry.Type {
	case DV_ENTRY_STRING:
		if !noIndentAtFirst {
			res = append(res, indentBuf...)
		}
		res = appendYamlEscapedString(res, []byte(dvEntry.StringValue))
	case DV_ENTRY_ARRAY:
		arrayAmount := len(dvEntry.ArrayValue)
		if arrayAmount > 0 {
			for i := 0; i < arrayAmount; i++ {
				res = append(res, indentBuf...)
				res = append(res, '-', ' ')
				res = dvEntry.ArrayValue[i].PrintToYamlAtLevel(res, nextLevel, indent, true)
			}
		} else {
			res = append(res, '[', ']')
		}
	case DV_ENTRY_MAP:
		mapAmount := len(dvEntry.MapValue)
		if mapAmount > 0 {
			entryList := getSortedDvEntryByOrder(dvEntry.MapValue)
			for _, v := range entryList {
				res = append(res, indentBuf...)
				res = appendYamlEscapedString(res, []byte(v.keyValue))
				res = append(res, ':', ' ')
				res = v.PrintToYamlAtLevel(res, nextLevel, indent, true)
			}
		} else {
			res = append(res, '{', '}')
		}
	default:
		if !noIndentAtFirst {
			res = append(res, indentBuf...)
		}
		res = append(res, []byte(dvEntry.StringValue)...)
	}
	return res
}

func (dvEntry *DvEntry) PrintToJson(indent int) []byte {
	res := make([]byte, 0, 64000)
	return dvEntry.PrintToJsonAtLevel(res, 0, indent, true)
}

func (dvEntry *DvEntry) PrintToYaml(indent int) []byte {
	if indent < 1 {
		indent = 2
	}
	res := make([]byte, 0, 64000)
	res = dvEntry.PrintToYamlAtLevel(res, 0, indent, true)
	n := len(res)
	if n > 0 && res[0] == 10 {
		res = res[1:]
		n--
	}
	if n > 0 && res[n-1] != 10 {
		res = append(res, byte(10))
	}
	return res
}
