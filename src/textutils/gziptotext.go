package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func convertToGzip(data []byte, base64Option bool) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	res := b.Bytes()
	if base64Option {
		res = []byte(base64.StdEncoding.EncodeToString(res))
	}
	return res, nil
}

func convertFromGzip(data []byte, base64Option bool) (res []byte, err error) {
	if base64Option {
		data, err = base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return
		}
	}
	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return
	}
	res, err = ioutil.ReadAll(r)
	return
}

func main() {
	n := len(os.Args)
	if n < 3 {
		fmt.Println("gziptotext [g - to gzip |t -to text |gb  - to gzip + base64 | tb- to text from gzip+base64] <fileName or wildcard> <optional output fileName>")
		return
	}
	options := os.Args[1]
	file := os.Args[2]
	output := ""
	if n > 3 {
		output = os.Args[3]
	}
	if strings.Index(file, "*") >= 0 || strings.Index(file, "?") >= 0 || strings.Index(file, "[") >= 0 {
		matches, err := filepath.Glob(file)
		if err != nil {
			fmt.Println(err)
			return
		}
		n := len(matches)
		if n == 0 {
			fmt.Println("no file matches your pattern")
		} else {
			for i := 0; i < n; i++ {
				outputReal := output
				if outputReal != "" && i > 0 {
					outputReal = outputReal + strconv.Itoa(i)
				}
				convertGzipToFrom(options, matches[i], outputReal)
			}
		}
	} else {
		convertGzipToFrom(options, file, output)
	}

}
func convertGzipToFrom(options string, file string, output string) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading %s: %v", file, err)
		return
	}
	gzip := options[0] == 'g'
	var ext string
	if gzip {
		ext = ".zip"
	} else {
		if options[0] != 't' {
			fmt.Printf("Not valid option: %s (valid options are g, t, gb, tb)", options)
			os.Exit(1)
			return
		}
		ext = ".txt"
	}
	if output == "" {
		output = file + ext
	}
	base64Option := false
	if len(options) >= 2 {
		if options[1] == 'b' {
			base64Option = true
		} else {
			fmt.Printf("Not valid options: %s (valid options are g, t, gb, tb)", options)
			os.Exit(1)
			return
		}
	}
	var res []byte
	if gzip {
		res, err = convertToGzip(data, base64Option)
	} else {
		res, err = convertFromGzip(data, base64Option)
	}
	if err != nil {
		fmt.Printf("Conversion error: %v", err)
		return
	}
	err = ioutil.WriteFile(output, res, 0644)
	if err != nil {
		fmt.Printf("File writing error %s: %v", output, err)
		return
	}
	fmt.Printf("Converstion successfully written in %s", output)
}
