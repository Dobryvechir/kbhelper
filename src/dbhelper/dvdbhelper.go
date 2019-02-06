// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

type AccessToken struct {
	access_token       string `json:"access_token"`
	token_type         string `json:"token_type"`
	expires_in         int    `json:"expires_in"`
	refresh_expires_in int    `json:"refresh_expires_in"`
	refresh_token      string `json:"refresh_token"`
	notBeforePolicy    int    `json:"not-before-policy"`
	session_state      string `json:"session_state"`
	scope              string `json:"scope"`
}

func readCredentials(fileName string) (user string, ps string) {
	file, err := os.Open(fileName)
	if err != nil {
		panic("File " + fileName + " not found")
	}
	scanner := bufio.NewScanner(file)
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
				if ps != "" {
					return
				}
			}
		}
	}
	panic("Credential file " + fileName + " is wrong")
	return
}

func presentDbProperties(params map[string]string, user string, ps string, m2mToken string) {
	err := ioutil.WriteFile("password", []byte(ps), 0644)
	if err != nil {
		fmt.Printf("Cannot write password file " + err.Error())
	}
	err = ioutil.WriteFile("user", []byte(user), 0644)
	if err != nil {
		fmt.Printf("Cannot write user file " + err.Error())
	}
	err = ioutil.WriteFile("m2m", []byte(m2mToken), 0644)
	if err != nil {
		fmt.Printf("Cannot write m2m file " + err.Error())
	}
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
	res, err1 := dvnet.NewRequest("PUT", url, body, headers)
	if err1 != nil {
		panic("Cannot get db properties: " + err1.Error())
	}
	err = ioutil.WriteFile("dbparams", res, 0644)
	if err != nil {
		fmt.Printf("Cannot write user file " + err.Error())
	}
	fmt.Printf(string(res))

}

func getM2MToken(m2mTokenUrl string, username string, passwrd string) (string, error) {
	body := "grant_type=client_credentials&client_secret=" + passwrd + "&client_id=" + username
	headers := map[string]string{"cache-control": "no-cache", "Content-Type": "application/x-www-form-urlencoded"}
	var accessToken AccessToken = AccessToken{}
	dvnet.DvNetLog = true
	err := dvnet.LoadStruct("POST", m2mTokenUrl, body, headers, &accessToken)
	return accessToken.token_type + " " + accessToken.access_token, err
}

func main() {
	l := len(os.Args)
	if l < 3 {
		fmt.Println(copyright)
		fmt.Println("dvdbhelper <properties file> <credential file>")
		return
	}
	params := dvparser.ReadPropertiesOrPanic(os.Args[1])
	username, passwrd := readCredentials(os.Args[2])
	m2mTokenUrl := params["M2MTOKEN_URL"]
	if m2mTokenUrl == "" {
		panic("Parameter M2MTOKEN_URL is not defined in the properties")
	}
	m2mToken, err := getM2MToken(m2mTokenUrl, username, passwrd)
	if err != nil {
		panic("Fatal error: cannot get M2MToken: " + err.Error())
	}
	presentDbProperties(params, username, passwrd, m2mToken)
}
