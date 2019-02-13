// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvlog"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func readCredentials(fileName string, pathName string) (user string, ps string) {
	file, err := os.Open(fileName)
	if err != nil {
		panic("File " + fileName + " not found")
	}
	scanner := bufio.NewScanner(file)
	passFileName := pathName + "/password"
	_ = dvlog.EnsureDirForFileExists(passFileName)
	var str []byte
	for scanner.Scan() {
		s := strings.Split(strings.TrimSpace(scanner.Text()), " ")
		if len(s) == 2 {
			switch s[0] {
			case "password:":
				str, err = base64.StdEncoding.DecodeString(s[1])
				if err != nil {
					fmt.Println("error:", err)
					panic("Fatal error")
				}
				ps = string(str)
				err = ioutil.WriteFile(passFileName, []byte(ps), 0644)
				if err != nil {
					fmt.Printf("Cannot write password file " + err.Error())
				}
				if user != "" {
					return
				}
			case "username:":
				str, err = base64.StdEncoding.DecodeString(s[1])
				if err != nil {
					fmt.Println("error:", err)
					panic("Fatal error")
				}
				user = string(str)
				err = ioutil.WriteFile(pathName+"/username", []byte(user), 0644)
				if err != nil {
					fmt.Printf("Cannot write user file " + err.Error())
				}
				if ps != "" {
					return
				}
			}
		}
	}
	panic("Credential file " + fileName + " is wrong")
	return
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("m2mcredentials <credential file> <secret-path>")
		return
	}
	m2mPath := args[1]
	username, passwrd := readCredentials(args[0], m2mPath)
	fmt.Printf("MICROSERVICE_USER=%s\nMICROSERVICE_PASS=%s\nMICROSERVICE_PATH=%s\n", username, passwrd, m2mPath)
}
