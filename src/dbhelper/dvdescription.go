// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("dvdescription <description file> <output file optionally>")
		return
	}
	output := ""
	if l > 1 {
		output = args[1]
	}
	lookNames := map[string]string{"MICROSERVICE_NAME": "MICROSERVICE_NAME", "OPENSHIFT_MICROSERVICE_NAME": "OPENSHIFT_MICROSERVICE_NAME"}
	result, err := dvparser.LookInDescriptionFile(args[0], lookNames, true)
	if err == nil {
		err = dvparser.PresentMapAsProperties(result, output)
	}
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}
