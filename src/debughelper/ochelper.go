// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"bufio"
	"encoding/base64"
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
	openShiftLogin        = "login https://{{{OPENSHIFT_SERVER}}}.{{{OPENSHIFT_DOMAIN}}}:{{{OPENSHIFT_PORT}}} -u {{{OPENSHIFT_LOGIN}}} -p {{{OPENSHIFT_PASS}}} --insecure-skip-tls-verify=true -n {{{OPENSHIFT_NAMESPACE}}}"
	openShiftProject      = "\"{{{OPENSHIFT_NAMESPACE}}}\""
	openShiftSecrets      = "export --insecure-skip-tls-verify secret %1-client-credentials"
	author                = " -  Volodymyr Dobryvechir 2019"
	openShiftExpose       = "expose svc/%1"
	openShiftEnsureRoutes = "OPENSHIFT_ENSURE_ROUTES"
	openShiftDeleteSecret = "delete secret "
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
	if logDebug {
		log.Printf("Executing: oc %s", params)
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
		log.Printf("Error reading temporary file: %s", fileName)
		log.Print("Probably you have problems with oc (openshift client) program")
		return "", false
	}
	os.Remove(fileName)
	res:=string(data)
	if logDebug {
		log.Println("-------------------START EXECUTING OC RESULT --------------------")
		log.Println(res)
		log.Println("____________________END EXECUTING OC RESULT______________________")
	}
	return res, true
}

func ocLogin() bool {
	if isOCLogined {
		return true
	}
	projectName, err1 := dvparser.ConvertByteArrayByGlobalProperties([]byte(openShiftProject), "OPENSHIFT_NAME")
	if err1 != nil {
		log.Printf("Make sure you specified OPENSHIFT_NAME (project name) in file dvserver.properties")
		return false
	}
	cmdLine, err := dvparser.ConvertByteArrayByGlobalProperties([]byte(openShiftLogin), "oc login parameters")
	if err != nil {
		log.Printf("Make sure you specified all constants related to oc login (%s)  (%v)", openShiftLogin, err)
		return false
	}
	res, ok := runOCCommand(cmdLine)
	if !ok {
		return false
	}
	pnt := strings.Index(res, projectName)
	if pnt < 0 {
		log.Print(res)
		log.Printf("Project %s is missing, specify it in OPENSHIFT_NAME property in dvserver.properties", projectName)
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
	for scanner.Scan() {
		s := strings.Split(strings.TrimSpace(scanner.Text()), " ")
		if len(s) == 2 {
			switch s[0] {
			case "password:":
				str, err := base64.StdEncoding.DecodeString(s[1])
				if err != nil {
					log.Print(info)
					log.Printf("Cannot get secret for microservice %s error: %v", microserviceName, err)
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
					log.Printf("Cannot get secret for microservice %s error: %v", microserviceName, err)
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
	log.Printf("Cannot get secret for microservice %s", microserviceName)
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
		log.Printf("Make sure you specified all constants in %s file dvserver.properties: %v", m2mTokenUrlRaw, err1)
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
			log.Printf("Cannot get M2M Access Token for %s (%v)", microserviceName, err)
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

func getPodName(microserviceName string, silent bool) (name string, ok bool) {
	if !ocLogin() {
		return
	}
	cmdLine := "get pods"
	info, ok := runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to get pods %s", info)
		return
	}
	candidates := make([]string, 0, 2)
	pos := strings.Index(info, microserviceName)
	for pos >= 0 {
		c := uint8(0)
		if pos > 0 {
			c = info[pos-1]
		}
		if c <= ' ' {
			c = info[pos+len(microserviceName)]
			if c == '-' {
				endPos := pos + len(microserviceName) + 1
				for ; endPos < len(info); endPos++ {
					c := info[endPos]
					if !(c == '-' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
						break
					}
				}
				candidates = append(candidates, info[pos:endPos])
				pos = endPos
			}
		}
		pos++
		pos = strings.Index(info[pos:], microserviceName) + pos
	}
	n := len(candidates)
	if n == 0 {
		if !silent {
			log.Printf("Pod for microservice %s does not exist in the cloud's project", microserviceName)
		}
		return
	}
	candidate := candidates[0]
	for j := 1; j < n; j++ {
		s := candidates[j]
		if len(s) < len(candidate) {
			candidate = s
		}
	}
	return candidate, true
}

func getCurrentServiceName() (name string, ok bool) {
	name = dvparser.GlobalProperties[fragmentServiceName]
	if name == "" {
		name = dvparser.GlobalProperties[fragmentMicroServiceName]
		if name == "" {
			log.Printf("you must specify the fragment microservice name in %s in dvserver.properties", fragmentMicroServiceName)
			return
		}
	}
	return name, true
}

func getCurrentPodName(silent bool) (name string, ok bool) {
	name, ok = getCurrentServiceName()
	if !ok {
		return
	}
	return getPodName(name, silent)
}

func deleteCurrentPod() bool {
	name, ok := getCurrentPodName(false)
	if !ok {
		return false
	}
	cmdLine := "delete pod " + name
	info, ok := runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to delete pod %s", name)
		return false
	}
	log.Println(info)
	return true
}

func downWholeMicroservice(microserviceName string) bool {
	if !ocLogin() {
		return false
	}
	cmdLine := "delete all -l name=" + microserviceName
	info, ok := runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
	} else {
		log.Println(info)
	}
	cmdLine = "delete deploymentconfig " + microserviceName
	info, ok = runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
	} else {
		log.Println(info)
	}
	cmdLine = "delete configmaps " + microserviceName + ".monitoring-config"
	info, ok = runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
	} else {
		log.Println(info)
	}
	cmdLine = "delete service " + microserviceName
	info, ok = runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
	} else {
		log.Println(info)
	}
	cmdLine = "delete routes " + microserviceName
	info, ok = runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
	} else {
		log.Println(info)
	}
	return true
}

