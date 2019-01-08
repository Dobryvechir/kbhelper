// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"io/ioutil"
)

func isCurrentFormatJson(data []byte) bool {
	first := getFirstNonSpaceByte(data)
	last := getLastNonSpaceByte(data)
	return first == '[' && last == ']' || first == '{' && last == '}'
}

func convertYamlToJsonOrBack(src string, indentation int) {
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
		panic("Fatal error")
	}
	ext := "json"
	var newData []byte
	if isCurrentFormatJson(data) {
		dvEntry, err := readJsonAsEntries(data)
		if err != nil {
			panic(err.Error())
		}
		newData = dvEntry.PrintToYaml(indentation)
		ext = "yaml"
	} else {
		dvEntry, err := readYamlAsEntries(data)
		if err != nil {
			panic(err.Error())
		}
		newData = dvEntry.PrintToJson(indentation)
	}
	newSrc := changeExtension(src, ext)
	e = ioutil.WriteFile(newSrc, newData, 0644)
	if e != nil {
		fmt.Printf("Cannot write file %s: %s\n", newSrc, e.Error())
		panic("Fatal error")
	}
	fmt.Printf("Converted to %s", newSrc)
}

func getFirstNonSpaceByte(data []byte) byte {
	l := len(data)
	for i := 0; i < l; i++ {
		if data[i] > 32 {
			return data[i]
		}
	}
	return 0
}

func getLastNonSpaceByte(data []byte) byte {
	l := len(data)
	for i := l - 1; i >= 0; i-- {
		if data[i] > 32 {
			return data[i]
		}
	}
	return 0
}
