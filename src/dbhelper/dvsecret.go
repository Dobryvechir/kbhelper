// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"os"
	"strings"
)

var copyrightDvSecret = "Copyright by Volodymyr Dobryvechir 2019"

var help = copyrightDvSecret + "\ndvsecret [folder for all microservices in project, defaults to SECRET_PATH environment variable] [microservice name or * for all, default=*]"

var commonOk = true

func createSingleSecret(folder string, microservice string) {
	if !dvoc.SaveOpenshiftSecret(folder+microservice, microservice) {
		fmt.Printf("Failed to save secrets for %s", microservice)
		commonOk = false
	}
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	params := dvparser.GlobalProperties
	folder := params["SECRET_PATH"]
	l := len(args)
	if l >= 1 && (args[0] == "--help" || args[0] == "version" || args[0] == "-version" || args[0] == "--version") {
		fmt.Println(help)
		fmt.Println("SECRET_PATH defaults to [%s]", folder)
		os.Exit(1)
		return
	}
	if l >= 1 {
		folder = args[0]
	}
	microservice := "*"
	if l >= 2 {
		microservice = args[1]
	}
	if folder == "" {
		fmt.Println(help)
		fmt.Println("SECRET_PATH is not defined and not set in the command line")
		os.Exit(1)
		return
	}
	c := folder[len(folder)-1]
	if c != '\\' && c != '/' {
		folder += "/"
	}
	if strings.Index(microservice, "*") >= 0 {
		list, err := dvoc.GetMicroServiceFullList()
		if err != nil {
			commonOk = false
			fmt.Printf("Failed to get the list of microservices")
			os.Exit(1)
			return
		}
		for _, microservice = range list {
			createSingleSecret(folder, microservice)
		}
	} else {
		createSingleSecret(folder, microservice)
	}
	if commonOk {
		fmt.Printf("Successfully saved in %s", folder)
	} else {
		os.Exit(1)
	}
}
