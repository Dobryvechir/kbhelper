// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

func changeExtension(src string, ext string) string {
	lastPos := strings.LastIndex(src, ".")
	if lastPos > 0 {
		slashPos := strings.Index(src[lastPos:], "\\")
		if slashPos < 0 {
			slashPos = strings.Index(src[lastPos:], "/")
		}
		if slashPos < 0 {
			src = src[:lastPos]
		}
	}
	var newName string
	for i := 0; i < 1000000; i++ {
		if i == 0 {
			newName = src + "." + ext
		} else {
			newName = src + strconv.Itoa(i) + "." + ext
		}
		if _, err := os.Stat(newName); os.IsNotExist(err) {
			return newName
		}
	}
	return "tmp." + ext
}

func convertYamlToJsonOrBack(src string, indentation int) {
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
		panic("Fatal error")
	}
	ext := "json"
	var newData []byte
	if dvjson.IsCurrentFormatJson(data) {
		dvEntry, err := dvjson.ReadJsonAsDvFieldInfo(data)
		if err != nil {
			panic(err.Error())
		}
		newData = dvEntry.PrintToYaml(indentation)
		ext = "yaml"
	} else {
		dvEntry, err := dvjson.ReadYamlAsDvFieldInfo(data)
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

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("or jsonyaml <filename in yaml format, to be converted to json format or back> <integer indentation defaults to 2>")
		return
	}
	fileName := os.Args[1]
	indentStr := "2"
	if l >= 3 {
		indentStr = os.Args[2]
	}
	indentation := 2
	if indent, e := strconv.Atoi(indentStr); e == nil && indent >= 0 {
		indentation = indent
	}
	convertYamlToJsonOrBack(fileName, indentation)
}
