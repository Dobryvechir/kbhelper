// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("dvenvironment newTextFileName oldTextFileName newTextFileName1 oldTextFileName1 ...")
		return
	}
	n := l / 2
	if l&1 != 0 {
		os.MkdirAll(args[l-1], os.ModePerm)
	}
	for i := 0; i < n; i++ {
		src := args[i<<1]
		dst := args[i<<1|1]
		if src == "" || dst == "" {
			panic("Pair " + strconv.Itoa(i+1) + " is not defined (" + src + "," + dst + ")")
		}
		data, err := dvparser.ConvertFileByGlobalProperties(src)
		if err != nil {
			panic("Error: " + err.Error())
		}
		dir := filepath.Dir(dst)
		if dir != "" && dir != "." {
			os.MkdirAll(dir, os.ModePerm)
		}
		err = ioutil.WriteFile(dst, data, 0466)
		if err != nil {
			panic("Error: " + err.Error())
		}
	}
}