func downCurrentMicroservice() bool {
	name, ok := getCurrentPodName(false)
	if !ok {
		return false
	}
	return downWholeMicroservice(name)
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

func createClientCredentials(user string, pw string, microserviceName string) bool {
	yaml := "apiVersion: v1\ndata:\n  password: >-\n    " +
		pw +
		"\n  username: " +
		user +
		"\nkind: Secret\nmetadata:\n  name: " +
		microserviceName +
		"-client-credentials" +
		"\n  namespace: " +
		dvparser.GlobalProperties["OPENSHIFT_NAMESPACE"] +
		"\ntype: Opaque\n"
	path := getTempPathSlashed() + "__dobryvechir__debug_fragments_secret.yaml"
	err := ioutil.WriteFile(path, []byte(yaml), 0466)
	if err != nil {
		log.Printf("Cannot temporarily save a secret file %s", path)
		return false
	}
	cmdLine := "create -f " + path
	info, ok := runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
		return false
	} else {
		log.Println(info)
	}
	return true

}

func getIdentityProviderClientCredentials(microServiceName string) (user string, pw string, ok bool) {
	//TODO: create client credentials in keycloak
	res:= base64.StdEncoding.EncodeToString([]byte(microServiceName))
	return res, res, true
}

func registerUserCredentialsWithIdentityProvider(user, pw, microServiceName string) bool {
	if !ocLogin() {
		return false
	}
	cmdLine := strings.ReplaceAll(openShiftSecrets, "%1", "identity-provider")
	info, ok := runOCCommand(cmdLine)
	if !ok {
		return false
	}
	pos := strings.Index(info, "  "+microServiceName+":")
	if pos > 0 {
		return true
	}
	pos = strings.Index(info, "  mui-platform:")
	if pos < 0 {
		log.Println(info)
		log.Println("could not find mui-platform here")
		return false
	}
	line := "  " + microServiceName + ": " + pw + "\n"
	newSecret := info[:pos] + line + info[pos:]
	path := getTempPathSlashed() + "__dobryvechir__debug_fragment_idcred.yaml"
	err := ioutil.WriteFile(path, []byte(newSecret), 0466)
	if err != nil {
		log.Printf("Cannot temporarily save a secret file %s", path)
		return false
	}
	cmdLine = openShiftDeleteSecret + "identity-provider-client-credentials"
	info, ok = runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
		return false
	} else {
		log.Println(info)
	}

	cmdLine = "create -f " + path
	info, ok = runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
		return false
	} else {
		log.Println(info)
	}
	return true
}

func provideKeycloakAutoUpdate() {

	//LATER TODO:  find the way to start autoupdate of keycloak

}

func createMicroservice(microserviceName string, templateImage string) bool {
	if !ocLogin() {
		return false
	}
	_, _, ok := getOpenshiftSecrets(microserviceName)
	if !ok {
		user, pw, ok := getIdentityProviderClientCredentials(microserviceName)
		if !ok {
			return false
		}
		if !createClientCredentials(user, pw, microserviceName) {
			return false
		}
		if !registerUserCredentialsWithIdentityProvider(user, pw, microserviceName) {
			return false
		}
		provideKeycloakAutoUpdate()
	}
	json, err := ioutil.ReadFile(templateImage)
	if err != nil {
		json = composeOpenShiftJsonTemplate(microserviceName, templateImage)
	}
	path := getTempPathSlashed() + "__dobryvechir__debug_fragments_template.json"
	err = ioutil.WriteFile(path, json, 0466)
	if err != nil {
		log.Printf("Cannot temporarily save a template file %s", path)
		return false
	}
	cmdLine := "new-app -f " + path
	info, ok := runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
	} else {
		log.Println(info)
	}
	return true
}

func synchronizeDirectory(podName string, distributionFolder string, htmlFolder string) bool {
	if _, err := os.Stat(distributionFolder); err != nil {
		log.Printf("DISTRIBUTION_FOLDER must point to a real folder, but it points to %s", distributionFolder)
		return false
	}
	if !ocLogin() {
		return false
	}
	cmdLine := "rsync " + distributionFolder + " " + podName + ":" + htmlFolder
	info, ok := runOCCommand(cmdLine)
	if !ok {
		log.Printf("Failed to execute %s", cmdLine)
	} else {
		log.Println(info)
	}
	return true
}
