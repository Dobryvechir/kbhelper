// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"encoding/json"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

type DbaasInfo struct {
	ConnectionProperties map[string]string `json:"connectionProperties"`
}

func presentDbProperties(params map[string]string, m2mToken string, secretPath string) {
	url := params["DBAAS_URL"]
	if url == "" {
		panic("Parameter DBAAS_URL is not defined in the properties")
	}
	tenant := params["tenant"]
	if tenant == "" {
		panic("Parameter tenant is not defined in the properties")
	}
	body := params["DBAAS_REQUEST"]
	if body == "" {
		panic("Parameter DBAAS_REQUEST is not defined in the properties")
	}
	headers := map[string]string{"Authorization": m2mToken, "Content-Type": "application/json;charset=UTF-8", "Accept": "application/json, application/x-jackson-smile, application/cbor, application/*+json", "Tenant": tenant}
	res, err := dvnet.NewRequest("PUT", url, body, headers)
	if err != nil {
		panic("Cannot get db properties: " + err.Error())
	}
	err = ioutil.WriteFile(secretPath+"/dbaas-info", res, 0644)
	if err != nil {
		fmt.Printf("Cannot write user file " + err.Error())
	}
	dbaasInfo := &DbaasInfo{}
	if err = json.Unmarshal(res, dbaasInfo); err != nil || dbaasInfo.ConnectionProperties == nil {
		fmt.Printf("Cannot get Dbaas Info %v ", err)
	} else {
		prefix := params["MICROSERVICE_NAME"] + "-"
		for k, v := range dbaasInfo.ConnectionProperties {
			fmt.Printf("%s%s=%s\n", prefix, k, v)
		}
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
		fmt.Println("dvdbaas <secret path>")
		return
	}
	secretPath := args[0]
	params := dvparser.GlobalProperties
	m2mToken, err := getM2MToken(secretPath)
	if err != nil || m2mToken == "" {
		panic("Fatal error: cannot get M2MToken: " + err.Error())
	}
	presentDbProperties(params, m2mToken, secretPath)
}
