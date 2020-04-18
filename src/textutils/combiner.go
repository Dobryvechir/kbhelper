// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var copyright = "Copyright by Danyil Dobryvechir 2012-2020"

//TODO now we remove completely -es5 javascripts, probably we should correct them additionally

var prioritiesJs = map[string]int{
	"runtime-es2015.js":   1,
	"runtime-es5.js":      -2,
	"runtime.js":          3,
	"polyfills-es5.js":    -4,
	"polyfills-es2015.js": 5,
	"polyfills.js":        6,
	"styles-es2015.js":    7,
	"styles.js":           8,
	"styles-es5.js":       -9,
	"vendor-es2015.js":    10,
	"vendor.js":           11,
	"vendor-es5.js":       -12,
	"main-es2015.js":      13,
	"main.js":             14,
	"main-es5.js":         -15,
}

var prioritiesCss = map[string]int{}

func findLastFolderName(path string) string {
	n := len(path) - 1
	if n < 0 {
		return ""
	}
	if path[n] == '/' || path[n] == '\\' {
		path = path[:n]
	}
	n = strings.LastIndex(path, "\\")
	na := strings.LastIndex(path, "/")
	if na > n {
		n = na
	}
	return path[n+1:]
}

func collectFileList(path string, relPath string, ext string, priorities map[string]int) ([]string, error) {
	list := make([]string, 0, 7)
	finalPath := path
	if relPath != "" {
		finalPath = path + "/" + relPath
	}
	files, err := ioutil.ReadDir(finalPath)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		name := f.Name()
		if name != "." && name != ".." {
			if f.IsDir() {
				addList, err := collectFileList(path, relPath+"/"+name, ext, priorities)
				if err != nil {
					return nil, err
				}
				list = append(list, addList...)
			} else if f.Mode().IsRegular() && strings.HasSuffix(name, ext) && priorities[name] >= 0 {
				if relPath != "" {
					name = relPath + "/" + name
				}
				list = append(list, name)
			}
		}
	}
	return list, nil
}

func orderFileList(list []string, priorities map[string]int) []string {
	sort.Slice(list, func(i, j int) bool {
		k := priorities[list[j]] - priorities[list[i]]
		res := k > 0
		if k == 0 {
			res = list[i] > list[j]
		}
		return res
	})
	return list
}

func combiningListFiles(path string, list []string, name string) (int, error) {
	file, err := os.Create(name)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	n := len(list)
	eol := []byte{13, 10}
	for i := 0; i < n; i++ {
		data, err := ioutil.ReadFile(path + "/" + list[i])
		if err != nil {
			return 0, err
		}
		if i != 0 {
			_, err = file.Write(eol)
			if err != nil {
				return 0, err
			}
		}
		_, err = file.Write(data)
		if err != nil {
			return 0, err
		}
		fmt.Printf("+ %s\n", list[i])
	}
	err = file.Sync()
	if err != nil {
		return 0, err
	}
	return n, nil
}

func processCombiningFiles(path string, name string, ext string, priorities map[string]int) int {
	name += ext
	n := len(path) - 1
	if path[n] == '\\' || path[n] == '/' {
		path = path[:n]
	}
	if strings.Index(name, "\\") < 0 && strings.Index(name, "/") < 0 {
		name = path + "/" + name
	}
	os.Remove(name)
	list, err := collectFileList(path, "", ext, priorities)
	if err != nil {
		fmt.Printf("Error collecting files *%s:%v", ext, err)
		return 0
	}
	list = orderFileList(list, priorities)
	if len(list) == 0 {
		return 0
	}
	nmb, err := combiningListFiles(path, list, name)
	if err != nil {
		fmt.Printf("Error combining files to %s:%v", name, err)
		return 0
	}
	return nmb
}

func main() {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	l := len(os.Args)
	if l > 1 && (os.Args[1] == "-version" || os.Args[1] == "-help" || os.Args[1] == "--version" || os.Args[1] == "--help") {
		fmt.Printf(copyright)
		fmt.Printf("\nCommand line: combiner <optional filename> <options> <optional dirfile> ")
		return
	}
	if l > 3 {
		path, err = filepath.Abs(os.Args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	options := "cj"
	if l > 2 {
		options = os.Args[2]
	}
	lastFolder := findLastFolderName(path)
	if lastFolder == "" {
		fmt.Printf("Please do not specify the disk root folder as the path: %s\n", path)
		return
	}
	if l > 1 {
		lastFolder = os.Args[1]
	}
	res := 0
	if strings.Index(options, "j") >= 0 {
		res += processCombiningFiles(path, lastFolder, ".js", prioritiesJs)
	}
	if strings.Index(options, "c") >= 0 {
		res += processCombiningFiles(path, lastFolder, ".css", prioritiesCss)
	}
	if res == 0 {
		fmt.Println("No files to process at all")
	} else {
		fmt.Printf("%d files were combined to %s\n", res, lastFolder)
	}
}
