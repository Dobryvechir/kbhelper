// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"bytes"
	"encoding/json"
	"github.com/Dobryvechir/dvserver/src/dvlog"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"github.com/Dobryvechir/dvserver/src/dvdir"
	"io/ioutil"
	"os"
	"strings"
)

type UiFragment struct {
	FragmentName              string   `json:"fragmentName"`
	JsResources               []string `json:"jsResources"`
	CssResources              []string `json:"cssResources"`
	Labels                    []string `json:"labels"`
	ImageUrl                  string   `json:"imageUrl"`
	DescriptionLocalizationId string   `json:"descriptionLocalizationId"`
	SkipLocalization          bool     `json:"skipLocalization"`
}

var fragmentPartsToBeRemoved = [][]byte{
	[]byte(",\"labels\":\"\""),
	[]byte(",\"imageUrl\":\"\""),
	[]byte(",\"labels\":null"),
	[]byte(",\"imageUrl\":null"),
	[]byte(",\"descriptionLocalizationId\":\"\""),
	[]byte(",\"descriptionLocalizationId\":null"),
	[]byte(",\"skipLocalization\":false"),
	[]byte(",\"skipLocalization\":null"),
}

type FragmentListConfig struct {
	MicroServiceName string       `json:"microserviceName"`
	Fragments        []UiFragment `json:"fragments"`
}

type FragmentItemConfig struct {
	Id               string     `json:"id"`
	Version          int        `json:"version"`
	TransactionId    string     `json:"transactionId"`
	InternalName     string     `json:"internalName"`
	MicroServiceName string     `json:"microserviceName"`
	Fragment         UiFragment `json:"fragment"`
}

const (
	fragmentListConfiguration = "FRAGMENT_LIST_CONFIGURATION"
	fragmentMicroServiceName  = "FRAGMENT_MICROSERVICE_NAME"
	fragmentServiceName       = "FRAGMENT_SERVICE_NAME"
	MuiPlatformUrl            = "MUI_URL"
	MuiListUrl                = "MUI_LIST_URL"
)

func readFragmentListConfigurationFromFile() (conf *FragmentListConfig, ok bool) {
	fileName := dvparser.GlobalProperties[fragmentListConfiguration]
	if fileName == "" {
		dvlog.PrintlnError("specify FRAGMENT_LIST_CONFIGURATION as a path of file where fragment list is configured")
		return
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		dvlog.PrintfError("Your FRAGMENT_LIST_CONFIGURATION does not point to a valid file name: %s", fileName)
		return
	}
	conf = &FragmentListConfig{}
	err = json.Unmarshal(data, conf)
	if err != nil {
		dvlog.PrintfError("Your file %s has not valid structure: %v", fileName, err)
		return
	}
	if len(conf.Fragments) == 0 {
		dvlog.PrintfError("Your file %s has no fragment lists")
		return
	}
	if conf.MicroServiceName == "" {
		dvlog.PrintfError("Your file %s has fragment with empty microserviceName")
		return
	}
	if conf.MicroServiceName != dvparser.GlobalProperties[fragmentMicroServiceName] {
		dvlog.PrintfError("Your file %s has microserviceName (%s) different from the name specified in %s (%s), but they must coincide", fileName, conf.MicroServiceName, fragmentMicroServiceName, dvparser.GlobalProperties[fragmentMicroServiceName])
		return
	}
	ok = true
	return
}

func isFragmentListConfigurationForProduction(conf *FragmentListConfig) bool {
	n := len(conf.Fragments)
	for i := 0; i < n; i++ {
		js := conf.Fragments[i].JsResources
		k := len(js)
		if k == 0 {
			return false
		}
		for j := 0; j < k; j++ {
			if strings.HasPrefix(js[j], "http") {
				return false
			}
		}
	}
	return true
}

func getMicroServiceTemporaryFileName(isOrigin bool) string {
	path := dvdir.GetTempPathSlashed()
	if path == "" {
		return ""
	}
	name := dvdir.GetSafeFileName(dvparser.GlobalProperties[fragmentMicroServiceName])
	if name == "" {
		dvlog.PrintfError("You must specify %s", fragmentMicroServiceName)
		return ""
	}
	if isOrigin {
		name += "_origin"
	} else {
		name += "_debug"
	}
	return path + "__dobryvechir__debug__fragments__" + name + ".json"
}

func checkSaveProductionFragmentListConfiguration(conf *FragmentListConfig) bool {
	isOrigin := true
	if !isFragmentListConfigurationForProduction(conf) {
		isOrigin = false
	}
	name := getMicroServiceTemporaryFileName(isOrigin)
	if name == "" {
		return false
	}
	configStr, err := json.Marshal(conf)
	if err != nil {
		dvlog.PrintfError("Error converting the config to json: %s", err.Error())
		return false
	}
	err = ioutil.WriteFile(name, configStr, os.ModePerm)
	if err != nil {
		dvlog.PrintfError("Error %s writing the config to file %s", err.Error(), name)
		return false
	}
	return true
}

