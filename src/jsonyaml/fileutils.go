// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"os"
	"strconv"
	"strings"
)

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
