// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"errors"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var copyrightDvReadDcParams = "Copyright by Volodymyr Dobryvechir 2019"

var helpDvReadDcParams = copyrightDvReadDcParams + "\ndvreaddcparams <microservice name> [<original template>]"

var commonReadDcOk = true

const (
	paramsMapPureFile         = "paramsMapPure.properties"
	paramsMapFile             = "paramsMap.properties"
	paramsSetCmd              = "launchRun.cmd"
	templatePropertiesFile    = "template.properties"
	templateEnvPath           = "objects.find({\"kind\":\"DeploymentConfig\"}).spec.template.spec.containers[0].env"
	serviceUpFile             = "serviceUp.cmd"
	serviceUpFileContent      = "@oc new-app -f dvtemplate.json\n"
	serviceDownFile           = "serviceDown.cmd"
	updateTemplateFile        = "updateTemplate.cmd"
	updateTemplateFileContent = "@ocdvhelper.exe template.json dvtemplate.json\n"
	updateFullFile            = "updateFull.cmd"
	updateFullFileContent     = "@call " + updateTemplateFile + "\n@call " + serviceDownFile + "\n@call " + serviceUpFile + "\n"
)

type ocTemplateParameter struct {
	value               string
	path                string
	openShiftObjectType string
	provided            bool
}

type OcTemplateProcessingContext struct {
	FieldByObjectType map[string]*dvjson.DvFieldInfo
	MicroServiceName  string
	ObjectNameByType  map[string]string
}

func writeFile(fileName string, data string) {
	err := ioutil.WriteFile(fileName, []byte(data), 0644)
	if err != nil {
		fmt.Printf("Error writing %s: %v", fileName, err)
	}
}

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
		commonReadDcOk = false
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

func isRealMicroService(name string) bool {
	if len(name) == 0 {
		return false
	}
	ok, err := dvoc.IsMicroServicePresent(name)
	if err != nil {
		fmt.Printf("Error %v", err)
		return false
	}
	return ok
}

