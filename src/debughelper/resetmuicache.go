// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	atStartSetEnv     = "AT_START_SET_ENV"
	atFinishSetEnv    = "AT_FINISH_SET_ENV"
	atStartScaleZero  = "AT_START_SCALE_ZERO"
	atStartScaleOne   = "AT_START_SCALE_ONE"
	atFinishScaleOne  = "AT_FINISH_SCALE_ONE"
	atFinishScaleZero = "AT_FINISH_SCALE_ZERO"
	atStartReset      = "AT_START_RESET"
	atFinishReset     = "AT_FINISH_RESET"
	atStartEnvWait    = "AT_START_ENV_WAIT"
	atFinishEnvWait   = "AT_FINISH_ENV_WAIT"
)

func resetMuiCache() bool {
	return true
}

func openShiftConvertListToMap(list string) (res map[string][]string, ok bool) {
	items := strings.Split(list, ";;")
	res = make(map[string][]string)
	n := len(items)
	ok = true
	for i := 0; i < n; i++ {
		s := strings.TrimSpace(items[i])
		if s == "" {
			continue
		}
		pos := strings.Index(s, ":")
		if pos <= 0 {
			log.Println("Format for environment is as follows:")
			log.Println("microservice:envName=envName;;microservice:envName=envValue")
			log.Printf("You missed : at %d in %s ", i, list)
			return nil, false
		}
		name := strings.TrimSpace(s[:pos])
		rest := strings.TrimSpace(s[pos+1:])
		if strings.Index(name, " ") >= 0 || len(name) == 0 {
			log.Println("Format for environment is as follows:")
			log.Println("microservice:envName=envName;;microservice:envName=envValue")
			log.Printf("You have extra space in microservice name at %d in %s ", i, list)
			return nil, false
		}
		pos = strings.Index(rest, "=")
		if pos <= 0 {
			log.Println("Format for environment is as follows:")
			log.Println("microservice:envName=envName;;microservice:envName=envValue")
			log.Printf("You did not specify = at %d in %s ", i, list)
			return nil, false
		}
		k := strings.TrimSpace(rest[:pos])
		v := strings.TrimSpace(rest[pos+1:])
		if k == "" || strings.Index(k, " ") >= 0 || v == "" || strings.Index(v, " ") >= 0 {
			log.Println("Format for environment is as follows:")
			log.Println("microservice:envName=envName;;microservice:envName=envValue")
			if k == "" {
				log.Printf("You did not specify envName at %d in %s ", i, list)
			}
			if strings.Index(k, " ") >= 0 {
				log.Printf("You specified envName with has extra space at %d in %s ", i, list)
			}
			if v == "" {
				log.Printf("You did not specify envValue at %d in %s ", i, list)
			}
			if strings.Index(v, " ") >= 0 {
				log.Printf("You specified envValue with has extra space at %d in %s ", i, list)
			}
			return nil, false
		}
		k = k + "=" + v
		if res[name] == nil {
			res[name] = make([]string, 1, n)
			res[name][0] = k
		} else {
			res[name] = append(res[name], k)
		}
	}
	return
}

func openShiftSetEnv(list string) bool {
	envs, ok := openShiftConvertListToMap(list)
	if !ok {
		return false
	}
	res := true
	for k, v := range envs {
		if !openShiftSetEnvironment(k, v) {
			res = false
		}
	}
	return res
}

func openShiftScale(list string, replicas int) bool {
	microServices := dvparser.ConvertToNonEmptyList(list)
	n := len(microServices)
	res := true
	for i := 0; i < n; i++ {
		if !openShiftScaleToReplicas(microServices[i], replicas) {
			res = false
		}
	}
	return res
}

func openShiftResetUp(service string) {
	time.Sleep(60 * time.Second)
	openShiftScaleToReplicas(service, 1)
}

func openShiftReset(list string) bool {
	microServices := dvparser.ConvertToNonEmptyList(list)
	n := len(microServices)
	res := true
	for i := 0; i < n; i++ {
		name := microServices[i]
		if !openShiftScaleToReplicas(name, 0) {
			res = false
		}
		go openShiftResetUp(name)
	}
	return res
}

func atStartExecutions() bool {
	res := true
	if !openShiftSetEnv(dvparser.GlobalProperties[atStartSetEnv]) {
		res = false
	}
	if !openShiftScale(dvparser.GlobalProperties[atStartScaleZero], 0) {
		res = false
	}
	if !openShiftScale(dvparser.GlobalProperties[atStartScaleOne], 1) {
		res = false
	}
	if !processCommonWaitings(atStartEnvWait) {
		res = false
	}
	if !openShiftReset(dvparser.GlobalProperties[atStartReset]) {
		res = false
	}
	return res
}

