// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)
package main

import (
        "fmt"
	"github.com/Dobryvechir/microcore/pkg/dvsearch"
        "log" 
	"os"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

func main() {
	args := os.Args
	l := len(args)
	if l < 4 {
		fmt.Println(copyright)
		fmt.Println("Command line: <file mask> <search> <replace> [<start dir or . as default>] [vrt[LRB]csXlX] [log file]")
		fmt.Println("v-verbose, r-readonly, t - trim space: L-at left, R-at right,B-at both (default), c - add '.check' as extension s-skip, l-limit")
		fmt.Println("search and replace support \\n \\r \\ooo where ooo is octal number")
		return
	}
	search := args[2]
	replace := args[3]
	pattern := args[1]
	startDir := "."
	if l > 4 {
		startDir = args[4]
	}
	options := ""
	if l > 5 {
		options = args[5]
	}
	if l > 6 {
		logFile := args[6]
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
	searchOptions := dvsearch.GenerateSearchOptions(search, replace, options, pattern)
	if len(searchOptions.Search) == 0 {
		fmt.Println("Search cannot be empty")
		return
	}
	all, found := dvsearch.SearchProcessDir(startDir, searchOptions)
	fmt.Printf("Files %d found %d\n", all, found)
}
