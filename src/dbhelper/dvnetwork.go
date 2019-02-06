// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"os"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func main() {
	l := len(os.Args)
	if l < 3 {
		fmt.Println(copyright)
		fmt.Println("dvnetwork <properties file> <url property> <method (default - GET)> <header,,,list> <body> <addMessage>")
		return
	}
	params := dvparser.ReadPropertiesOrPanic(os.Args[1])
	url := os.Args[2]
	if params[url] != "" {
		url = params[url]
	}
	method := "GET"
	if l > 3 {
		method = os.Args[3]
	}
	headers := make(map[string]string)
	if l > 4 {
		dvparser.PutDescribedAttributesToMapFromCommaSeparatedList(params, headers, os.Args[4])
	}
	body := ""
	if l > 5 {
		body = os.Args[5]
		if params[body] != "" {
			body = params[body]
		}
	}
	addMessage := ""
	if l > 6 {
		addMessage = os.Args[6]
	}
	data, err := dvnet.NewRequest(method, url, body, headers)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		fmt.Printf("%s%s\n", addMessage, string(data))
	}
}
