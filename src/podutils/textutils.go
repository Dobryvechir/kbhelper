// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func normalizeForLinux(buf []byte) ([]byte, bool) {
	change := false
	l := len(buf)
	p := 0
	for i := 0; i < l; i++ {
		if buf[i] != 13 {
			buf[p] = buf[i]
			p++
		} else if i+1 < l && buf[i+1] != 10 {
			buf[p] = 10
			p++
		}
	}
	if p > 0 && buf[p-1] != 10 {
		if p == l {
			buf = append(buf, 10)
			return buf, true
		}
		buf[p] = 10
		p++
		change = true
	}
	if p < l {
		buf = buf[:p]
		change = true
	}
	return buf, change
}

func normalizeFileForLinux(src string) ([]byte, bool, error) {
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
		return nil, false, e
	}
	buf, changed := normalizeForLinux(data)
	return buf, changed, nil
}

func removeCRLF(src string) error {
	data, change, err := normalizeFileForLinux(src)
	if err != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, err.Error())
		return err
	}
	if !change {
		fmt.Printf("-l %s\n", src)
	} else {
		err = ioutil.WriteFile(src, data, 0644)
		if err != nil {
			fmt.Printf("Cannot write file %s: %s\n", src, err.Error())
			return err
		}
		fmt.Printf("+l %s\n", src)
	}
	return nil
}

func checkAlreadyPresentLine(buf []byte, line string) bool {
	return strings.Index(string(buf), line) >= 0
}

func addNonRepeatedLine(src string, line string) {
	data, _, err := normalizeFileForLinux(src)
	if err != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, err.Error())
		panic("Fatal error")

	}
	byteLine := []byte(line + "\n")
	if checkAlreadyPresentLine(data, line) {
		fmt.Printf("Line %s is already present\n", line)
	} else {
		data = append(data, byteLine...)
		err = ioutil.WriteFile(src, data, 0644)
		if err != nil {
			fmt.Printf("Cannot write file %s: %s\n", src, err.Error())
			panic("Fatal error")
		}
	}
}

func normalizeForWindows(buf []byte) ([]byte, bool) {
	change := 0
	l := len(buf)
	for i := 0; i < l; i++ {
		c := buf[i]
		if c == 13 {
			if i+1 == l || buf[i+1] != 10 {
				change++
			}
		} else if c == 10 {
			if i == 0 || buf[i-1] != 13 {
				change++
			}
		}
	}
	if l > 0 && buf[l-1] != 10 && buf[l-1] != 13 {
		change += 2
	}
	if change == 0 {
		return buf, false
	}
	n := l + change
	pool := make([]byte, n)
	p := 0
	for i := 0; i < l; i++ {
		c := buf[i]
		if c == 13 {
			pool[p] = c
			p++
			if i+1 == l || buf[i+1] != 10 {
				pool[p] = 10
				p++
			}
		} else if c == 10 {
			if i == 0 || buf[i-1] != 13 {
				pool[p] = 13
				p++
			}
			pool[p] = c
			p++

		} else {
			pool[p] = c
			p++
		}
	}
	if buf[l-1] != 10 && buf[l-1] != 13 {
		pool[p] = 13
		p++
		pool[p] = 10
	}
	return pool, true
}

func normalizeFileForWindows(src string) ([]byte, bool, error) {
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
		return nil, false, e
	}
	buf, checked := normalizeForWindows(data)
	return buf, checked, nil
}

func addCRLF(src string) error {
	data, change, err := normalizeFileForWindows(src)
	if err != nil {
		return err
	}
	if !change {
		fmt.Printf("-w %s\n", src)
	} else {
		e := ioutil.WriteFile(src, data, 0644)
		if e != nil {
			fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
			return e
		}
		fmt.Printf("+w %s\n", src)
	}
	return nil
}
