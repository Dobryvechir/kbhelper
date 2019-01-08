// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"os"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("kbhelper <podname> <podlist, pods.txt by default> <cmd name, r.cmd by default> <command call by default>")
		fmt.Println("or kbhelper - <filename to convert all cr/lf to lf for linux")
		fmt.Println("or kbhelper + <filename> <line to be added if it is not present yet, everything in Linux style>")
		return
	}
	podName := os.Args[1]
	podList := "pods.txt"
	podCmd := "r.cmd"
	podCaller := "call"
	if l >= 3 {
		podList = os.Args[2]
	}
	if l >= 4 {
		podCmd = os.Args[3]
	}
	if l >= 5 {
		podCaller = os.Args[4]
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
