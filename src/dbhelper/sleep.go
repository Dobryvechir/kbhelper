// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	args := os.Args
	timeSec := 5
	if len(args) > 1 {
		timeSec, _ = strconv.Atoi(os.Args[1])
	}
	fmt.Printf("Waiting for %ds\n", timeSec)
	time.Sleep(time.Duration(timeSec) * time.Second)
}
