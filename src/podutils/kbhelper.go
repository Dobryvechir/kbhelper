// Copyright by Danyil Dobryvechir 2020 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/microcore/pkg/dvparser"
)

var copyright = "Copyright by Danyil Dobryvechir 2020"

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("kbhelper <podname> <podlist, pods.txt by default> <cmd name, r.cmd by default> <command call by default>")
		fmt.Println("or kbhelper l <file/dir to convert all cr/lf or cr to lf for linux>")
		fmt.Println("or kbhelper w <file/dir to convert all cr or lf to cr/lf for windows>")
		fmt.Println("or kbhelper + <filename> <line to be added if it is not present yet, everything in Linux style>")
		return
	}
	podName := args[0]
	podList := "pods.txt"
	podCmd := "r.cmd"
	podCaller := "call"
	if l >= 2 {
		podList = args[1]
	}
	if l >= 3 {
		podCmd = args[2]
	}
	if l >= 4 {
		podCaller = args[3]
	}
	switch podName {
	case "l", "L":
		if l < 2 {
			fmt.Println("File/dir name is not specified")
		} else {
			walkRemoveCrLf(podList)
		}
	case "w", "W":
		if l < 2 {
			fmt.Println("File/dir name is not specified")
		} else {
			walkAddCrLf(podList)
		}
	case "+":
		if l < 4 {
			fmt.Println("File name/line is not specified")
		} else {
			addNonRepeatedLine(podList, podCmd)
		}
	default:
		pod, project := findPod(podList, podName)
		fmt.Println("@" + podCaller + " " + podCmd + " " + pod + " " + project)
	}
}
