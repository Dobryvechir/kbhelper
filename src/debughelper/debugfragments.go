// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"github.com/Dobryvechir/dvserver/src/dvlog"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"github.com/Dobryvechir/dvserver/src/dvdir"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	programName            = "Debug Fragments 1.0" + author
	contentSecurityPolicy  = "CONTENT_SECURITY_POLICY"
	SuccessExitCode        = 0
	ErrorExitCode          = 1
	commandLineExpectation = "Command line: DebugFragment [start | finish | up | down | reset | execute [name]] | json [file]"
)

var logDebugFragments = 0
var logDebug = false

func startDebugFragment() int {
	microServiceName, ok := getCurrentServiceName()
	if !ok {
		dvlog.PrintfError("Microservice is not specified in properties")
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
	ok = startUiConfiguration()
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
		dvlog.PrintlnError("Successfully started fragment debug")
	}
	return SuccessExitCode
}

func startDebugFragmentShort() int {
	fragmentListConfig, ok := readFragmentListConfigurationFromFile()
	if !ok {
		return ErrorExitCode
	}
	_, specials, ok := createDebugFragmentListConfig(fragmentListConfig)
	if !ok {
		return ErrorExitCode
	}
	if runDvServer(specials) {
		dvlog.PrintlnError("Successfully started server shortly")
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
		dvlog.PrintlnError("Fragments are already in production state")
		return SuccessExitCode
	}
	muiContent, ok := retrieveProductionFragmentListConfiguration()
	if !ok {
		return ErrorExitCode
	}
	//deregisterFragment()
	ok = registerFragment(muiContent)
	if !ok {
		return ErrorExitCode
	}
	ok = finishUiConfiguration()
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
	dvlog.PrintlnError("Successfully finished fragment debug")
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
		dvlog.PrintfError("For up command, you must specify all of these parameters in dvserver.properties: DISTRIBUTION_FOLDER  TEMPLATE_IMAGE POD_HTML_FOLDER")
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
		if !dvoc.CreateMicroService(params, files, nil) {
			dvlog.PrintfError("Failed to create microservice for %s (%s)", serviceName, templateImage)
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

func restoreCurrentMicroService() int {
	serviceName, ok := getCurrentServiceName()
	if !ok {
		return ErrorExitCode
	}
	ok = dvoc.ExecuteSingleCommand(0, 0, dvoc.CommandMicroServiceRestore, serviceName)
	if !ok {
		return ErrorExitCode
	}
	return SuccessExitCode
}

func main() {
	dvlog.PrintlnError(programName)
	args := dvparser.InitAndReadCommandLine()
	dvparser.SetNumberOfBracketsInConfigParsing(2)
	l := len(args)
	if l < 1 {
		dvlog.PrintlnError(commandLineExpectation)
		return
	}
	debugLevel := dvparser.GlobalProperties["DEBUG_LEVEL"]
	if debugLevel != "" {
		n, err := strconv.Atoi(debugLevel)
		if err != nil {
			dvlog.PrintlnError("DEBUG_LEVEL must be integer")
		} else {
			logDebugFragments = n
		}
		logDebug = logDebugFragments > 0
	}
	if logDebugFragments&2 != 0 {
		dvnet.Log = dvnet.LogInfo
		dvoc.Log = dvoc.LogInfo
		if logDebugFragments&4 != 0 {
			dvnet.Log = dvnet.LogDetail
			dvoc.Log = dvoc.LogDetail
		}
		if logDebugFragments&8 != 0 {
			dvnet.Log = dvnet.LogDebug
			dvoc.Log = dvoc.LogDebug
		}
	}
	exitCode := ErrorExitCode
	noCache := dvparser.GlobalProperties["NO_CACHE"]
	if noCache != "" && noCache != "false" {
		dvdir.ResetAllLocalFileCache()
	}
	switch args[0] {
	case "start":
		exitCode = startDebugFragment()
	case "server":
		exitCode = startDebugFragmentShort()
	case "finish":
		exitCode = finishDebugFragment()
	case "up":
		exitCode = raiseUpInCloud()
	case "down":
		exitCode = removeFromCloud()
	case "restore":
		exitCode = restoreCurrentMicroService()
	case "reset":
		exitCode = resetPod()
	case "execute":
		if l < 2 {
			dvlog.PrintlnError("Execute requires an additional parameter - name and EXECUTE_NAME_1, EXECUTE_NAME_2 ... properties in dvserver..properties")
		} else {
			prefix := "EXECUTE_" + strings.ToUpper(args[1])
			ok := dvoc.ExecuteSequence(prefix)
			if ok {
				exitCode = SuccessExitCode
			}
		}
	case "json":
		if l < 2 {
			dvlog.PrintlnError("json requires an additional parameter - file name (optional property JSON_TRASH, JSON_INDENT)")
		} else {
			trashStr := dvparser.ConvertToNonEmptyList(dvparser.GlobalProperties["JSON_TRASH"])
			trash := dvjson.ConvertStringArrayToByteByteArray(trashStr)
			indent := 4
			s := dvparser.GlobalProperties["JSON_INDENT"]
			if s != "" {
				n, err := strconv.Atoi(s)
				if err != nil {
					dvlog.PrintlnError("Error in JSON_INDENT - not an integer number")
				} else {
					indent = n
				}
			}
			fileName := args[1]
			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				dvlog.PrintfError("Error reading %s: %v", fileName, err)
			} else {
				data, err = dvjson.ReformatJson(data, indent, trash, 2)
				if err != nil {
					dvlog.PrintfError("Error in json of %s: %v", fileName, err)
				} else {
					err = ioutil.WriteFile(fileName, data, 0664)
					if err != nil {
						dvlog.PrintfError("Error writing %s: %v", fileName, err)
					}
				}
			}
		}
	default:
		dvlog.PrintlnError(programName)
		dvlog.PrintlnError(commandLineExpectation)
	}
	dvlog.FlushStream()
	if exitCode > 0 {
		os.Exit(exitCode)
	}
}
