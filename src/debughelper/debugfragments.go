// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
)


const (
        programName = "Debug Fragments 1.0" + author 
)

func startDebugFragment() {
        token, ok:= getM2MToken("mui-fragments") 
	if !ok {
             return
        }
        uiConfiguration, ok:=readUiConfiguration()
        if !ok {
              return
        }
        ok = preserveContentDelivery(token)
        if !ok {
              return
        }
        ok = preserveMuiFragments(token)
        if !ok {
              return
        }
        scripts,css,ok := readIndexScripts()
        if !ok {
              return
        }
        muiDebug, ok:= createFragmentForScript(scripts, css, uiConfiguration)
        if !ok {
              return
        }
        ok = registerFragment(token, muiDebug)
        if !ok {
              return
        }
        fmt.Println("Successfully started fragment debug")
}

func finishDebugFragment() {
        token, ok:= getM2MToken("mui-fragments") 
	if !ok {
             return
        }
        ok = removeMuiDebug(token)
        if !ok {
              return
        }
        ok = restoreContentDelivery(token)
        if !ok {
              return
        }
        ok = restoreMuiFragments(token)
        if !ok {
              return
        }
        fmt.Println("Successfully finished fragment debug")
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(programName)
		fmt.Println("Command line: DebugFragment start | DebugFragment finish")
		return
	}
        switch args[0] {
            case "start":
              startDebugFragment()
            case "finish":
              finishDebugFragment()
            default:
		fmt.Println(programName)
		fmt.Println("Command line: DebugFragment start | DebugFragment finish")
        }
}
