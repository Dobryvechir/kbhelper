package main

import (
	"fmt"
	"os"
)

const bufSize = 1024 * 1024

var buf = make([]byte, bufSize)

func copyFile(f *os.File, dst string) int {
	wf, err := os.Create(dst)
	if err != nil {
		fmt.Printf("\nW1 Error at %s %v", dst, err)
		return 0
	}
	for {
		n, er := f.Read(buf)
		if er != nil {
			if er.Error() == "EOF" {
				break
			}
			fmt.Printf("\nW2 Error at %s %v", dst, er)
			wf.Close()
			return 0
		}
		if n == 0 {
			break
		}
		_, er = wf.Write(buf[:n])
		if er != nil {
			fmt.Printf("\nW3 Error at %s %v", dst, er)
			wf.Close()
			return 0
		}
		if n != bufSize {
			break
		}
	}
	wf.Close()
	return 1
}

func getQuickBase(s string) string {
	n := len(s)
	i := n - 1
	for ; i >= 0; i-- {
		p := s[i]
		if p == '/' || p == '\\' {
			break
		}
	}
	return s[i+1:]
}

func copyDir(src string, dst string) (dcnt int, fcnt int) {
	file, err := os.Open(src)

	if err != nil {
		fmt.Printf("\nError at %s %v", src, err)
		return 0, 0
	}
	stat, _ := file.Stat()
	mode := stat.Mode()
	name := getQuickBase(src)
	dst = dst + "/" + name
	if mode.IsRegular() {
		fcnt = copyFile(file, dst)
	} else if mode.IsDir() {
		os.Mkdir(dst, 0777)
		dcnt++
		names, er := file.Readdirnames(0)
		if er != nil {
			fmt.Printf("\nError at %s %v", src, er)
		} else {
			n := len(names)
			for i := 0; i < n; i++ {
				r := names[i]
				if r != "." && r != ".." {
					dc, fc := copyDir(src+"/"+r, dst)
					dcnt += dc
					fcnt += fc
				}
			}
		}
	}
	file.Close()
	return
}

func main() {
	n := len(os.Args)
	if n < 2 {
		fmt.Println("dvrsync src dst")
		return
	}
	src := os.Args[1]
	dst := os.Args[2]
	d, f := copyDir(src, dst)
	fmt.Printf("\n%d dirs %d files copied", d, f)
}
