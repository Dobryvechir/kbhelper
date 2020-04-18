// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

func main() {
	args := os.Args
	l := len(args)
	if l < 3 {
		fmt.Printf(copyright)
		fmt.Printf("\nCommand line: tailor fileName partSize <optional suffix, default is teil>")
		return
	}
	fileName := args[1]
	partSize := args[2]
	tail := ".tail"
	if l > 3 {
		tail = args[4]
	}
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error %v", err)
		return
	}
	defer file.Close()
	size, err := strconv.Atoi(partSize)
	if err != nil {
		fmt.Printf("Part size must be a number, but it is %s", partSize)
		return
	}
	finalFileName := fileName + tail
	n := 0
	buffer := make([]byte, size)
	for {
		quantity, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}
		n++
		err = ioutil.WriteFile(finalFileName+strconv.Itoa(n), buffer[:quantity], 0644)
		if err != nil {
			fmt.Printf("Error: %v", err)
			return
		}
	}
	fmt.Printf("Successfully writen in %d parts", n)
}
