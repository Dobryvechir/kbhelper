// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"log"
	"strconv"
	"time"
)

const (
	programName = "Debug Fragments 1.0" + author
)

var logDebugFragments = 0
var logDebug = false

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
	newConfig, specials, ok := createDebugFragmentListConfig(fragmentListConfig)
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
	if runDvServer(specials) {
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

func resetPod() {
	deleteCurrentPod()
}

func raiseUpInCloud() {
	distributionFolder := dvparser.GlobalProperties["DISTRIBUTION_FOLDER"]
	templateImage := dvparser.GlobalProperties["TEMPLATE_IMAGE"]
	htmlFolder := dvparser.GlobalProperties["POD_HTML_FOLDER"]
	if distributionFolder == "" || templateImage == "" || htmlFolder == "" {
		log.Printf("For up command, you must specify all of these parameters in dvserver.properties: DISTRIBUTION_FOLDER  TEMPLATE_IMAGE POD_HTML_FOLDER")
		return
	}
	podName, _ := getCurrentPodName(true)
	if podName == "" {
		downCurrentMicroservice()
		serviceName, ok := getCurrentServiceName()
		if !ok {
			return
		}
		if !createMicroservice(serviceName, templateImage) {
			log.Printf("Failed to create microservice for %s (%s)", serviceName, templateImage)
			return
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
			return
		}
		if logDebug {
			log.Printf("Waiting for 10 seconds until the pod %s is ready (distribution folder=%s, html folder=%s)", podName, distributionFolder, htmlFolder)
		}
		time.Sleep(10 * time.Second)
	}
	synchronizeDirectory(podName, distributionFolder, htmlFolder)
}

func removeFromCloud() {
	downCurrentMicroservice()
}

func main() {
	log.Println(programName)
	args := dvparser.InitAndReadCommandLine()
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
	if logDebugFragments & 4!=0 {
		dvnet.DvNetLog = true
	}
	switch args[0] {
	case "start":
		startDebugFragment()
	case "finish":
		finishDebugFragment()
	case "up":
		raiseUpInCloud()
	case "down":
		removeFromCloud()
	case "reset":
		resetPod()
	default:
		log.Println(programName)
		log.Println("Command line: DebugFragment [start | finish | up | down | reset]")
	}
}
