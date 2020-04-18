// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
)

const (
	author                = " -  Danyil Dobryvechir 2019"
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


func atStartExecutions() bool {
	res := true
	if !dvoc.OpenShiftSetEnv(dvparser.GlobalProperties[atStartSetEnv]) {
		res = false
	}
	if !dvoc.OpenShiftScale(dvparser.GlobalProperties[atStartScaleZero], 0) {
		res = false
	}
	if !dvoc.OpenShiftScale(dvparser.GlobalProperties[atStartScaleOne], 1) {
		res = false
	}
	if !processCommonWaitings(atStartEnvWait) {
		res = false
	}
	if !dvoc.OpenShiftReset(dvparser.GlobalProperties[atStartReset]) {
		res = false
	}
	return res
}

func processCommonWaitings(prefix string) bool {
       return dvoc.ExecuteSequence(prefix)
}

func atFinishExecutions() bool {
	res := true
	if !dvoc.OpenShiftSetEnv(dvparser.GlobalProperties[atFinishSetEnv]) {
		res = false
	}
	if !dvoc.OpenShiftScale(dvparser.GlobalProperties[atFinishScaleZero], 0) {
		res = false
	}
	if !dvoc.OpenShiftScale(dvparser.GlobalProperties[atFinishScaleOne], 1) {
		res = false
	}
	if !processCommonWaitings(atFinishEnvWait) {
		res = false
	}
	if !dvoc.OpenShiftReset(dvparser.GlobalProperties[atFinishReset]) {
		res = false
	}
	return res
}

