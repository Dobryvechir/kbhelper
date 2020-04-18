// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	copyright = "Copyright by Danyil Dobryvechir 2019"
	maxErrors = 10
)

func getFileInfo(fileName string, seek int64) (file *os.File, size int64) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		message := fmt.Sprintf("Error in %s: %v", fileName, err)
		panic(message)
	}
	size = fileInfo.Size() - seek
	file, err = os.Open(fileName)
	if err != nil {
		message := fmt.Sprintf("Error opening %s: %v", fileName, err)
		panic(message)
	}
	if seek > 0 {
		file.Seek(seek, 0)
	}
	return
}

func readSeek(index int) int64 {
	seek := int64(0)
	if len(os.Args) > index {
		v := strings.Replace(os.Args[index], "0x", "", -1)
		r, _ := strconv.ParseUint(v, 16, 64)
		seek = int64(r)
	}
	return seek
}

func fcBinary(f1 *os.File, f2 *os.File, size int64) bool {
	n := int(size)
	b1 := make([]byte, n)
	b2 := make([]byte, n)
	_, err1 := f1.Read(b1)
	_, err2 := f2.Read(b2)
	if err1 != nil {
		panic("File 1 is not readable")
	}
	if err2 != nil {
		panic("File 2 is not readable")
	}
	count := 0
	var lastError int
	for i := 0; i < n; i++ {
		if b1[i] != b2[i] {
			lastError = i
			count++
			if count <= maxErrors {
				fmt.Printf("at %x 1=%x but 2=%x\n", i, b1[i], b2[i])
			}
		}
	}
	if count > maxErrors {
		fmt.Printf("and %d more, last error is at %x 1=%x 2=%x \n", count-maxErrors, lastError,b1[lastError],b2[lastError])
	}
	return count == 0
}

func main() {
	args := os.Args
	l := len(args)
	if l < 3 {
		fmt.Printf(copyright)
		fmt.Printf("\nCommand line: fcb file1 file2 [seek1 in hex] [seek2 in hex]")
		return
	}
	seek1 := readSeek(3)
	seek2 := readSeek(4)
	file1, len1 := getFileInfo(args[1], seek1)
	file2, len2 := getFileInfo(args[2], seek2)
	res1 := len1 == len2
	if !res1 {
		fmt.Printf("Files have different length: %d (%x) and %d (%x)\n", len1, len1, len2, len2)
		if len2 < len1 {
			len1 = len2
		}
	}
	res := fcBinary(file1, file2, len1)
	if res && res1 {
		fmt.Printf("Complete coincidence between %s and %s!\n", args[1], args[2])
	}
}
