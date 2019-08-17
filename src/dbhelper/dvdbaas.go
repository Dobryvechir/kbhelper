// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"encoding/json"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

type ConnectionPropertiesInfo struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Url      string `json:"url"`
	UserName string `json:"username"`
}

type DbaasInfo struct {
	ConnectionProperties ConnectionPropertiesInfo `json:"connectionProperties"`
}

func presentDbProperties(params map[string]string, m2mToken string, isDbTenantAware bool, prefix string) {
	secretPath := params["MICROSERVICE_PATH"]
	if secretPath == "" {
		panic("Parameter MICROSERVICE_PATH is not defined in the properties")
	}
	url := params["DBAAS_URL"]
	if url == "" {
		panic("Parameter DBAAS_URL is not defined in the properties")
	}
	tenant := ""
	if isDbTenantAware {
		tenant = params["tenant"]
		if tenant == "" {
			panic("Parameter tenant is not defined in the properties")
		}
	}
	body := params["DBAAS_REQUEST"]
	if body == "" {
		panic("Parameter DBAAS_REQUEST is not defined in the properties")
	}
	headers := map[string]string{"Authorization": m2mToken}
	fileName := "/dbaas-common"
	if tenant != "" {
		fileName = "/dbaas-tenant"
		headers["Tenant"] = tenant
	}
	res, err := dvnet.NewJsonRequest("PUT", url, body, headers, dvnet.AveragePersistentOptions)
	if err != nil {
		fmt.Printf("Error getting db properties - %s (url=%s, body=%s, headers=%v)", err.Error(), url, body, headers)
		panic("Fatal error")
	}

	err = ioutil.WriteFile(secretPath+fileName, res, 0644)
	if err != nil {
		fmt.Printf("Cannot write user file " + err.Error())
	}
	dbaasInfo := &DbaasInfo{}
	if err = json.Unmarshal(res, dbaasInfo); err != nil {
		fmt.Printf("db properties (url=%s, body=%s, headers=%v)", url, body, headers)
		fmt.Printf("Cannot get Dbaas Info %s %s", err.Error(), string(res))
	} else {
		fmt.Printf("%s%s=%s\n", prefix, "HOST", dbaasInfo.ConnectionProperties.Host)
		fmt.Printf("%s%s=%d\n", prefix, "PORT", dbaasInfo.ConnectionProperties.Port)
		fmt.Printf("%s%s=%s\n", prefix, "URL", dbaasInfo.ConnectionProperties.Url)
		fmt.Printf("%s%s=%s\n", prefix, "USERNAME", dbaasInfo.ConnectionProperties.UserName)
		fmt.Printf("%s%s=%s\n", prefix, "PASSWORD", dbaasInfo.ConnectionProperties.Password)
	}
}

func getM2MToken(secretPath string) (string, error) {
	m2mToken, err := ioutil.ReadFile(secretPath + "/m2mtoken")
	return string(m2mToken), err
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("dvdbaas <prefix> <'tenant' | 'common'>")
		return
	}
	prefix := args[0]
	isDbTenantAware := true
	if l >= 1 {
		tenantAwareOption := strings.ToLower(args[1])
		switch tenantAwareOption {
		case "tenant", "":
		case "common":
			isDbTenantAware = false
		default:
			panic("The second option must be either 'tenant' or 'common', but you specified " + tenantAwareOption)
		}
	}
	params := dvparser.GlobalProperties
	secretPath := params["MICROSERVICE_PATH"]
	if secretPath == "" {
		panic("Parameter MICROSERVICE_PATH is not defined in the properties")
	}
	m2mToken, err := getM2MToken(secretPath)
	if err != nil || m2mToken == "" {
		panic("Fatal error: cannot get M2MToken: " + err.Error())
	}
	presentDbProperties(params, m2mToken, isDbTenantAware, prefix)
}
