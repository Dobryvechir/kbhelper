// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
)

var copyrightDvSecret = "Copyright by Volodymyr Dobryvechir 2019"

var help = copyrightDvSecret + "\ndvreaddcparams <microservice name> [original template]"

var commonOk = true

func writeEnvironment(fileName string, params map[string]string, prefix string) {
	n := len(params)
	prefixBytes := []byte(prefix)
	prefixLen := len(prefixBytes)
	count := (prefixLen + 3) * n
	for k, v := range params {
		count += len([]byte(k)) + len([]byte(v))
	}
	buf := make([]byte, count)
	pos := 0
	for k, v := range params {
		for i := 0; i < prefixLen; i++ {
			buf[pos] = prefixBytes[i]
			pos++
		}
		for _, z := range []byte(k) {
			buf[pos] = z
			pos++
		}
		buf[pos] = '='
		pos++
		for _, z := range []byte(v) {
			buf[pos] = z
			pos++
		}
		buf[pos] = byte(13)
		pos++
		buf[pos] = byte(10)
		pos++
	}
	err := ioutil.WriteFile(fileName, buf, 0644)
	if err != nil {
		fmt.Printf("Error writing to %s: %v", fileName, err)
		commonOk = false
	}
}

func mainTest() bool {
	data, err := ioutil.ReadFile("C:/temp/t.yml")
	if err != nil {
		fmt.Printf("Failed to read %v", err)
		return false
	}
	info, err := dvjson.ReadYamlAsDvFieldInfo(data)
	if err != nil {
		fmt.Printf("Failed to convert %v", err)
		return true
	}
	fmt.Printf("Successful: %s", info.GetStringValue())
	return true
}

func main() {
	if mainTest() {
		return
	}
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 || (args[0] == "--help" || args[0] == "version" || args[0] == "-version" || args[0] == "--version") {
		fmt.Println(help)
		os.Exit(1)
		return
	}
	microServiceName := args[0]
	paramsMapFile := "paramsMap.properties"
	paramsSetCmd := "launchRun.cmd"
	openshiftParams := "template.properties"
	templateOrig := ""
	if l >= 2 {
		templateOrig = args[1]
	} else {
		openshiftParams = ""
	}
	envMap, err := dvoc.ReadPodReadyEnvironment(microServiceName)
	if err != nil {
		fmt.Printf("Cannot read pod environment: %v", err)
		os.Exit(1)
	}
	writeEnvironment(paramsMapFile, envMap, "")
	writeEnvironment(paramsSetCmd, envMap, "SET ")
	if templateOrig != "" {
		templateData, err := ioutil.ReadFile(templateOrig)
		if err != nil {
			fmt.Printf("Cannot read original template %s: %v", templateOrig, err)
			os.Exit(1)
		}
		ocMap, obj, err := dvoc.ReadTemplateParameters(templateData)
		if err != nil {
			fmt.Printf("Failed to parse original template %s: %v", templateOrig, err)
			os.Exit(1)
		}
		err = smartDiscoveryOfOpenShiftParameters(ocMap, microServiceName, envMap, templateData, obj)
		if err != nil {
			fmt.Printf("Failed to discover template parameters: %v", err)
			os.Exit(1)
		}
		writeEnvironment(openshiftParams, ocMap, "")
	}
	if commonOk {
		fmt.Printf("Successfully saved in %s %s %s", paramsMapFile, paramsSetCmd, openshiftParams)
	} else {
		os.Exit(1)
	}
}

func smartDiscoveryOfOpenShiftParameters(ocMap map[string]string, microServiceName string, envMap map[string]string, template []byte, obj *dvjson.DvFieldInfo) error {
	return nil
}
