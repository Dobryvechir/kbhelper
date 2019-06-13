package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args
	l := len(args)
	if l < 2 {
		fmt.Println("<input file> [<output file>] [options:B-big endian F-force (not referenced) G-remove GPU models L-linux eol A-Apple eol]")
		return
	}
	input := args[1]
	inputJson := strings.HasSuffix(input, ".json")
	inputT7 := strings.HasSuffix(input, ".t7")
	if !inputJson && !inputT7 {
		fmt.Println("Only .json or .t7 extensions are supported")
		return
	}
	output := ""
	if l > 2 {
		output = args[2]
	} else {
		if inputJson {
			output = input[:len(input)-5] + ".t7"
		} else {
			output = input[:len(input)-3] + ".json"
		}
	}
	outputT7 := strings.HasSuffix(output, ".t7")
	var err error = nil
	options := ""
	if l > 3 {
		options = args[3]
	}
	eol := 1
	if strings.Contains(options, "L") {
		eol = 0
	} else if strings.Contains(options, "A") {
		eol = -1
	}
	luaContext := &LuaContext{
		BigEndian: strings.Contains(options, "B"),
		Force:     strings.Contains(options, "F"),
		RemoveGpu: strings.Contains(options, "G"),
		Eol:       eol,
	}
	var luaResult *LuaResult
	if inputT7 {
		luaResult, err = ReadLuaResultFromJson(input, luaContext)
	} else {
		luaResult, err = ReadLuaResultFromT7(input, luaContext)
	}
	if err == nil {
		if outputT7 {
			err = WriteLuaResultToT7(output, luaResult, luaContext)
		} else {
			err = WriteLuaResultToJson(output, luaResult, luaContext)
		}
	}
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Printf(input + " --> " + output)
	}
}
