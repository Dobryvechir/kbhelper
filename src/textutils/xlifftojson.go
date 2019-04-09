// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
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
	XMLName xml.Name          `xml:"file"`
        Datatype string           `xml:"datatype,attr"`
	Body    XliffDocumentBody `xml:"body"`
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
		res = append(res, []byte(unit.Target)...)
		res = append(res, '"')
	}
	res = append(res, 13, 10, '}', 13, 10)
	return res, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf(copyright)
		fmt.Printf("xlifftojson <input xliff file> <output json file name>")
		return
	}
	xmlData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	doc, err1 := readXliffDocument(xmlData)
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}
	json, err2 := convertXliffToJson(doc)
	if err2 != nil {
		fmt.Println(err2.Error())
		return
	}
	err = ioutil.WriteFile(os.Args[2], json, 0466)
	if err != nil {
		fmt.Println(err)
		return
	}
}
