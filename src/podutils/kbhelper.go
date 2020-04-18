// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("kbhelper <podname> <podlist, pods.txt by default> <cmd name, r.cmd by default> <command call by default>")
		fmt.Println("or kbhelper - <filename to convert all cr/lf to lf for linux")
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
	case "-":
		if l < 3 {
			fmt.Println("File name is not specified")
		} else {
			removeCRLF(podList)
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
