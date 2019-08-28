// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

const (
	Authorization = "Authorization"
	ContentType   = "Content-Type"
)

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
	if strings.HasPrefix(headers[Authorization], "M2M_") {
		microServiceName := headers[Authorization][4:]
		m2mToken, ok := dvoc.GetM2MToken(microServiceName)
		headers[Authorization] = m2mToken
		if !ok {
			panic("Error fatal: Cannot read M2M token")
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
	options := map[string]interface{}{
		"repeats": repeats,
	}
	data, err := dvnet.NewRequest(method, url, body, headers, options)
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
				fileName := properties["SAVE_RESULT"]
				if fileName != "" {
					err = ioutil.WriteFile(fileName, []byte(addMessage), 0644)
					if err != nil {
						fmt.Printf("Cannot save results to %s: %v", fileName, err)
					} else {
						return
					}

				} else {
					fmt.Printf("%s\n", addMessage)
					return
				}
			}
		}
	}
	os.Exit(1)
}
