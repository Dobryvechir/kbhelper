// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"os"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("dvdescription <description file> <properties file>")
		return
	}
	output := ""
	if l > 2 {
		output = os.Args[2]
	}
	lookNames := map[string]string{"MICROSERVICE_NAME": "MICROSERVICE_NAME", "OPENSHIFT_MICROSERVICE_NAME": "OPENSHIFT_MICROSERVICE_NAME"}
	result, err := dvparser.LookInDescriptionFile(os.Args[1], lookNames, true)
	if err == nil {
		err = dvparser.PresentMapAsProperties(result, output)
	}
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}
