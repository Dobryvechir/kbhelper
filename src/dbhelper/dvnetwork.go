// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"os"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("dvnetwork <url property> <method (default - GET)> <header,,,list> <body> <addMessage>")
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
	body := ""
	if l > 3 {
		body = args[3]
		if params[body] != "" {
			body = params[body]
		}
	}
	addMessage := ""
	if l > 4 {
		addMessage = os.Args[4]
	}
	data, err := dvnet.NewRequest(method, url, body, headers)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Printf("%s%s\n", addMessage, string(data))
	}
}