func processCommonWaitings(prefix string) bool {
	for n := 1; n < 1000000; n++ {
		p := prefix + "_" + strconv.Itoa(n)
		waitCommand := strings.TrimSpace(dvparser.GlobalProperties[p])
		if waitCommand == "" {
			return true
		}
		pos := strings.Index(waitCommand, ",")
		if pos <= 0 {
			log.Printf("in %s the first parameter must be the first idle time in seconds followed by comma", p)
			return false
		}
		idleTime, err := strconv.Atoi(waitCommand[:pos])
		if err != nil {
			log.Printf("in %s the first parameter must be the first idle time (integer) in seconds followed by comma", p)
			return false
		}
		waitCommand = strings.TrimSpace(waitCommand[pos+1:])
		pos = strings.Index(waitCommand, ",")
		if pos <= 0 {
			log.Printf("in %s the second parameter must be the total wait time in seconds followed by comma", p)
			return false
		}
		totalTime, err := strconv.Atoi(waitCommand[:pos])
		if err != nil {
			log.Printf("in %s the second parameter must be the total wait time (integer) in seconds followed by comma", p)
			return false
		}
		waitCommand = strings.TrimSpace(waitCommand[pos+1:])
		kind := 0;
		if strings.HasPrefix(waitCommand, "http:") {
			kind = 1
		} else if strings.HasPrefix(waitCommand, "env:") {
			kind = 2
		} else {
			log.Printf("in %s the third parameter must start with either http: or env:", p)
			return false
		}
		if idleTime > 0 {
			if logDebug {
				log.Printf("idle waiting for %d seconds before %s", idleTime, waitCommand)
			}
			time.Sleep(time.Duration(idleTime) * time.Second)
		}
		if logDebug {
			log.Printf("starting waiting %d seconds for %s", totalTime, waitCommand)
		}
		switch kind {
		case 1:
			if !waitWithNetRequest(totalTime, waitCommand) {
				return false
			}
		case 2:
			if !waitWithEnvSetting(totalTime, waitCommand) {
				return false
			}
		}
	}
	return true
}

func atFinishExecutions() bool {
	res := true
	if !openShiftSetEnv(dvparser.GlobalProperties[atFinishSetEnv]) {
		res = false
	}
	if !openShiftScale(dvparser.GlobalProperties[atFinishScaleZero], 0) {
		res = false
	}
	if !openShiftScale(dvparser.GlobalProperties[atFinishScaleOne], 1) {
		res = false
	}
	if !processCommonWaitings(atFinishEnvWait) {
		res = false
	}
	if !openShiftReset(dvparser.GlobalProperties[atFinishReset]) {
		res = false
	}
	return res
}

func convertToHeader(list []string) (res map[string]string) {
	n := len(list)
	res = make(map[string]string)
	for i := 0; i < n; i++ {
		s := strings.TrimSpace(list[i])
		if s == "" {
			continue
		}
		pos := strings.Index(s, "=")
		if pos <= 0 {
			log.Printf("Incorrect header %s (no =)", s)
			continue
		}
		k := strings.TrimSpace(s[:pos])
		v := strings.TrimSpace(s[pos+1:])
		res[k] = v
	}
	return
}

func waitWithNetRequest(totalTime int, command string) bool {
	params := dvparser.ConvertToNonEmptyList(command)
	url := params[0]
	headers := convertToHeader(params[1:])
	for ; totalTime >= 0; totalTime-- {
		_, err := dvnet.NewRequest("GET", url, "", headers, 5)
		if err == nil {
			return true
		}
		if totalTime > 0 {
			if logDebug {
				log.Printf("Waiting for 1 second")
			}
			time.Sleep(time.Second)
		}
	}
	log.Printf("%s command timed out", command)
	return false
}

func waitWithEnvSetting(totalTime int, command string) bool {
	if !strings.HasPrefix(command, "env:") {
		log.Printf("command %s must start with env:", command)
		return false
	}
	command = strings.TrimSpace(command[4:])
	n := len(command) - 1
	if n < 0 || command[0] != '{' || command[n] != '}' {
		log.Printf("env format is env:{microservice:key=value;;microservice:key=value} but it is %s", command)
		return false
	}
	command = command[1:n]
	for ; totalTime >= 0; totalTime-- {
		if openShiftSetEnv(command) {
			return true
		}
		if totalTime > 0 {
			if logDebug {
				log.Printf("Waiting for 1 second")
			}
			time.Sleep(time.Second)
		}
	}
	log.Printf("%s command timed out", command)
	return false
}
