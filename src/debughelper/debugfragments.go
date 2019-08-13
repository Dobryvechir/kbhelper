// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	programName           = "Debug Fragments 1.0" + author
	contentSecurityPolicy = "CONTENT_SECURITY_POLICY"
	SuccessExitCode       = 0
	ErrorExitCode         = 1
)

var logDebugFragments = 0
var logDebug = false

func startDebugFragment() int {
	microServiceName, ok := getCurrentServiceName()
	if !ok {
		log.Printf("Microservice is not specified in properties")
		return ErrorExitCode
	}
	token, ok := dvoc.GetM2MToken("mui-platform")
	if !ok {
		return ErrorExitCode
	}
	dvoc.ReduceMicroServiceSaveInfo(microServiceName)
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
	ok = dvoc.UpdateContentSecurityPolicyOnPods(dvparser.GlobalProperties[contentSecurityPolicy], dvparser.GlobalProperties[hostNameParam])
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
		downCurrentMicroService()
		serviceName, ok := getCurrentServiceName()
		if !ok {
			return ErrorExitCode
		}
		params := map[string]string{
			"MICROSERVICE":          serviceName,
			"MANAGING_MICROSERVICE": "mui-platform",
		}
		files := make(map[string]string)
		files[htmlFolder] = distributionFolder
		if !dvoc.CreateMicroService(params, files) {
			log.Printf("Failed to create microservice for %s (%s)", serviceName, templateImage)
			return ErrorExitCode
		}
	}
	return SuccessExitCode
}

func removeFromCloud() int {
	if !downCurrentMicroService() {
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
		log.Println("Command line: DebugFragment [start | finish | up | down | reset | execute [name]]")
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
		dvoc.LogDebug = 0
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
	case "execute":
		if l < 2 {
			log.Println("Execute requires an additional parameter - name and EXECUTE_NAME_1, EXECUTE_NAME_2 ... properties in dvserver..properties")
		} else {
			prefix := "EXECUTE_" + strings.ToUpper(args[1])
			ok := dvoc.ExecuteSequence(prefix)
			if ok {
				exitCode = SuccessExitCode
			}
		}
	default:
		log.Println(programName)
		log.Println("Command line: DebugFragment [start | finish | up | down | reset | execute [name]]")
	}
	if exitCode > 0 {
		os.Exit(exitCode)
	}
}