func tryLearnMicroServiceNameFromCurrentFolderPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	sep := "/"
	if strings.Index(dir, "\\") > strings.Index(dir, "/") {
		sep = "\\"
	}
	dirParts := strings.Split(dir, sep)
	n := len(dirParts) - 1
	if dirParts[n] == "" {
		n--
	}
	if n >= 0 && dirParts[n] == "openshift" {
		n--
	}
	if n >= 0 && isRealMicroService(dirParts[n]) {
		return dirParts[n], nil
	}
	n--
	if n >= 0 && isRealMicroService(dirParts[n]) {
		return dirParts[n], nil
	}
	return "", errors.New("failed to determine the microService name from your path")
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
	if l > 1 {
		data, err = ioutil.ReadFile(args[1])
		if err != nil || len(data) == 0 {
			fmt.Printf("Error in %s: %v", args[1], err)
			os.Exit(1)
		}
	}
	if l > 0 {
		microServiceName = args[0]
		isPresent := false
		if strings.Contains(microServiceName, ":") || strings.Contains(microServiceName, "/") || strings.Contains(microServiceName, "\\") {
			if l != 1 {
				fmt.Printf("Bad microservice name: %s", microServiceName)
				os.Exit(1)
			}
		} else {
			isPresent, err = dvoc.IsMicroServicePresent(microServiceName)
		}
		if l == 1 {
			if !isPresent {
				data, err = ioutil.ReadFile(args[1])
				if err != nil || len(data) == 0 {
					fmt.Printf("%s is neither microservice name nor template file name")
					os.Exit(1)
				} else {
					isPresent = true
					err = nil
					microServiceName = ""
				}
			}
		}
		if err != nil {
			fmt.Printf("Error %v", err)
			os.Exit(1)
		}
		if !isPresent {
			fmt.Printf("Microservice %s does not exist", microServiceName)
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
	if commonReadDcOk {
		fmt.Printf("Successfully saved in %s %s", paramsMapFile, paramsSetCmd)
	} else {
		os.Exit(1)
	}
}

func setObjectNameByType(context *OcTemplateProcessingContext, kind string, name string, env map[string]*ocTemplateParameter) {
	if strings.Index(name, "${") >= 0 {
		var rest []string
		name, rest = dvparser.UpdateModelByParamGetter(name, func(key string) (string, bool) {
			value, ok := env[key]
			if ok && (value.provided || (value.value != "" && strings.Index(value.value, "${") < 0)) {
				return value.value, true
			}
			return "", false
		})
		if len(rest) == 1 && rest[0] == "OPENSHIFT_SERVICE_NAME" {
			if objectName, ok := dvoc.ResolveMostSimilarObjectByMicroserviceNameAndObjectType(context.MicroServiceName, kind); ok {
				rest = nil
				name = objectName
			}
		}
		if len(rest) > 0 {
			return
		}
	}
	context.ObjectNameByType[kind] = name
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

func createSimpleMapFromOcTemplateParameters(ocMap map[string]*ocTemplateParameter) map[string]string {
	res := make(map[string]string)
	list := ""
	for k, v := range ocMap {
		val := v.value
		if !v.provided {
			list += " " + k
			if val == "" && strings.HasPrefix(k, "ENV_") && ocMap[k[4:]] != nil && ocMap[k[4:]].provided {
				val = ocMap[k[4:]].value
			}
		}
		res[k] = val
	}
	if list != "" {
		fmt.Printf("Not provided parameters: %s\n", list)
	}
	return res
}

func smartDiscoveryOfOpenShiftParameters(microServiceName string, envMap map[string]string, template *dvjson.DvFieldInfo, dc *dvjson.DvFieldInfo) error {
	fieldByObjectType := make(map[string]*dvjson.DvFieldInfo)
	fieldByObjectType["DeploymentConfig"] = dc
	context := &OcTemplateProcessingContext{FieldByObjectType: fieldByObjectType, MicroServiceName: microServiceName, ObjectNameByType: make(map[string]string)}
	ocMap, err := collectAllTemplateParameters(template, envMap, context)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		commonReadDcOk = false
		return err
	}
	paramMap := createSimpleMapFromOcTemplateParameters(ocMap)
	writeEnvironment(templatePropertiesFile, paramMap, "", nil)
	writeFile(serviceUpFile, serviceUpFileContent)
	writeFile(serviceDownFile, getMicroServiceDownInfo(context, template, paramMap))
	writeFile(updateTemplateFile, updateTemplateFileContent)
	writeFile(updateFullFile, updateFullFileContent)
	return nil
}

func addSingleVariableParam(res map[string]*ocTemplateParameter, k string, v string) {
	if res[k] != nil {
		res[k].value = v
		res[k].provided = true
	} else {
		res[k] = &ocTemplateParameter{value: v, provided: true}
	}
}

func addVariableParams(res map[string]*ocTemplateParameter, params map[string]string) {
	if params != nil {
		for k, v := range params {
			addSingleVariableParam(res, k, v)
		}
	}
}

func setVariablesBy(model string, pattern string, res map[string]*ocTemplateParameter) {
	hintOccurrence, hintStrict := getHintForPattern(pattern)
	if params, doubt, ok := dvparser.SubstitutionMatchModelByPattern(model, pattern, hintOccurrence, hintStrict); ok {
		if doubt {
			fmt.Printf("Dubious match in %s by %s", model, pattern)
		}
		addVariableParams(res, params)
	}
}

func getHintForPattern(pattern string) (int, bool) {
	if strings.HasSuffix(pattern, "${TAG}") {
		return -1, true
	}
	return 0, false
}

func collectAllTemplateParameters(template *dvjson.DvFieldInfo, envMap map[string]string, context *OcTemplateProcessingContext) (res map[string]*ocTemplateParameter, err error) {
	parameters := template.ReadSimpleChild("parameters")
	if parameters == nil || parameters.Kind != dvjson.FIELD_ARRAY {
		err = fmt.Errorf("template has no parameters section")
		return
	}
	fields := parameters.Fields
	n := len(fields)
	res = make(map[string]*ocTemplateParameter)
	for i := 0; i < n; i++ {
		if fields[i] == nil {
			continue
		}
		d := fields[i].ReadSimpleChild("name")
		if d == nil {
			continue
		}
		v := fields[i].ReadSimpleChild("value")
		key := string(d.Value)
		value := ""
		if v != nil {
			value = string(v.Value)
		}
		res[key] = &ocTemplateParameter{value: value}
	}
	envField, err := template.ReadChild(templateEnvPath, nil)
	if err != nil || envField == nil {
		return
	}
	n = len(envField.Fields)
	for i := 0; i < n; i++ {
		k := envField.Fields[i].ReadSimpleChild("name")
		v := envField.Fields[i].ReadSimpleChild("value")
		if k != nil && v != nil {
			key := string(k.Value)
			value := string(v.Value)
			if list, rest, ok := dvparser.SubstitutionGetList(k.Value); ok {
				found := ""
				for lk, liveValue := range envMap {
					params, doubt, ok := dvparser.MatchParametersByVariableSubstitution(lk, list, rest, 0, false)
					if ok {
						if doubt {
							fmt.Printf("Not unique between %s and %s", lk, key)
						}
						if found != "" {
							fmt.Printf("Not unique environment variable: both %s and %s match %s", lk, found, key)
						}
						found = lk
						addVariableParams(res, params)
						setVariablesBy(liveValue, value, res)
					}
				}
			} else {
				if liveValue, ok := envMap[key]; ok {
					setVariablesBy(liveValue, value, res)
				}
			}
		}
	}
	objects := template.ReadSimpleChild("objects")
	if objects == nil {
		err = errors.New("objects is not present in the template")
		return
	}
	objects.Fields, _ = dvoc.OrderTemplateObjectsByDependencies(objects.Fields, true)
	n = len(objects.Fields)
	for i := 0; i < n; i++ {
		item := objects.Fields[i]
		if item == nil {
			continue
		}
		kind := strings.ToLower(item.ReadSimpleChildValue("kind"))
		if kind == "" {
			err = fmt.Errorf("objects at %d does not contain kind", i)
			continue
		}
		name := item.ReadChildStringValue("metadata.name")
		if name != "" {
			setObjectNameByType(context, kind, name, res)
		}
		provideObjectParameters(item, kind, "", res, context)
	}
	return
}

func listAlreadyProvided(list []string, rest []string, res map[string]*ocTemplateParameter) ([]string, []string) {
	n := len(list)
	for i := 0; i < n; i++ {
		k := list[i]
		if res[k] != nil && res[k].provided {
			list, rest, i = dvparser.ReduceListRestByKnownValueAtIndex(list, rest, i, res[k].value)
			n = len(list)
		}
	}
	return list, rest
}

func provideObjectParameters(item *dvjson.DvFieldInfo, objectType string, path string, res map[string]*ocTemplateParameter, context *OcTemplateProcessingContext) {
	if item == nil {
		return
	}
	switch item.Kind {
	case dvjson.FIELD_ARRAY, dvjson.FIELD_OBJECT:
		n := len(item.Fields)
		for i := 0; i < n; i++ {
			if item.Fields[i] == nil {
				continue
			}
			var pathMore string
			if item.Kind == dvjson.FIELD_ARRAY {
				pathMore = "[" + strconv.Itoa(i) + "]"
			} else {
				pathMore = dvjson.GetNextPathPartByKey(string(item.Fields[i].Name))
			}
			provideObjectParameters(item.Fields[i], objectType, path+pathMore, res, context)
		}
		return
	}
	list, rest, need := dvparser.SubstitutionGetList(item.Value)
	if !need {
		return
	}
	list, rest = listAlreadyProvided(list, rest, res)
	if len(list) == 0 {
		return
	}
	model, err := getLiveSimpleValueByObjectType(objectType, path, context)
	if err != nil {
		fmt.Printf("Missing values for %v  (%v)\n", list, err)
		for _, v := range list {
			if _, ok := res[v]; !ok {
				res[v] = &ocTemplateParameter{value: ""}
			}
		}
	} else {
		hintOccurrence, hintStrict := getHintForPattern(string(item.Value))
		if params, doubt, ok := dvparser.MatchParametersByVariableSubstitution(model, list, rest, hintOccurrence, hintStrict); ok {
			if doubt {
				fmt.Printf("Dubious match in %s by %s\n", model, string(item.Value))
			}
			addVariableParams(res, params)
		} else {
			fmt.Printf("Missing values for %v (%s, %v)\n", list, model, rest)
			for _, v := range list {
				if _, ok := res[v]; !ok {
					res[v] = &ocTemplateParameter{value: ""}
				}
			}
		}
	}
}

func getLiveSimpleValueByObjectType(objectType string, path string, context *OcTemplateProcessingContext) (model string, err error) {
	fieldInfo, ok := context.FieldByObjectType[objectType]
	if !ok {
		objectName, ok := context.ObjectNameByType[objectType]
		if ok {
			fieldInfo, err = dvoc.GetConfigurationByOpenShiftObjectTypeAndName(objectName, objectType)
		} else {
			fieldInfo, err = dvoc.GetConfigurationByOpenShiftObjectType(context.MicroServiceName, objectType)
		}
		if err != nil {
			return
		}
		context.FieldByObjectType[objectType] = fieldInfo
	}
	value, err := fieldInfo.ReadChild(path, nil)
	if err != nil {
		return
	}
	model = dvparser.GetUnquotedString(value.GetStringValue())
	return
}

func getMicroServiceDownInfo(context *OcTemplateProcessingContext, template *dvjson.DvFieldInfo, paramMap map[string]string) string {
	list := ""
	objects := template.ReadSimpleChild("objects")
	if objects != nil {
		n := len(objects.Fields)
		for i := 0; i < n; i++ {
			o := objects.Fields[i]
			kind := o.ReadSimpleChildValue("kind")
			name := o.ReadChildStringValue("metadata.name")
			objType, err := dvoc.GetShortOpenShiftNameForObjectType(kind)
			if err != nil {
				fmt.Printf("delete %s problem: %v", kind, err)
			} else {
				if name == "" {
					fmt.Printf("delete problem: cannot detect name for %s", kind)
				} else {
					finalName, err := dvparser.UpdateModelByParams(name, paramMap)
					if err != nil {
						fmt.Printf("delete problem: %v", err)
					} else if finalName == "" {
						fmt.Printf("delete problem: cannot detect the name for %s", kind)
					} else {
						list += "@oc delete " + objType + " " + finalName + "\n"
					}
				}
			}
		}
	}
	return list
}
