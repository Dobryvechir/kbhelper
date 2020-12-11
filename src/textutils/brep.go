// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)
/*********************************************************************
binreplace fileName addr newhex(2 for each byte)
********************************************************************/
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	copyright = "Copyright by Danyil Dobryvechir 2020 Christmas"
)

func readHexLetter(b byte) (byte, error) {
	if b >= '0' && b <= '9' {
		return b - '0', nil
	}
	if b >= 'a' && b <= 'f' {
		return b - 'a' + 10, nil
	}
	if b >= 'A' && b <= 'F' {
		return b - 'A' + 10, nil
	}
	return 0, errors.New("Incorrect hex character " + strconv.Itoa(int(b)))
}

func readHexInBuf(data []byte) ([]byte, error) {
	n := len(data)
	if (n & 1) != 0 {
		data = append([]byte{'0'}, data...)
		n++
	}
	m := n >> 1
	r := make([]byte, m)
	for i := 0; i < m; i++ {
		p := i << 1
		h, err := readHexLetter(data[p])
		if err != nil {
			return nil, err
		}
		l, err := readHexLetter(data[p+1])
		if err != nil {
			return nil, err
		}
		r[i] = (h << 4) | l
	}
	return r, nil
}

func readHexAsInt(data []byte) (int, error) {
	b, err := readHexInBuf(data)
	if err != nil {
		return 0, err
	}
	a := 0
	n := len(b)
	for i := 0; i < n; i++ {
		a = (a << 8) | int(b[i])
	}
	return a, nil
}

func changeInBuf(buf []byte, addr string, hx string) error {
	a, err := readHexAsInt([]byte(addr))
	if err != nil {
		return err
	}
	b, err := readHexInBuf([]byte(hx))
	if err != nil {
		return err
	}
	n := len(buf)
	m := len(b)
	if a < 0 || a > n-m {
		return errors.New("The address " + strconv.Itoa(a) + " is out of bounds [0.." + strconv.Itoa(n) + "-" + strconv.Itoa(m) + "]")
	}
	fmt.Printf("Address %d\n", a)
	for i := 0; i < m; i++ {
		buf[a+i] = b[i]
	}
	return nil
}

func changeInFile(fl string, addr string, hx string) error {
	buf, err := ioutil.ReadFile(fl)
	if err != nil {
		return err
	}
	err = changeInBuf(buf, addr, hx)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fl, buf, 0644)
	return err
}

func main() {
	args := os.Args
	l := len(args)
	if l < 4 {
		fmt.Println(copyright)
		fmt.Println("Command line: <file> <hex addr like 12a> <hex bytes>")
		return
	}
	fl := args[1]
	addr := args[2]
	hx := args[3]
	err := changeInFile(fl, addr, hx)
	if err != nil {
		fmt.Printf("Error: %s %s [%s]: %v\n", fl, addr, hx, err)
	} else {
		fmt.Printf("Modified %s %s %s\n", fl, addr, hx)
	}
}
