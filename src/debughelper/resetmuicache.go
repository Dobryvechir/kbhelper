// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"log"
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
	time.Sleep(10 * time.Second)
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
	if !openShiftReset(dvparser.GlobalProperties[atStartReset]) {
		res = false
	}
	return res
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
	if !openShiftReset(dvparser.GlobalProperties[atFinishReset]) {
		res = false
	}
	return res
}
