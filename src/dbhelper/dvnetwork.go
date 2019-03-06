// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"strconv"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

const (
	Authorization = "Authorization"
	ContentType   = "Content-Type"
)

func getM2MToken(secretPath string) (string, error) {
	m2mToken, err := ioutil.ReadFile(secretPath + "/m2mtoken")
	return string(m2mToken), err
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("dvnetwork <url property> <method (default - GET)> <header,,,list> <body> <addMessage> <repeats>")
		return
	}
	params := dvparser.GlobalProperties
	url := args[0]
	if params[url] != "" {
		url = params[url]
	}
	if strings.Index(url, "http") != 0 {
		err := dvnet.UpdatePropertiesThruNetRequest(url)
		if err != nil {
			panic("Error: " + err.Error())
		}
		return
	}
	method := "GET"
	if l > 1 {
		method = args[1]
	}
	headers := make(map[string]string)
	if l > 2 {
		dvparser.PutDescribedAttributesToMapFromCommaSeparatedList(params, headers, args[2])
	}
	if headers[Authorization] == "M2M" {
		secretPath := params["MICROSERVICE_PATH"]
		if secretPath == "" {
			panic("Parameter MICROSERVICE_PATH is not defined in the properties")
		}
		m2mToken, err := getM2MToken(secretPath)
		headers[Authorization] = m2mToken
		if err != nil {
			panic("Cannot read M2M token" + err.Error())
		}
	}
	body := ""
	if l > 3 {
		body = args[3]
		if params[body] != "" {
			body = params[body]
		}
	}
	addMessage := ""
	if l > 4 {
		addMessage = args[4]
	}
	repeats := 0
	if l > 5 {
		if nrepeats, err1 := strconv.Atoi(args[5]); err1 != nil || nrepeats < 0 {
			fmt.Printf("Incorrect number of repeats: %s\n", args[5])
		} else {
			repeats = nrepeats
		}
	}
	data, err := dvnet.NewRequest(method, url, body, headers, repeats)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		if addMessage == "" || addMessage[:1] != "@" {
			fmt.Printf("%s%s\n", addMessage, string(data))
		} else {
			properties := dvparser.CloneGlobalProperties()
			properties["RESPONSE"] = string(data)
			addMessage, err = dvparser.SmartReadFileAsString(addMessage[1:], properties)
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
			} else {
				fmt.Printf("%s\n", addMessage)
			}
		}
	}
}