func retrieveProductionFragmentListConfiguration() (data []byte, ok bool) {
	name := getMicroServiceTemporaryFileName(true)
	if name == "" {
		return
	}
	data, err := ioutil.ReadFile(name)
	if err != nil {
		dvlog.PrintlnError("Already ok")
		return
	}
	ok = true
	return
}

func deregisterFragment() bool {
	headers := map[string]string{"cache-control": "no-cache"}
	url := getMuiUrl()
	if url == "" {
		return false
	}
	microServiceName := dvparser.GlobalProperties[fragmentMicroServiceName]
	if microServiceName == "" {
		dvlog.PrintfError("Please define %s in the properties file", fragmentMicroServiceName)
		return false
	}
	body := "{\"microserviceName\":\"" + microServiceName + "\"}"
	res, err := dvnet.NewRequest("DELETE", url, body, headers, dvnet.AveragePersistentOptions)
	if err != nil {
		dvlog.PrintlnError(string(res))
		dvlog.PrintfError("Error registering mui fragment at %s: %v", url, err)
		return false
	}
	return true

}

func getMuiUrl() string {
	url := dvparser.GlobalProperties[MuiPlatformUrl]
	if url == "" {
		dvlog.PrintfError("You must specify mui url as %s\n", MuiPlatformUrl)
		return ""
	}
	muiUrl, err := dvparser.ConvertByteArrayByGlobalProperties([]byte(url), "MUI_URL")
	if err != nil {
		dvlog.PrintfError("Make sure you specified all constants in %s file dvserver.properties: %v", url, err)
		return ""
	}
	return muiUrl
}

func getMuiFragmentUrl(name string) string {
	url := dvparser.GlobalProperties[MuiListUrl]
	if url == "" {
		dvlog.PrintfError("You must specify mui url as %s\n", MuiListUrl)
		return ""
	}
	muiUrl, err := dvparser.ConvertByteArrayByGlobalProperties([]byte(url), "MUI_URL")
	if err != nil {
		dvlog.PrintfError("Make sure you specified all constants in %s file dvserver.properties: %v", url, err)
		return ""
	}
	if name == "" {
		dvlog.PrintfError("Please define %s in the properties file", fragmentMicroServiceName)
		return ""
	}
	return strings.ReplaceAll(muiUrl, "%name", name)
}

func registerFragment(muiContent []byte) bool {
	headers := map[string]string{"cache-control": "no-cache", "Content-Type": "application/json"}
	url := getMuiUrl()
	if url == "" {
		return false
	}
	n := len(fragmentPartsToBeRemoved)
	for i := 0; i < n; i++ {
		s := fragmentPartsToBeRemoved[i]
		pos := bytes.Index(muiContent, s)
		if pos >= 0 {
			k := pos + len(s)
			muiContent = append(muiContent[:pos], muiContent[k:]...)
		}
	}
	res, err := dvnet.NewRequest("POST", url, string(muiContent), headers, dvnet.AveragePersistentOptions)
	message := string(res)
	if err != nil || strings.Index(message, "SERVER_ERROR") > 0 {
		dvlog.PrintlnError(message)
		dvlog.PrintfError("Error registering mui fragment at %s: %v", url, err)
		return false
	}
	return true
}

func readCurrentFragmentListConfigurationFromCloud(names []string) (conf *FragmentListConfig, ok bool) {
	headers := map[string]string{"cache-control": "no-cache"}
	conf = &FragmentListConfig{Fragments: make([]UiFragment, 0, 10)}
	for _, name := range names {
		url := getMuiFragmentUrl(name)
		if url == "" {
			continue
		}
		res, err := dvnet.NewRequest("GET", url, "", headers, dvnet.AveragePersistentOptions)
		if err != nil {
			errMessage := err.Error()
			if strings.Index(errMessage, "404") >= 0 {
				continue
			}
			dvlog.PrintlnError(string(res))
			dvlog.PrintfError("Error registering mui fragment at %s: %v", url, errMessage)
			return
		}
		fragment := &FragmentItemConfig{}
		err = json.Unmarshal(res, fragment)
		if err != nil {
			dvlog.PrintfError("Error in structure of mui fragment in %s ", string(res))
			conf = nil
			return
		}
		if fragment.Fragment.FragmentName != "" {
			conf.Fragments = append(conf.Fragments, fragment.Fragment)
			if conf.MicroServiceName == "" {
				conf.MicroServiceName = fragment.MicroServiceName
			}
		}
	}
	if len(conf.Fragments) == 0 || conf.MicroServiceName == "" || !isFragmentListConfigurationForProduction(conf) {
		return
	}
	ok = true
	return
}
