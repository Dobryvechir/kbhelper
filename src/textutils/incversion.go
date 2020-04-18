// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"strconv"
	"strings"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

type ChangeVersionItem struct {
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Match      string `json:"match"`
	BeforePut  string `json:"beforePut"`
	AfterPut   string `json:"afterPut"`
	Src        string `json:"src"`
	BeforeTake string `json:"beforeTake"`
	AfterTake  string `json:"afterTake"`
	// Options can contain V for Version or C for content (or both), if empty, V is assumed
	Options string `json:"options"`
}

type ChangeVersionBlock struct {
	Source string              `json:"source"`
	Items  []ChangeVersionItem `json:"items"`
}

type ChangeVersionConfig struct {
	Places []ChangeVersionBlock `json:"places"`
	Format string               `json:"format"`
}

func findItem(data []byte, item *ChangeVersionItem, src string) (int, error) {
	if item.Match == "" {
		return 0, errors.New("Empty match is not supported yet")
	}
	_, pos := dvparser.FindSubStringSmartInByteArray(data, []byte(item.Match))
	if pos >= 0 {
		n := len(data)
		for pos < n && data[pos] <= 32 {
			pos++
		}
		if pos == n {
			return 0, errors.New("No version after match for " + item.Match + " in " + src)
		}
		return pos, nil
	}
	return 0, errors.New("No match for " + item.Match + " in " + src)
}

func takeExtractFromSource(src string, before string, after string) ([]byte, error) {
	if src == "" {
		return nil, errors.New("File src not specified for " + before + " ... " + after)
	}
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}
	res, stat := dvparser.ExtractFromBufByBeforeAfterKeys(data, []byte(before), []byte(after))
	if stat != 0 {
		if stat == -1 {
			return nil, errors.New("Not found " + before + " in " + src)
		}
		return nil, errors.New("Not found " + after + " in " + src)
	}
	return res, nil
}

func processChangeByItem(data []byte, item *ChangeVersionItem, version []byte, newVersion []byte, src string) ([]byte, error) {
	isVersion := item.Options == "" || strings.Contains(item.Options, "V")
	isContent := strings.Contains(item.Options, "C")
	if !isVersion && !isContent {
		return nil, errors.New("item options must have either V for version or C for content")
	}
	pos, err := findItem(data, item, src)
	if isVersion && err != nil {
		return data, err
	}
	n := len(data)
	if isVersion {
		vlen := len(version)
		nlen := len(newVersion)
		if pos+vlen > n {
			return data, errors.New("no value found for " + item.Match + " in " + src)
		}
		for i := 0; i < vlen; i++ {
			if data[pos+i] != version[i] {
				return data, errors.New("no value " + string(version) + " goes after " + item.Match + " in " + src)
			}
			if nlen <= vlen {
				data[pos+i] = newVersion[i]
			}
		}
		if vlen != nlen {
			if vlen > nlen {
				pos += nlen
				dif := vlen - nlen
				n -= dif
				for i := pos; i < n; i++ {
					data[i] = data[i+dif]
				}
				data = data[:n]
			} else {
				rest := data[pos+vlen:]
				data = append(data[:pos:pos], newVersion...)
				data = append(data, rest...)
				pos += vlen
				n = len(data)
			}
		} else {
			pos += nlen
		}
	} else {
		pos = 0
	}
	if isContent {
		dat, err := takeExtractFromSource(item.Src, item.BeforeTake, item.AfterTake)
		if err != nil {
			return nil, err
		}
		_, posStart := dvparser.FindSubStringSmartInByteArray(data[pos:], []byte(item.BeforePut))
		if posStart < 0 {
			return nil, errors.New("Not found: " + item.BeforePut)
		}
		posStart += pos
		posEnd, _ := dvparser.FindSubStringSmartInByteArray(data[posStart:], []byte(item.AfterPut))
		if posEnd < 0 {
			return nil, errors.New("Not found: " + item.AfterPut)
		}
		data = dvparser.ReplaceTextInsideByteArray(data, posStart, posEnd + posStart, dat)
	}
	return data, nil
}

func findFirstItemWithVersion(items []ChangeVersionItem) int {
	n := len(items)
	for i := 0; i < n; i++ {
		if items[i].Options == "" || strings.Contains(items[i].Options, "V") {
			return i
		}
	}
	return -1
}

func readConfigVersion(config *ChangeVersionConfig) (string, error) {
	placeSize := len(config.Places)
	if placeSize == 0 {
		return "", errors.New("no places specified")
	}
	place := 0
	itemNr := -1
	for place < placeSize {
		itemNr = findFirstItemWithVersion(config.Places[place].Items)
		if itemNr >= 0 {
			break
		}
		place++
	}
	if place >= placeSize {
		return "0", nil
	}
	if config.Places[place].Source == "" {
		return "", errors.New("no source specified")
	}
	data, err := ioutil.ReadFile(config.Places[place].Source)
	if err != nil {
		return "", err
	}
	pos, err1 := findItem(data, &config.Places[place].Items[itemNr], config.Places[place].Source)
	if err1 != nil {
		return "", err1
	}
	n := len(data)
	npos := pos
	for npos < n && data[npos] >= '0' && data[npos] <= '9' {
		npos++
	}
	if pos == npos {
		return "", errors.New("no version in first version place in " + config.Places[place].Source)
	}
	return string(data[pos:npos]), nil
}

func increaseConfigVersion(config *ChangeVersionConfig, version string) (string, error) {
	n, err := strconv.Atoi(version)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(n + 1), nil
}

func increaseVersionInSources(config *ChangeVersionConfig, version []byte, newVersion []byte) error {
	n := len(config.Places)
	for i := 0; i < n; i++ {
		src := config.Places[i].Source
		items := config.Places[i].Items
		p := len(items)
		if src == "" || p == 0 {
			return errors.New("Undefined source in config at place " + strconv.Itoa(i+1))
		}
		data, err := ioutil.ReadFile(src)
		if err != nil {
			return err
		}
		for j := 0; j < p; j++ {
			data, err = processChangeByItem(data, &items[j], version, newVersion, src)
			if err != nil {
				return err
			}
		}
		err = ioutil.WriteFile(src, data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 1 {
		fmt.Printf(copyright)
		fmt.Printf("\nCommand line: incversion configFileName")
		fmt.Printf("\nConfig structure is as follows: {\"places\":[{\"source\":\"file name\",\"items\":[{\"match\":\"string\"},...]},...]}")
		return
	}
	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Printf("Error %s reading file %s", err.Error(), args[0])
		return
	}
	config := &ChangeVersionConfig{}
	err = json.Unmarshal(data, config)
	if err != nil {
		fmt.Printf("Error %s parsing file %s", err.Error(), args[0])
		return
	}
	version, err1 := readConfigVersion(config)
	if err1 != nil {
		fmt.Printf("%s Error defining version in %s", err1.Error(), args[0])
		return
	}
	newVersion, err2 := increaseConfigVersion(config, version)
	if err2 != nil {
		fmt.Printf("%s Error increasing version in %s", err2.Error(), args[0])
		return
	}
	err1 = increaseVersionInSources(config, []byte(version), []byte(newVersion))
	if err1 != nil {
		fmt.Printf("Error %s", err1.Error())
		return
	}
	fmt.Printf("Version increased from %s to %s", version, newVersion)

}
