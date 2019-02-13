// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvlog"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

type GitInfo struct {
      ServiceName string
      Branch string
      Artifactory string
      Latest string
      LatestTime string
      GitFolder string
}

func tryFolder(string folder) (gitInfo *GitInfo, ok bool) {
	data, err:=ioutil.ReadFile(folder + "/.git/HEAD")
        if err!=nil {
              return nil, false
        }       
        s := string(data)
        search:="heads/"
        r:=strings.Index(s, search)
        if r<0 {
             return nil, false   
        }                  
	branch:=strings.TrimSpace(s[r + len(search):])
        gitInfo:=&GitInfo{Branch: branch, GitFolder: folder}
        ok:=true
        return
}

func tryFolders() *GitInfo {

}

func readGitInfo() (gitInfo *GitInfo, err error) {
      gitInfo:=tryFolders()
      if gitInfo==nil {
         err = errors.New("No git folder found neither in the current folder nor by the global variables")
         return
      } 
      return
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Println(copyright)
		fmt.Println("gitinfo <options: R - read info C - update in the cloud A - open artifactory in the browser>")
		return
	}
	//options := args[0]
        gitInfo, err := readGitInfo() 
        if err!=nil {
             panic("Git error: "+err.Error())
        }
	fmt.Printf("GIT_BRANCH=%s\nGIT_ARTIFACTORY=%s\nGIT_LATEST=%s\n", gitInfo.Branch, gitInfo.Artifactory, gitInfo.Latest)
}
