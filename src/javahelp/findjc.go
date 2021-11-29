// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"
var commandLine = "--d=[folder] --h=<host><port> classes"
var currentDir string

const bufSize = 1 << 20
const bufSize64 = int64(bufSize)

func main() {
	option := collectOptions()
	klassen := collectNonOptions()
	currentDir = option["d"]
	if currentDir == "" {
		currentDir = "./"
	} else {
		c := currentDir[len(currentDir)-1]
		if c != '/' {
			currentDir = currentDir + "/"
		}
	}
	host := option["h"]
	if host != "" {
		startNetService(host)
	} else {
		if len(klassen) == 0 {
			fmt.Println(copyright)
			fmt.Println(commandLine)
			return
		}
		res := klassenFinden(strings.Join(klassen, ":"))
		fmt.Println(res)
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
	http.HandleFunc("/find", func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["cl"]
		if !ok || len(keys) == 0 {
			fmt.Fprintln(w, "!! cl is missing\n")
			return
		}
		d := strings.Join(keys, " ")
		rs := klassenFinden(d)
		fmt.Fprintln(w, rs)
	})
	if host[0] >= '0' && host[0] <= '9' {
		host = ":" + host
	}
	s := &http.Server{
		Addr:           host,
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Println("Listening at " + host)
	log.Fatal(s.ListenAndServe())
}

func findClassEntry(buf []byte, m int, src []byte) (bool, int) {
	n := len(src)
	b1 := src[0]
haupt:
	for i := 0; i < m; i++ {
		if buf[i] == b1 {
			r := m - i
			if r < n {
				for j := 1; j < r; j++ {
					if buf[i+j] != src[j] {
						continue haupt
					}
				}
				return false, r
			} else {
				for j := 1; j < n; j++ {
					if buf[i+j] != src[j] {
						continue haupt
					}
				}
				return true, 0
			}
		}
	}
	return false, 0
}

func bufferContainsClass(buf []byte, sz int, n int, lst [][]byte, pnt []int) bool {
	for i := 0; i < n; i++ {
		f, p := findClassEntry(buf, sz, lst[i])
		pnt[i] = p
		if f {
			return true
		}
	}
	return false
}

func fileContainsClass(path string, buf []byte, pnt []int, sz int64, n int, lst [][]byte) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	nr, err := f.Read(buf)
	if nr == 0 {
		f.Close()
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if bufferContainsClass(buf, nr, n, lst, pnt) {
		f.Close()
		return true, nil
	}
	if sz <= bufSize64 {
		f.Close()
		return false, nil
	}
	sz -= int64(nr)
	for sz > 0 {
		nr, err = f.Read(buf)
		if err != nil {
			return false, err
		}
		if nr == 0 {
			f.Close()
			return false, nil
		}
		for i := 0; i < n; i++ {
			k := pnt[i]
			if k > 0 {
				m := len(lst[i]) - k
				if m > nr {
					continue
				}
				if fnd, _ := findClassEntry(buf, m, lst[i][k:]); fnd {
					f.Close()
					return true, nil
				}
			}
		}
		if bufferContainsClass(buf, nr, n, lst, pnt) {
			f.Close()
			return true, nil
		}
		sz -= int64(nr)
	}
	f.Close()
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
	found := false
	err := filepath.Walk(currentDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			lowName := strings.ToLower(path)
			if !info.IsDir() && (strings.HasSuffix(lowName, ".class") || strings.HasSuffix(lowName, ".jar") || strings.HasSuffix(lowName, ".war")) {
				enthaltet, err := fileContainsClass(path, buf, pnt, info.Size(), n, liste)
				if err != nil {
					res = append(res, "!!!! "+path+" !!! "+err.Error())
				} else if enthaltet {
					found = true
					res = append(res, "-- "+path+" --- "+strconv.FormatInt(info.Size(), 10)+" ---- "+info.ModTime().Format("2006-01-02 15:04:05"))
				}
			}
			return nil
		},
	)
	if err != nil {
		res = append(res, "!!! "+err.Error())
	}
	if !found {
		res = append(res, "!! No files found")
	}
	return strings.Join(res, "\n")
}

func dieListeDerKlassenMachen(s string) [][]byte {
	s = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, ";", " "), ":", " "), ".", "/"))
	l := strings.Split(s, " ")
	n := len(l)
	res := make([][]byte, 0, 3)
	for i := 0; i < n; i++ {
		t := strings.TrimSpace(l[i])
		if t != "" {
			res = append(res, []byte(t))
		}
	}
	return res
}
