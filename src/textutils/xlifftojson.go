// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

type TransUnit struct {
	XMLName xml.Name `xml:"trans-unit"`
	Id      string   `xml:"id,attr"`
	Source  string   `xml:"source"`
	Target  string   `xml:"target"`
}

type XliffDocumentBody struct {
	XMLName    xml.Name    `xml:"body"`
	TransUnits []TransUnit `xml:"trans-unit"`
}

type XliffDocumentFile struct {
	XMLName  xml.Name          `xml:"file"`
	Datatype string            `xml:"datatype,attr"`
	Body     XliffDocumentBody `xml:"body"`
}

type XliffDocument struct {
	XMLName xml.Name          `xml:"xliff"`
	File    XliffDocumentFile `xml:"file"`
}

func readXliffDocument(data []byte) (*XliffDocument, error) {
	var doc *XliffDocument = &XliffDocument{}
	err := xml.Unmarshal(data, doc)
	fmt.Printf("%d entries", len(doc.File.Body.TransUnits))
	return doc, err
}

func readJsonData(data []byte) (res map[string]string, err error) {
	res = make(map[string]string)
	n := len(data)
	exp := byte('{')
	k := ""
	for i := 0; i < n; i++ {
		c := data[i]
		if c <= 32 {
			continue
		}
		if c != exp {
			if exp == ',' && c == '}' {
				return
			}
			err = errors.New("Expected " + string([]byte{exp}) + " but found " + string([]byte{c}))
			return
		}
		switch exp {
		case '"':
			i++
			pos := i
			extra := 0
			for ; i < n; i++ {
				d := data[i]
				if d == '\\' {
					i++
					extra++
				} else if d == '"' {
					break
				}
			}
			if i == n {
				err = errors.New("Unclosed quote at " + strconv.Itoa(pos))
				return
			}
			v := string(data[pos:i])
			if k == "" {
				if v == "" {
					err = errors.New("Empty key is not allowed")
					return
				}
				k = v
				exp = ':'
			} else {
				res[k] = v
				k = ""
				exp = ','
			}
		case ':':
			exp = '"'
		case '{', ',':
			exp = '"'
		}
	}
	err = errors.New("Unclosed json with final }")
	return
}

func convertJsonToXliff(doc map[string]string) (res []byte, err error) {
	res = make([]byte, 0, 100000)
	indent := "			"
	for k, v := range doc {
		s := indent + "<trans-unit datatype=\"html\" id=\"" + k + "\">" +
			indent + "     <source>" + k + "</source>" +
			indent + "    <target>" + v + "</target>" +
			indent + "</trans-unit>"
		res = append(res, []byte(s)...)
	}
	return
}

func getJsonEscapedByteArray(src []byte) []byte {
	n := len(src)
	extra := 0
	for i := 0; i < n; i++ {
		c := src[i]
		if c == '"' || c == '\\' {
			extra++
		}
	}
	if extra == 0 {
		return src
	}
	dst := make([]byte, n+extra)
	j := 0
	for i := 0; i < n; i++ {
		c := src[i]
		if c == '"' || c == '\\' {
			dst[j] = '\\'
			j++
			dst[j] = c
			j++
		} else {
			dst[j] = c
			j++
		}
	}
	return dst
}

func convertXliffToJson(doc *XliffDocument) (res []byte, err error) {
	transUnits := doc.File.Body.TransUnits
	n := len(transUnits)
	res = make([]byte, 0, 65536)
	res = append(res, '{', 13, 10, ' ', ' ', ' ', ' ')
	for i := 0; i < n; i++ {
		unit := &transUnits[i]
		if i != 0 {
			res = append(res, ',', 13, 10, ' ', ' ', ' ', ' ')
		}
		res = append(res, '"')
		res = append(res, []byte(unit.Id)...)
		res = append(res, '"', ':', ' ', '"')
		res = append(res, getJsonEscapedByteArray([]byte(unit.Target))...)
		res = append(res, '"')
	}
	res = append(res, 13, 10, '}', 13, 10)
	return res, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf(copyright)
		fmt.Printf("xlifftojson <input xliff file or json> <output json file name or xliff>")
		return
	}
	xmlData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	var res []byte
	if strings.HasSuffix(strings.ToLower(os.Args[1]), ".xliff") {
		doc, err1 := readXliffDocument(xmlData)
		if err1 != nil {
			fmt.Println(err1.Error())
			return
		}

		res, err = convertXliffToJson(doc)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	} else {
		doc, err1 := readJsonData(xmlData)
		if err1 != nil {
			fmt.Println(err1.Error())
			return
		}

		res, err = convertJsonToXliff(doc)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

	}
	err = ioutil.WriteFile(os.Args[2], res, 0664)
	if err != nil {
		fmt.Println(err)
		return
	}
}
