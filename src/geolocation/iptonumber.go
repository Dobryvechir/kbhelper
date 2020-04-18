// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"   
	"os"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("iptonumber converts number to ip4/ip6 and back")
		fmt.Println("iptonumber  <number | ip4 | ip6>")
		return
	}
	n := os.Args[1]
	ip, ok := ReadIPOrString(n)
	if ok {
		fmt.Println(WriteBufInAllPresentations(ip))
	} else {
		fmt.Printf("%s is neither a valid number nor valid IP4 nor valid IP6", n)
	}
}
