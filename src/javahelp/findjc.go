// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"
var commandLine = "--D=[folder] --H=<host><port> classes"
var currentDir string

const bufSize = 1 << 20
const bufSize64 = int64(bufSize)

func main() {
	option := collectOptions()
	klassen := collectNonOptions()
	currentDir = option["D"]
	if currentDir == "" {
		currentDir = "./"
	} else {
		c := currentDir[len(currentDir)-1]
		if c != '/' {
			currentDir = currentDir + "/"
		}
	}
	host := option["H"]
	if host != "" {
		startNetService(host)
	} else {
		res := klassenFinden(strings.Join(klassen, ":"))
		log.Println(res)
	}
}

func collectNonOptions() []string {
	args := os.Args
	n := len(args)
	r := make([]string, 0, 7)
	for i := 1; i < n; i++ {
		s := args[i]
		if !strings.HasPrefix(s, "--") {
			r = append(r, s)
		}
	}
	return r
}

func collectOptions() map[string]string {
	args := os.Args
	n := len(args)
	r := make(map[string]string)
	for i := 1; i < n; i++ {
		s := args[i]
		if strings.HasPrefix(s, "--") {
			k := s[2:]
			p := strings.Index(k, "=")
			v := ""
			if p > 0 {
				v = k[p+1:]
				k = k[:p]
			}
			r[k] = v
		}
	}
	return r
}

func startNetService(host string) {
}
func bufferContainsClass(buf []byte,sz int,n int,lst []string,pnt []int) bool {

}

func fileContainsClass(path string, buf []byte, pnt []int, sz int64, n int, lst []string) (bool,error) {
    f, err := os.Open(path)
    if err!=nil {
        return false, err
    }
    nr, err:=f.Read(buf)
    if err!=nil {
        return false, err
    }
    if bufferContainsClass(buf, nr, n, lst, pnt) {
        f.Close()
        return true, nil
    }
    if sz<=bufSize64 {
        f.Close()
        return true, nil
    }
    sz -= int64(nr)
    for sz>0 {

    }
	return false, nil
}

func klassenFinden(klassen string) string {
	liste := dieListeDerKlassenMachen(klassen)
	n := len(liste)
	if n == 0 {
		return "!! No classes specified\n"
	}
	buf := make([]byte, bufSize)
	pnt := make([]int, n)
	res := make([]string, 0, 1024)
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			lowName := strings.ToLower(path)
            if !info.IsDir() && (strings.HasSuffix(lowName, ".class") || strings.HasSuffix(lowName, ".jar") || strings.HasSuffix(lowName, ".war")) {
				 enthaltet,err:= fileContainsClass(path, buf, pnt, info.Size(), n, liste)
				 if err!=nil {
                    res = append(res, "!!!! " + path + " !!! " + err.Error())
                 } else if enthaltet {
					res = append(res, "-- "+path+" --- "+strconv.FormatInt(info.Size(), 10)+" ---- "+info.ModTime().Format("2006-01-02 15:04:05"))
				}
			}
			fmt.Println(path, info.Size())
			return nil
		},
	)
	if err != nil {
		res = append(res, "!!! "+err.Error())
	}
	return strings.Join(res, "\n")
}

func dieListeDerKlassenMachen(s string) []string {
	s = strings.TrimSpace(strings.ReplaceAll(s, ";", ":"))
	l := strings.Split(s, ":")
	n := len(l)
	for i := 0; i < n; i++ {
		t := strings.TrimSpace(l[i])
		if t == "" {
			l = append(l[:i], l[i+1:]...)
			n = len(l)
			i--
		} else {
			t = strings.ReplaceAll(t, ".", "/")
			l[i] = t
		}
	}
	return l
}
