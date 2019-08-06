// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	programName     = "Debug Fragments 1.0" + author
	SuccessExitCode = 0
	ErrorExitCode   = 1
)

var logDebugFragments = 0
var logDebug = false

func startDebugFragment() int {
	token, ok := getM2MToken("mui-platform")
	if !ok {
		return ErrorExitCode
	}
	fragmentListConfig, ok := readFragmentListConfigurationFromFile()
	if !ok {
		return ErrorExitCode
	}
	ok = saveCloudConfigForThisFragment(fragmentListConfig, token)
	if !ok {
		return ErrorExitCode
	}
	newConfig, specials, ok := createDebugFragmentListConfig(fragmentListConfig)
	if !ok {
		return ErrorExitCode
	}
	muiContent, ok := convertListConfigToJson(newConfig)
	if !ok {
		return ErrorExitCode
	}
	//deregisterFragment()
	saveCloudDebugFragmentInfo(newConfig)
	ok = registerFragment(muiContent)
	if !ok {
		return ErrorExitCode
	}
	ok = resetMuiCache()
	if !ok {
		return ErrorExitCode
	}
	ok = atStartExecutions()
	if !ok {
		return ErrorExitCode
	}
	if runDvServer(specials) {
		log.Println("Successfully started fragment debug")
	}
	return SuccessExitCode
}

func finishDebugFragment() int {
	fragmentListConfig, ok := readFragmentListConfigurationFromFile()
	if !ok {
		return ErrorExitCode
	}
	ok = checkCloudConfigIsOriginal(fragmentListConfig)
	if ok {
		log.Println("Fragments are already in production state")
		return SuccessExitCode
	}
	muiContent, ok := retrieveProductionFragmentListConfiguration()
	if !ok {
		return ErrorExitCode
	}
	deregisterFragment()
	ok = registerFragment(muiContent)
	if !ok {
		return ErrorExitCode
	}
	ok = resetMuiCache()
	if !ok {
		return ErrorExitCode
	}
	ok = atFinishExecutions()
	if !ok {
		return ErrorExitCode
	}
	log.Println("Successfully finished fragment debug")
	return SuccessExitCode
}

func resetPod() int {
	if !deleteCurrentPod() {
		return ErrorExitCode
	}
	return SuccessExitCode
}

func raiseUpInCloud() int {
	distributionFolder := dvparser.GlobalProperties["DISTRIBUTION_FOLDER"]
	templateImage := dvparser.GlobalProperties["TEMPLATE_IMAGE"]
	htmlFolder := dvparser.GlobalProperties["POD_HTML_FOLDER"]
	if distributionFolder == "" || templateImage == "" || htmlFolder == "" {
		log.Printf("For up command, you must specify all of these parameters in dvserver.properties: DISTRIBUTION_FOLDER  TEMPLATE_IMAGE POD_HTML_FOLDER")
		return ErrorExitCode
	}
	podName, _ := getCurrentPodName(true)
	if podName == "" {
		downCurrentMicroservice()
		serviceName, ok := getCurrentServiceName()
		if !ok {
			return ErrorExitCode
		}
		if !createMicroservice(serviceName, templateImage) {
			log.Printf("Failed to create microservice for %s (%s)", serviceName, templateImage)
			return ErrorExitCode
		}
		for i := 0; i < 100; i++ {
			time.Sleep(2 * time.Second)
			podName, _ = getCurrentPodName(true)
			if podName != "" {
				break
			}
		}
		if podName == "" {
			log.Printf("Waiting for pod %s getting up is timed out", serviceName)
			return ErrorExitCode
		}
		if logDebug {
			log.Printf("Waiting for 10 seconds until the pod %s is ready (distribution folder=%s, html folder=%s)", podName, distributionFolder, htmlFolder)
		}
		time.Sleep(10 * time.Second)
	}
	if !synchronizeDirectory(podName, distributionFolder, htmlFolder) {
		return ErrorExitCode
	}
	return SuccessExitCode
}

func removeFromCloud() int {
	if !downCurrentMicroservice() {
		return ErrorExitCode
	}
	return SuccessExitCode
}

func main() {
	log.Println(programName)
	args := dvparser.InitAndReadCommandLine()
	dvparser.SetNumberOfBracketsInConfigParsing(2)
	l := len(args)
	if l < 1 {
		log.Println("Command line: DebugFragment [start | finish | up | down | reset]")
		return
	}
	debugLevel := dvparser.GlobalProperties["DEBUG_LEVEL"]
	if debugLevel != "" {
		n, err := strconv.Atoi(debugLevel)
		if err != nil {
			log.Println("DEBUG_LEVEL must be integer")
		} else {
			logDebugFragments = n
		}
		logDebug = logDebugFragments > 0
	}
	if logDebugFragments&2 != 0 {
		dvnet.DvNetLog = true
	}
	exitCode := ErrorExitCode
	switch args[0] {
	case "start":
		exitCode = startDebugFragment()
	case "finish":
		exitCode = finishDebugFragment()
	case "up":
		exitCode = raiseUpInCloud()
	case "down":
		exitCode = removeFromCloud()
	case "reset":
		exitCode = resetPod()
	default:
		log.Println(programName)
		log.Println("Command line: DebugFragment [start | finish | up | down | reset]")
	}
	if exitCode > 0 {
		os.Exit(exitCode)
	}
}
