// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func normalizeForLinux(buf []byte) ([]byte, bool) {
	change := false
	l := len(buf)
	p := 0
	for i := 0; i < l; i++ {
		if buf[i] != 13 {
			buf[p] = buf[i]
			p++
		}
	}
	if p > 0 && buf[p-1] != 10 {
		if p == l {
			buf = append(buf, 10)
			return buf, true
		}
		buf[p] = 10
		p++
		change = true
	}
	if p < l {
		buf = buf[:p]
		change = true
	}
	return buf, change
}

func normalizeFileForLinux(src string) ([]byte, bool) {
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
		panic("Fatal error")
	}
	return normalizeForLinux(data)
}

func removeCRLF(src string) {
	data, change := normalizeFileForLinux(src)
	if !change {
		fmt.Printf("File %s is already normal for Linux", src)
	} else {
		e := ioutil.WriteFile(src, data, 0644)
		if e != nil {
			fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
			panic("Fatal error")
		}
	}
}

func checkAlreadyPresentLine(buf []byte, line string) bool {
	return strings.Index(string(buf), line) >= 0
}

func addNonRepeatedLine(src string, line string) {
	data, _ := normalizeFileForLinux(src)
	byteLine := []byte(line + "\n")
	if checkAlreadyPresentLine(data, line) {
		fmt.Printf("Line %s is already present", line)
	} else {
		data = append(data, byteLine...)
		e := ioutil.WriteFile(src, data, 0644)
		if e != nil {
			fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
			panic("Fatal error")
		}
	}
}
