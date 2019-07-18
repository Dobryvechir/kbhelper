// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"context"
	"encoding/json"
	"github.com/Dobryvechir/dvserver/src/dvnet"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"log"
	"strings"
)

type RefInfo struct {
	ActualName 	  string
	MaskedName	  string
}

type RefRule struct {
	maskInfo string
	refInfo  []*RefInfo
}

const (
	debugSource = "DEBUG_SOURCE"
	HashSign = "___h_a_s_h___"
)

func isLocalDebugSource(src string) bool {
	return !strings.HasPrefix(src,"http")
}

func getDebugSource() string {
	src:=dvparser.GlobalProperties[debugSource]
	if src=="" {
		log.Printf("You must specify %s in properties file", debugSource)
		return ""
	}
	if strings.Index(src,"{{{")>=0 {
		s, err:=dvparser.ConvertByteArrayByGlobalProperties([]byte(src), debugSource)
		if err!=nil {
			log.Println(err.Error())
			return ""
		}
		src = s
	}
	if isLocalDebugSource(src) {
		c:=src[len(src)-1]
		if c=='/' || c=='\\' {
			src = src[:len(src)-1]
			if src=="" {
				log.Printf("Empty %s", debugSource)
				return ""
			}
		}
	}
	return src
}

func isGoodHash(s string) bool {
	if s=="{{hash}}" || s=="*" || s=="**"{
		return true
	}
	n:=len(s)
	if n==0 {
		return false
	}
	for i:=0;i<n;i++ {
		c:=s[i]
		if !(c>='a' && c<='f' || c>='A' && c<='F' || c>='0' && c<='9') {
			return false
		}
	}
	return true
}

func getMaskedName(src string) string {
	beg:=0
	s:=src
	pos:=strings.Index(src,"?")
	if pos>=0 {
		s = src[:pos]
	}
	pos = strings.Index(s, "#")
	if pos>=0 {
		s = src[:pos]
	}
	pos = strings.LastIndex(s, "/")
	if pos>=0 {
		beg = pos +1
		s = s[beg:]
	}
	pos = strings.LastIndex(s, ".")
	if pos>0 {
		ext := strings.ToLower(s[pos+1:])
		if ext!="css" && ext!="js" {
			return src
		}
	} else {
		return src
	}
	s = s[:pos]
	pos = strings.Index(s, ".")
	if pos<=0 {
		return src
	}
	if isGoodHash(s[pos+1:]) {
		beg += pos+1
		src = src[:beg] + HashSign + src[len(s):]
	}
	return src
}

func addResInfoFromSource(s string) *RefInfo {
	n:=strings.Index(s,"\"")
	if n<0 {
		n=len(s)
	}
	s = s[:n]
	maskedName :=getMaskedName(s)
	return &RefInfo{s, maskedName}
}

func getRuntimeScripts(data string) []*RefInfo {
	pos:=strings.Index(data,"src=\"runtime")
	if pos<0 {
		return nil
	}
	s:=data[pos+5:]
	pos = strings.Index(s, "</body")
	if pos<=0 {
		return nil
	}
	s = s[:pos]
	pos = strings.Index(s, "src=\"runtime")
	for pos>=0 {
		s = s[pos+5:]
		pos = strings.Index(s, "src=\"runtime")
	}
	res :=make([]*RefInfo, 1, 5)
	res[0] = addResInfoFromSource(s)
	pos = strings.Index(s, "src=\"")
	for pos>=0 {
		s = s[pos+5:]
		res = append(res, addResInfoFromSource(s))
		pos = strings.Index(s, "src=\"")
	}
	return res
}

