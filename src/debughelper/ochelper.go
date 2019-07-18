// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var isOCLogined = false

const (
	authorization         = "Authorization"
	contentType           = "Content-Type"
	openShiftLogin        = "login https://{{{OPENSHIFT_SERVER}}}.{{{OPENSHIFT_DOMAIN}}}:{{{OPENSHIFT_PORT}}} -u {{{OPENSHIFT_LOGIN}}} -p {{{OPENSHIFT_PASS}}} --insecure-skip-tls-verify=true -n {{{OPENSHIFT_NAMESPACE}}}"
	openShiftProject      = "\"{{{OPENSHIFT_NAMESPACE}}}\""
	openShiftSecrets      = "export --insecure-skip-tls-verify secret %1-client-credentials"
	author                = " -  Volodymyr Dobryvechir 2019"
	openShiftExpose       = "expose svc/%1"
	openShiftEnsureRoutes = "OPENSHIFT_ENSURE_ROUTES"
)

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

func getTempPath() string {
	path := os.Getenv("TEMP")
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return path
		}
	}
	path = os.Getenv("TMP")
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			return path
		}
	}
	path = "/temp"
	if _, err := os.Stat(path); err != nil {
		return path
	}
	log.Print("temporary folder is not available (define it is TEMP environment variable)")
	return ""
}

func getTempPathSlashed() string {
	path := getTempPath()
	if path == "" {
		return ""
	}
	if path[len(path)-1] != '/' && path[len(path)-1] != '\\' {
		path += "/"
	}
	return path
}

func getTemporaryFileName() string {
	path := getTempPathSlashed()
	if path == "" {
		return ""
	}
	for i := 0; i < 20000; i++ {
		fileName := path + "dbghelper" + strconv.Itoa(i)
		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			return fileName
		}
	}
	log.Print("Your temporary folder is not accessible, please define a good temporary folder in TEMP environment variable")
	return ""
}

func runOCCommand(params string) (string, bool) {
	fileName := getTemporaryFileName()
	if fileName == "" {
		return "", false
	}
	cmd := exec.Command("ddoc", params+" >"+fileName)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(fileName)
		log.Println(stdoutStderr)
		log.Println("Error: " + err.Error())
		log.Println("You should have installed openshift client (oc) and put it path to PATH environment variable")
		return "", false
	}
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading temporary file: %s", fileName)
		log.Print("Probably you have problems with oc (openshift client) program")
		return "", false
	}
	os.Remove(fileName)
	return string(data), true
}

func ocLogin() bool {
	if isOCLogined {
		return true
	}
	projectName, err1 := dvparser.ConvertByteArrayByGlobalProperties([]byte(openShiftProject), "OPENSHIFT_NAME")
	if err1 != nil {
		fmt.Printf("Make sure you specified OPENSHIFT_NAME (project name) in file dvserver.properties")
		return false
	}
	cmdLine, err := dvparser.ConvertByteArrayByGlobalProperties([]byte(openShiftLogin), "oc login parameters")
	if err != nil {
		fmt.Printf("Make sure you specified all constants related to oc login (%s)  (%v)", openShiftLogin, err)
		return false
	}
	res, ok := runOCCommand(cmdLine)
	if !ok {
		return false
	}
	pnt := strings.Index(res, projectName)
	if pnt < 0 {
		log.Print(res)
		fmt.Printf("Project %s is missing, specify it in OPENSHIFT_NAME property in dvserver.properties", projectName)
		return false
	}
	isOCLogined = openshiftEnsureExposeRoutes()
	return isOCLogined
}

func getOpenshiftSecrets(microserviceName string) (user string, ps string, okFinal bool) {
	cmdLine := strings.ReplaceAll(openShiftSecrets, "%1", microserviceName)
	if !ocLogin() {
		return
	}
	info, ok := runOCCommand(cmdLine)
	if !ok {
		return
	}
	scanner := bufio.NewScanner(strings.NewReader(info))
	var str []byte
	for scanner.Scan() {
		s := strings.Split(strings.TrimSpace(scanner.Text()), " ")
		if len(s) == 2 {
			switch s[0] {
			case "password:":
				str, err = base64.StdEncoding.DecodeString(s[1])
				if err != nil {
					log.Print(info)
					fmt.Printf("Cannot get secret for microservice %s error: %v", microserviceName, err)
					return
				}
				ps = string(str)
				if user != "" {
					okFinal = true
					return
				}
			case "username:":
				str, err := base64.StdEncoding.DecodeString(s[1])
				if err != nil {
					log.Print(info)
					fmt.Printf("Cannot get secret for microservice %s error: %v", microserviceName, err)
					return
				}
				user = string(str)
				if ps != "" {
					okFinal = true
					return
				}
			}
		}
	}
	log.Println(info)
	fmt.Printf("Cannot get secret for microservice %s", microserviceName)
	return
}

func getM2MToken(microserviceName string) (token string, okFinal bool) {
	username, passwrd, ok := getOpenshiftSecrets(microserviceName)
	if !ok {
		return
	}
	m2mTokenUrlRaw := dvparser.GlobalProperties["M2MTOKEN_URL"]

	if m2mTokenUrlRaw == "" {
		log.Print("Specify M2MTOKEN_URL in dvserver.properties")
		return
	}
	m2mTokenUrl, err1 := dvparser.ConvertByteArrayByGlobalProperties([]byte(m2mTokenUrlRaw), "M2M TOKEN URL")
	if err1 != nil {
		fmt.Printf("Make sure you specified all constants in %s file dvserver.properties: %v", m2mTokenUrlRaw, err1)
		return
	}
	body := map[string]string{"grant_type": "client_credentials",
		"client_secret": passwrd,
		"client_id":     username}
	headers := map[string]string{"cache-control": "no-cache", "Content-Type": "application/x-www-form-urlencoded"}
	var accessToken = &AccessToken{}
	err := dvnet.LoadStructFormUrlEncoded("POST", m2mTokenUrl, body, headers, accessToken, 30)
	if accessToken.TokenType == "" || accessToken.AccessToken == "" {
		dvnet.DvNetLog = true
		err = dvnet.LoadStructFormUrlEncoded("POST", m2mTokenUrl, body, headers, &accessToken, 1)
		if accessToken.TokenType == "" || accessToken.AccessToken == "" {
			fmt.Printf("Cannot get M2M Access Token for %s (%v)", microserviceName, err)
			return
		}
	}
	return accessToken.TokenType + " " + accessToken.AccessToken, true
}

func openshiftEnsureExposeRoutes() bool {
	routes := dvparser.ConvertToNonEmptyList(dvparser.GlobalProperties[openShiftEnsureRoutes])
	if len(routes) != 0 {
		for _, v := range routes {
			cmdLine := strings.ReplaceAll(openShiftExpose, "%1", v)
			res, ok := runOCCommand(cmdLine)
			if !ok {
				return false
			}
			if strings.Index(res, "exposed") < 0 && strings.Index(res, "AlreadyExist") < 0 {
				log.Printf("Unrecognized response to %s : %s", cmdLine, res)
			}
		}
	}
	return true
}

func getSafeFileName(src string) string {
	data := []byte(src)
	n := len(data)
	for i := 0; i < n; i++ {
		c := data[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			data[i] = '_'
		}
	}
	return string(data)
}
