// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"strings"
)

var copyrightDvReadDcParams = "Copyright by Volodymyr Dobryvechir 2019"

var helpDvReadDcParams = copyrightDvReadDcParams + "\ndvreaddcparams <microservice name> [<original template>]"

var commonOk = true

const (
	paramsMapPureFile = "paramsMapPure.properties"
	paramsMapFile     = "paramsMap.properties"
	paramsSetCmd      = "launchRun.cmd"
)

func writeEnvironment(fileName string, params map[string]string, prefix string, extra []string) {
	n := len(params)
	extraLen := len(extra)
	prefixBytes := []byte(prefix)
	prefixLen := len(prefixBytes)
	count := (prefixLen+3)*n + 2*extraLen
	for k, v := range params {
		count += len([]byte(k)) + len([]byte(v))
	}
	if extraLen > 0 {
		for _, k := range extra {
			if prefixLen > 0 {
				if len(k) > 0 && k[0] != '#' && k[0] > ' ' {
					count += prefixLen
				} else {
					count += 4
				}
			}
			count += len([]byte(k))
		}
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
	if extraLen > 0 {
		for _, k := range extra {
			if prefixLen > 0 {
				if len(k) > 0 && k[0] != '#' && k[0] > ' ' {
					for i := 0; i < prefixLen; i++ {
						buf[pos] = prefixBytes[i]
						pos++
					}
				} else {
					buf[pos] = 'R'
					pos++
					buf[pos] = 'E'
					pos++
					buf[pos] = 'M'
					pos++
					buf[pos] = ' '
					pos++
				}
			}
			d := []byte(k)
			dLen := len(d)
			for i := 0; i < dLen; i++ {
				buf[pos] = d[i]
				pos++
			}
			buf[pos] = 13
			pos++
			buf[pos] = 10
			pos++
		}
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

func tryLearnMicroServiceNameFromCurrentFolderPath() (string, error) {
	//TODO: learn it from path
	return "", nil
}

func main() {
	//if mainTest() {
	//	return
	//}
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	microServiceName := ""
	data, err := ioutil.ReadFile("template.json")
	if err != nil {
		data = nil
	}
	if data == nil && l == 0 || l > 0 && (args[0] == "--help" || args[0] == "version" || args[0] == "-version" || args[0] == "--version") {
		fmt.Println(helpDvReadDcParams)
		os.Exit(1)
	}
	if l > 0 {
		microServiceName = args[0]
	}
	if l > 1 {
		data, err = ioutil.ReadFile(args[1])
		if err != nil || len(data) == 0 {
			fmt.Printf("Error in %s: %v", args[1], err)
			os.Exit(1)
		}
	}
	if len(data) == 0 {
		presentMicroServiceInfo(microServiceName)
		return
	}
	templateOrig, err := dvjson.ReadJsonAsDvFieldInfo(data)
	if err != nil || templateOrig == nil {
		fmt.Printf("Error in template: %v", err)
		os.Exit(1)
	}
	if microServiceName == "" {
		microServiceName, err = tryLearnMicroServiceNameFromCurrentFolderPath()
		if err != nil || microServiceName == "" {
			fmt.Println("Failure. Please specify the microservice name")
			fmt.Println(helpDvReadDcParams)
			os.Exit(1)
		}
	}
	envMap, dc := presentMicroServiceInfo(microServiceName)
	smartDiscoveryOfOpenShiftParameters(microServiceName, envMap, templateOrig, dc)
	if commonOk {
		fmt.Printf("Successfully saved in %s %s", paramsMapFile, paramsSetCmd)
	} else {
		os.Exit(1)
	}
}

func readExtraPodFile() []string {
	fileName := dvparser.GlobalProperties["EXTRA_POD_PROPERTIES_FILE"]
	if fileName == "" {
		return nil
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading file %s: %v", fileName, err)
		return nil
	}
	res, err := dvparser.ConvertByteArrayByGlobalPropertiesInStringLines(data, fileName)
	if err != nil {
		fmt.Printf("Error processing file %s: %v", fileName, err)
		return nil
	}
	return res
}

func presentMicroServiceInfo(microServiceName string) (map[string]string, *dvjson.DvFieldInfo) {
	if microServiceName == "" {
		fmt.Printf("Error in command line! Please, specify microservice name ")
		os.Exit(1)
	}
	dvparser.SetGlobalPropertiesValue("MICROSERVICE_NAME", microServiceName)
	envMap, dc, err := dvoc.ReadPodReadyEnvironment(microServiceName)
	if err != nil {
		fmt.Printf("Cannot read pod environment: %v", err)
		os.Exit(1)
	}
	extra := readExtraPodFile()
	writeEnvironment(paramsMapPureFile, envMap, "", nil)
	writeEnvironment(paramsMapFile, envMap, "", extra)
	writeEnvironment(paramsSetCmd, envMap, "SET ", extra)
	return envMap, dc
}

func smartDiscoveryOfOpenShiftParameters(microServiceName string, envMap map[string]string, template *dvjson.DvFieldInfo, dc *dvjson.DvFieldInfo) error {
	return nil
}
