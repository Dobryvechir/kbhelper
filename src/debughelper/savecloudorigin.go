// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

func getFragmentNames(conf * FragmentListConfig) []string {
	res:=make([]string,0,10)
	n:=len(conf.Fragments)
	for i:=0;i<n;i++ {
		s:=conf.Fragments[i].FragmentName
		if s!="" {
			res = append(res, s)
		}
	}
	return res
}

func saveCloudConfigForThisFragment(fragmentListConfig *FragmentListConfig, token string) bool {
	names:=getFragmentNames(fragmentListConfig)
	if len(names)==0 {
		return true
	}
	cloudConfig, ok := readCurrentFragmentListConfigurationFromCloud(names)
	if !ok {
		return true
	}
	if !checkSaveProductionFragmentListConfiguration(cloudConfig) {
		return false
	}
	if !saveToMuiFragmentDatabase(cloudConfig, token) {
		return false
	}
	return true
}

func checkCloudConfigIsOriginal(config *FragmentListConfig) bool {
	names:=getFragmentNames(config)
	if len(names)==0 {
		return true
	}
	_, ok := readCurrentFragmentListConfigurationFromCloud(names)
	return ok
}

func saveToMuiFragmentDatabase(config *FragmentListConfig, token string) bool {
	//TODO: save to mongo database (collection = "debug_fragments")
	//TODO: id = microserviceName, data = config as json
	return true
}
