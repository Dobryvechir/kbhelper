// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

type ChangeVersionItem struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Match  string `json:"match"`
}
type ChangeVersionBlock struct {
	source string              `json:"source"`
	items  []ChangeVersionItem `json:"items"`
}
type ChangeVersionConfig struct {
	places []ChangeVersionBlock `json:"places"`
	format string               `json:"format"`
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Printf(copyright)
		fmt.Printf("\nCommand line: incversion configFileName")
		fmt.Printf("\nConfig structure is as follows:")
		return
	}

}
