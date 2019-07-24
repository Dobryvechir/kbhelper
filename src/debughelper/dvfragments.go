// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"encoding/json"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type UiFragment struct {
	FragmentName string   `json:"fragmentName"`
	JsResources  []string `json:"jsResources"`
	CssResources []string `json:"cssResources"`
	Labels       []string `json:"labels"`
}

type FragmentListConfig struct {
	MicroServiceName string       `json:"microserviceName"`
	Fragments        []UiFragment `json:"fragments"`
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
		log.Println("specify FRAGMENT_LIST_CONFIGURATION as a path of file where fragment list is configured")
		return
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("Your FRAGMENT_LIST_CONFIGURATION does not point to a valid file name: %s", fileName)
		return
	}
	conf = &FragmentListConfig{}
	err = json.Unmarshal(data, conf)
	if err != nil {
		log.Printf("Your file %s has not valid structure: %v", fileName, err)
		return
	}
	if len(conf.Fragments) == 0 {
		log.Printf("Your file %s has no fragment lists")
		return
	}
	if conf.MicroServiceName == "" {
		log.Printf("Your file %s has fragment with empty microserviceName")
		return
	}
	if conf.MicroServiceName != dvparser.GlobalProperties[fragmentMicroServiceName] {
		log.Printf("Your file %s has microserviceName (%s) different from the name specified in %s (%s), but they must coincide", fileName, conf.MicroServiceName, fragmentMicroServiceName, dvparser.GlobalProperties[fragmentMicroServiceName])
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

func getMicroServiceTemporaryFileName() string {
	path := getTempPathSlashed()
	if path == "" {
		return ""
	}
	name := getSafeFileName(dvparser.GlobalProperties[fragmentMicroServiceName])
	if name == "" {
		log.Printf("You must specify %s", fragmentMicroServiceName)
		return ""
	}
	return path + "__dobryvechir__debug__fragments__" + name + ".json"
}

func checkSaveProductionFragmentListConfiguration(conf *FragmentListConfig) bool {
	if !isFragmentListConfigurationForProduction(conf) {
		return true
	}
	name := getMicroServiceTemporaryFileName()
	if name == "" {
		return false
	}
	configStr, err := json.Marshal(conf)
	if err != nil {
		log.Printf("Error converting the config to json: %s", err.Error())
		return false
	}
	err = ioutil.WriteFile(name, configStr, os.ModePerm)
	if err != nil {
		log.Printf("Error %s writing the config to file %s", err.Error(), name)
		return false
	}
	return true
}

func retrieveProductionFragmentListConfiguration() (data []byte, ok bool) {
	name := getMicroServiceTemporaryFileName()
	if name == "" {
		return
	}
	data, err := ioutil.ReadFile(name)
	if err != nil {
		log.Println("Already ok")
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
		log.Printf("Please define %s in the properties file", fragmentMicroServiceName)
		return false
	}
	body := "{\"microserviceName\":\"" + microServiceName + "\"}"
	res, err := dvnet.NewRequest("DELETE", url, body, headers, 30)
	if err != nil {
		log.Println(string(res))
		log.Printf("Error registering mui fragment at %s: %v", url, err)
		return false
	}
	return true

}

func getMuiUrl() string {
	url := dvparser.GlobalProperties[MuiPlatformUrl]
	if url == "" {
		log.Printf("You must specify mui url as %s\n", MuiPlatformUrl)
		return ""
	}
	muiUrl, err := dvparser.ConvertByteArrayByGlobalProperties([]byte(url), "MUI_URL")
	if err != nil {
		log.Printf("Make sure you specified all constants in %s file dvserver.properties: %v", url, err)
		return ""
	}
	return muiUrl
}

func getMuiUrlList() string {
	url := dvparser.GlobalProperties[MuiListUrl]
	if url == "" {
		log.Printf("You must specify mui url as %s\n", MuiListUrl)
		return ""
	}
	muiUrl, err := dvparser.ConvertByteArrayByGlobalProperties([]byte(url), "MUI_URL")
	if err != nil {
		log.Printf("Make sure you specified all constants in %s file dvserver.properties: %v", url, err)
		return ""
	}
	name := dvparser.GlobalProperties[fragmentMicroServiceName]
	if name == "" {
		log.Printf("Please define %s in the properties file", fragmentMicroServiceName)
		return ""
	}
	return strings.ReplaceAll(muiUrl, "%name", name)
}

func registerFragment(muiContent []byte) bool {
	headers := map[string]string{"cache-control": "no-cache"}
	url := getMuiUrl()
	if url == "" {
		return false
	}
	res, err := dvnet.NewRequest("POST", url, string(muiContent), headers, 30)
	if err != nil {
		log.Println(string(res))
		log.Printf("Error registering mui fragment at %s: %v", url, err)
		return false
	}
	return true
}

func readCurrentFragmentListConfigurationFromCloud() (conf *FragmentListConfig, ok bool) {
	headers := map[string]string{"cache-control": "no-cache"}
	url := getMuiUrlList()
	if url == "" {
		return
	}
	res, err := dvnet.NewRequest("GET", url, "", headers, 30)
	if err != nil {
		errMessage := err.Error()
		if strings.Index(errMessage, "404") >= 0 {
			ok = true
			return
		}
		log.Println(string(res))
		log.Printf("Error registering mui fragment at %s: %v", url, errMessage)
		return
	}
	conf = &FragmentListConfig{}
	err = json.Unmarshal(res, conf)
	if err != nil {
		log.Printf("Error in structure of mui fragment in %s ", string(res))
		conf = nil
		return
	}
	if !checkSaveProductionFragmentListConfiguration(conf) {
		return
	}
	ok = true
	return
}
