// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"os"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("ipsearch searches in ip4.bin /ip6.bin for ip")
		fmt.Println("ipsearch <number | ip4 | ip6>")
		return
	}
	n := os.Args[1]
	ip, ok := ReadIPOrString(n)
	if ok {
		fmt.Println(WriteBufInAllPresentations(ip))
		code, err := LookupCountryCode(ip)
		if err != nil {
			code = "Error: " + err.Error()
		}
		fmt.Printf("Country code: %s", code)
	} else {
		fmt.Printf("%s is neither a valid number nor valid IP4 nor valid IP6", n)
	}
}
