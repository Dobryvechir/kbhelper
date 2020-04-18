// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

type AccessToken struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func readCredentials(secretPath string) (user string, pass string, err error) {
	userData, errUser := ioutil.ReadFile(secretPath + "/username")
	user = string(userData)
	passData, errPass := ioutil.ReadFile(secretPath + "/password")
	pass = string(passData)
	err = errPass
	if errUser != nil {
		err = errUser
	}
	return
}

func writeToken(secretPath string, m2mToken string) error {
	return ioutil.WriteFile(secretPath+"/m2mtoken", []byte(m2mToken), 0664)
}

func getM2MToken(m2mTokenUrl string, username string, passwrd string) (string, error) {
	body := map[string]string{"grant_type": "client_credentials",
		"client_secret": passwrd,
		"client_id":     username}
	headers := map[string]string{"cache-control": "no-cache", "Content-Type": "application/x-www-form-urlencoded"}
	var accessToken AccessToken = AccessToken{}
	err := dvnet.LoadStructFormUrlEncoded("POST", m2mTokenUrl, body, headers, &accessToken, dvnet.AveragePersistentOptions)
	if accessToken.TokenType == "" || accessToken.AccessToken == "" {
		dvnet.DvNetLog = true
		err = dvnet.LoadStructFormUrlEncoded("POST", m2mTokenUrl, body, headers, &accessToken, dvnet.AveragePersistentOptions)
		if accessToken.TokenType == "" || accessToken.AccessToken == "" {
			panic("Fatal error")
		}
	}
	return accessToken.TokenType + " " + accessToken.AccessToken, err
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("m2mtoken <specific secret path>")
		return
	}
	secretPath := args[0]
	username, passwrd, err1 := readCredentials(secretPath)
	if err1 != nil {
		panic("Secret path problems: " + secretPath + " : " + err1.Error())
	}
	params := dvparser.GlobalProperties
	m2mTokenUrl := params["M2MTOKEN_URL"]
	if m2mTokenUrl == "" {
		panic("Parameter M2MTOKEN_URL is not defined in the properties")
	}
	m2mToken, err := getM2MToken(m2mTokenUrl, username, passwrd)
	if err != nil {
		panic("Fatal error: cannot get M2MToken: " + err.Error())
	}
	err = writeToken(secretPath, m2mToken)
	if err != nil {
		panic("Fatal error: cannot write M2MToken: " + err.Error())
	}
}
