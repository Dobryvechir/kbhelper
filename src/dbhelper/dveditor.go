// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"io/ioutil"
	"os"
)

func GetTempPath() string {
	tempPath := os.Getenv("TEMP")
	if tempPath != "" {
		if _, err := os.Stat(tempPath); err == nil {
			return tempPath
		}
	}
	tempPath = os.Getenv("TMP")
	if tempPath != "" {
		if _, err := os.Stat(tempPath); err == nil {
			return tempPath
		}
	}
	tempPath = "/tmp"
	if _, err := os.Stat(tempPath); err == nil {
		return tempPath
	}
	tempPath = "/temp"
	if _, err := os.Stat(tempPath); err == nil {
		return tempPath
	}
	return ""
}

func main() {
	infoFile := os.Getenv("DVEDITOR_INFO")
	if infoFile == "" {
		infoFile = GetTempPath() + "/dveditor.info.txt"
	}
	contentFile := os.Getenv("DVEDITOR_CONTENT")
	if contentFile == "" {
		contentFile = GetTempPath() + "/dveditor.content.txt"
	}
	args := os.Args
	n := len(args)
	info1 := "Ok"
	if n > 1 {
		name := os.Args[1]
		data, err := ioutil.ReadFile(name)
		if err != nil {
			info1 = "Error: " + err.Error()
		} else {
			err = ioutil.WriteFile(contentFile, data, 0664)
			if err != nil {
				info1 = "Error: " + err.Error()
			}
		}
	} else {
		info1 = "Error: no file specified"
	}
	info2 := os.Args[0]
	for i := 1; i < n; i++ {
		info2 += " " + os.Args[i]
	}
	info := info1 + "\n" + info2 + "\n"
	ioutil.WriteFile(infoFile, []byte(info), 0664)
}
