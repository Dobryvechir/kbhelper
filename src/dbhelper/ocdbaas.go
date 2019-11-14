// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
)

const (
	ocDbaaSCopyRight = "Copyright by Volodymyr Dobryvechir 2019"
	ocDbaaSHelp      = "ocdbaas <microservice name> <tenant or - if none> <p = postgres or m=mongodb (default)> <output file>"
)

func presentDbProperties(microServiceName string, m2mToken string, tenantId string, database string, output string) {
	dbaasInfo, err := dvoc.GetDbaasProperties(microServiceName, m2mToken, database, tenantId)
	if err != nil {
		fmt.Printf("Fatal error: %v", err)
		return
	}
	info := fmt.Sprintf("HOST=%s\nPORT=%d\nURL=%s\nDB=%s\nUSERNAME=%s\nPASSWORD=%s\n",
		dbaasInfo.ConnectionProperties.Host,
		dbaasInfo.ConnectionProperties.Port,
		dbaasInfo.ConnectionProperties.Url,
		dbaasInfo.ConnectionProperties.DbName,
		dbaasInfo.ConnectionProperties.UserName,
		dbaasInfo.ConnectionProperties.Password)

	if output != "" {
		err = ioutil.WriteFile(output, []byte(info), 0644)
		if err != nil {
			fmt.Println(info)
			fmt.Printf("Error in %s saving: %v", output, err)
		}
	} else {
		fmt.Println(info)
	}
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	dvoc.OpenShiftAddRoutesTOBeExposed("dbaas-agent")
	l := len(args)
	if l < 2 {
		fmt.Println(ocDbaaSCopyRight)
		fmt.Println(ocDbaaSHelp)
		return
	}
	microServiceName := args[0]
	tenant := args[1]
	if tenant == "-" {
		tenant = ""
	}
	database := dvoc.OcMongoDb
	if l > 2 {
		s := args[2]
		switch s[0] {
		case 'p':
			database = dvoc.OcPostgreSql
		case 'm':
		default:
			fmt.Printf("3rd parameter must be either p (postgresql) or m (mongodb) but not %s", s)
			return
		}
	}
	output := ""
	if l > 3 {
		output = args[3]
	}
	m2mToken := ""
	tenantId := ""
	var err error
	if tenant != "" {
		tenantId, err = dvoc.ResolveTenantIdByTenant(tenant)
		if err != nil {
			fmt.Printf("Fatal error: cannot get tenant id for tenant %s: %v", tenant, err)
			return
		}
	}
	presentDbProperties(microServiceName, m2mToken, tenantId, database, output)
}
