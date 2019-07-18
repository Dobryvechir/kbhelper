// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"log"
)

const (
	programName = "Debug Fragments 1.0" + author
)

func startDebugFragment() {
	_, ok := getM2MToken("mui-fragments")
	if !ok {
		return
	}
	fragmentListConfig, ok := readCurrentFragmentListConfigurationFromCloud()
	if !ok {
		return
	}
	if fragmentListConfig == nil {
		fragmentListConfig, ok = readFragmentListConfigurationFromFile()
		if !ok {
			return
		}
	}
	newConfig, ok := createDebugFragmentListConfig(fragmentListConfig)
	if !ok {
		return
	}
	muiContent, ok := convertListConfigToJson(newConfig)
	if !ok {
		return
	}
	deregisterFragment()
	ok = registerFragment(muiContent)
	if !ok {
		return
	}
	if runDvServer() {
		log.Println("Successfully started fragment debug")
	}
}

func finishDebugFragment() {
	muiContent, ok := retrieveProductionFragmentListConfiguration()
	if !ok {
		return
	}
	deregisterFragment()
	ok = registerFragment(muiContent)
	if !ok {
		return
	}
	log.Println("Successfully finished fragment debug")
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(programName)
		fmt.Println("Command line: DebugFragment start | DebugFragment finish")
		return
	}
	switch args[0] {
	case "start":
		startDebugFragment()
	case "finish":
		finishDebugFragment()
	default:
		fmt.Println(programName)
		fmt.Println("Command line: DebugFragment start | DebugFragment finish")
	}
}
