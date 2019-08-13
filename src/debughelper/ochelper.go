// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"github.com/Dobryvechir/dvserver/src/dvoc"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"log"
)

func deleteCurrentPod() bool {
	name, ok := getCurrentPodName(false)
	if !ok {
		return false
	}
	return dvoc.DeletePod(name)

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
	return dvoc.GetPodName(name, silent)
}

func getMicroServiceDeleteOption() int {
	mode := dvoc.MicroServiceDeleteForced
	switch dvparser.GlobalProperties["MICROSERVICE_DELETE_MODE"] {
	case "0", "FORCED", "":
		mode = dvoc.MicroServiceDeleteForced
	case "1", "SAVED":
		mode = dvoc.MicroServiceDeleteTrySaveAndForceDelete
	case "2", "SAFE":
		mode = dvoc.MicroServiceDeleteSaveAndSafeDelete
	default:
		log.Printf("Unknown MICROSERVICE_DELETE_MODE option (available are FORCED (0), SAVED (1), SAFE (2))")
	}
	return mode
}

func getMicroServiceSaveNonDebug() bool {
	saveNonDebug := true
	switch dvparser.GlobalProperties["MICROSERVICE_SAVE_NON_DEBUG_ONLY"] {
	case "false":
		saveNonDebug = false
	}
	return saveNonDebug
}

func downCurrentMicroService() bool {
	name, ok := getCurrentPodName(false)
	if !ok {
		return false
	}
	return dvoc.DownWholeMicroService(name, getMicroServiceDeleteOption(), getMicroServiceSaveNonDebug())
}
