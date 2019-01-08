// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"os"
	"strconv"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

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
