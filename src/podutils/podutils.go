// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func findPod(src string, pod string) (string, string) {
	project := ""
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("@echo Cannot read file %s: %s\n", src, e.Error())
		fmt.Printf("@exit\n")
		return "error pod", "failed"
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "PROJECT ") {
			project = strings.TrimSpace(s[8:])
		}
		if strings.HasPrefix(s, pod) {
			k := strings.Index(s, " ")
			if k > 0 {
				s = s[:k]
			}
			return s, project
		}
	}
	fmt.Printf("@echo Cannot find pod %s in file %s\n", pod, src)
	fmt.Printf("@exit\n")
	return "error pod", "failed"
}