func makeRefInfo(name string, runtime []*RefInfo) ([]*RefRule, bool) {
	data := strings.TrimSpace(dvparser.GlobalProperties[name])
	if data=="" {
		log.Printf("Please, specify rules %s ({key:value, key:value})", name)
		return nil, false
	}
	if data[0]!='{' || data[len(data)-1]!='}' {
		log.Printf("Rules %s must begin with opening curl bracket and end with closing curl bracket ({key:value, key:value})", name)
		return nil, false
	}
	info:=make(map[string]string)
	err:=json.Unmarshal([]byte(data), info)
	if err!=nil {
		log.Printf("Rules %s must be a string map as follows:{\"key\":\"value\", \"key\":\"value\"}", name)
		return nil, false
	}
	res:=make([]*RefRule, 0, len(info))
	for k,v:=range info {
		listValues:=dvparser.ConvertToNonEmptyList(v)
		refInfo:=make([]*RefInfo, 0, 5)
		n:=len(listValues)
		for i:=0;i<n;i++ {
			if listValues[i]=="RUNTIME" {
				refInfo = append(refInfo, runtime...)
			} else {
				refInfo = append(refInfo, addResInfoFromSource(listValues[i]))
			}
		}
		TODO
		res = append(res, &RefRule{maskInfo: k, refInfo: refInfo})
	}
	return res, true
}

func applyRefRule(s string, rule *RefRule) (string, bool) {
	TODO
}

func convertReferences(src []string,rules []*RefRule) (dst []string) {
	n:=len(src)
	m:=len(rules)
	dst = make([]string, 0, n)
	for i:=0;i<n;i++ {
		s:=src[i]
		d:=s
		for j:=0;j<m;j++ {
			newD, ok:= applyRefRule(s, rules[j])
			if ok {
				d = newD
				break
			}
		}
		if d!="" {
			dst = append(dst, d)
		}
	}
	return
}

func createDebugFragmentListConfig(fragmentListConfig *FragmentListConfig) (conf *FragmentListConfig, ok bool) {
	src:=getDebugSource()
	if src=="" {
		return
	}
	var indexContent string
	if isLocalDebugSource(src) {
		fileName:=src + "/index.html"
		data, err:=ioutil.ReadFile(fileName)
		if err!=nil {
			log.Printf("Since %s does not start with http, it is considered as a folder, which must contain index.html, but index.html was not found", fileName)
			return
		}
		indexContent = string(data)
	} else {
		res, err:=dvnet.NewRequest("GET", src, "", nil, 30)
		if err!=nil {
			log.Printf("Cannot GET %s", src)
			return
		}
		indexContent = string(res)
	}
	runtimeInfo:=getRuntimeScripts(indexContent)
	cssRules, okCss:=makeRefInfo("CSS_REPLACEMENT", nil)
	jsRules, okJs:=makeRefInfo("JS_REPLACEMENT", runtimeInfo)
	if !okCss || !okJs {
		return
	}
	n:=len(fragmentListConfig.Fragments)
	conf=&FragmentListConfig{MicroServiceName:fragmentListConfig.MicroServiceName,Fragments:make([]UiFragment, n)}
	for i:=0;i<n;i++ {
		orig:=fragmentListConfig.Fragments[i]
		js:=convertReferences(orig.JsResources, jsRules)
		css:=convertReferences(orig.CssResources, cssRules)
		if !okJs || !okCss {
			return
		}
		conf.Fragments[i] = UiFragment{FragmentName:orig.FragmentName, JsResources:js, CssResources:css, Labels:orig.Labels}
	}
	ok = true
	return
}

func convertListConfigToJson(fragmentListConfig *FragmentListConfig) (data []byte, ok bool) {
	data, err:=json.Marshal(fragmentListConfig) {
		if err!=nil {
			log.Printf("Error converting fragment to json: %s", err.Error())
			return
		}
	}
	ok = true
	return
}

func runDvServer() bool {
	src:=getDebugSource()
	if src=="" {
		return false
	}
	//define port and using default 80 is no
	//define extraServer if net, or baseServer if base
	//create whole config
	//for base, provide method replacing "__hash__" in js and css to real files
	//for net, provide replaces which __webpack transforms to unique webpack
	//start dvserver
}
