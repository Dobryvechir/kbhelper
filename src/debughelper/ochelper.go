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

func downCurrentMicroservice() bool {
	name, ok := getCurrentPodName(false)
	if !ok {
		return false
	}
	return dvoc.DownWholeMicroservice(name)
}
