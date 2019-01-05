// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func convertMapToProperties(aMap map[string]string) []byte {
	lines := make([]string, 0, 300)
	for k, v := range aMap {
		lines = append(lines, k+"="+v)
	}
	sort.Strings(lines)
	return []byte(strings.Join(lines, "\n") + "\n")
}

func placeNumberToBuf(buf []byte, size int, ipStr string) bool {
	for i := 0; i < size; i++ {
		buf[i] = 0
	}
	n := len(ipStr)
	for j := 0; j < n; j++ {
		var carry int = int(ipStr[j]) - 48
		if carry < 0 || carry > 9 {
			return false
		}
		for i := 0; i < size; i++ {
			var val int = int(buf[i])*10 + carry
			buf[i] = byte(val & 255)
			carry = val >> 8
		}
		if carry > 0 {
			return false
		}
	}
	return true
}
func scanGeoLocationCsvFile(src string, dst string, countryDst string, size int) {
	file, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	// Set the split function for the scanning operation.
	scanner.Split(bufio.ScanLines)
	// Count the words.
	count := 0
	if size == 6 {
		size = 16
	}
	countryMap := make(map[string]string)
	pool := make([]byte, 0, (size+2)*300000)
	buf := make([]byte, size+2)
	for scanner.Scan() {
		t := scanner.Text()
		p1 := strings.Index(t, ",")
		if p1 > 0 && t[:1] == "\"" && t[p1-1:p1] == "\"" {
			count++
			ipStr := t[1 : p1-1]
			p2 := strings.Index(t[p1+1:], ",") + p1 + 1
			if p2 < p1+1 || t[p2+1:p2+2] != "\"" {
				fmt.Printf("Incorrect format at %d in %s\n", count, t)
				continue
			}
			countryCode := t[p2+2 : p2+4]
			p2 += 4
			if countryCode == "-\"" {
				p2--
				countryCode = "--"
			}
			if t[p2:p2+3] != "\",\"" {
				fmt.Printf("Incorrect country code at %d in %s\n", count, t)
				continue
			}
			if _, ok := countryMap[countryCode]; !ok {
				rest := t[p2+3:]
				p2 = strings.Index(rest, "\"")
				if p2 <= 0 {
					fmt.Printf("Incorrect country at %d in %s\n", count, t)
				} else {
					countryMap[countryCode] = rest[:p2]
				}
			}
			if placeNumberToBuf(buf, size, ipStr) {
				buf[size] = countryCode[0]
				buf[size+1] = countryCode[1]
				pool = append(pool, buf...)
			} else {
				fmt.Println("Incorrect IP number in first position at %d in %s\n", count, t)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		panic("reading error:" + err.Error())
	}
	if err := ioutil.WriteFile(dst, pool, 0644); err != nil {
		panic("writing error:" + err.Error())
	}
	countryPool := convertMapToProperties(countryMap)
	if err := ioutil.WriteFile(countryDst, countryPool, 0644); err != nil {
		panic("country writing error:" + err.Error())
	}

	fmt.Printf("Processed %d records and saved in %s\n", count, dst)
}

func main() {
	l := len(os.Args)
	if l < 3 || os.Args[2] != "4" && os.Args[2] != "6" {
		fmt.Println(copyright)
		fmt.Println("csvtobin processes geolocation files in csv format and produces minimum binaries to search for the countries")
		fmt.Println("csvtobin <src file name in csv format> <4 | 6 (4 for IP4, 6 for IP6)> <dst file name (defaults to ipX.bin)> <country dst file name(defaults to countries.properties)>")
		return
	}
	src := os.Args[1]
	size := 4
	dst := "ip4.bin"
	if os.Args[2] == "6" {
		size = 6
		dst = "ip6.bin"
	}
	countries := "countries.properties"
	if l >= 4 {
		dst = os.Args[3]
	}
	if l >= 5 {
		countries = os.Args[4]
	}
	scanGeoLocationCsvFile(src, dst, countries, size)
}
