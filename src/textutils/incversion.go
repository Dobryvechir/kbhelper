// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"strconv"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

type ChangeVersionItem struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Match  string `json:"match"`
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
	needle := []byte(item.Match)
	needleLen := len(needle)
	n := len(data) - needleLen
	c := needle[0]
bigLoop:
	for i := 0; i < n; i++ {
		if data[i] == c {
			for j := 1; j < needleLen; j++ {
				if data[i+j] != needle[j] {
					continue bigLoop
				}
			}
			i += needleLen
			n = len(data)
			for i < n && data[i] <= 32 {
				i++
			}
			if i == n {
				return 0, errors.New("No version after match for " + item.Match + " in " + src)
			}
			return i, nil
		}
	}
	return 0, errors.New("No match for " + item.Match + " in " + src)

}

func processChangeByItem(data []byte, item *ChangeVersionItem, version []byte, newVersion []byte, src string) ([]byte, error) {
	pos, err := findItem(data, item, src)
	if err != nil {
		return data, err
	}
	n := len(data)
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
			for ; pos+dif < n; pos++ {
				data[pos] = data[pos+dif]
			}
			data = data[:n-dif]
		} else {
			rest := data[pos+vlen:]
			data = append(data[:pos:pos], newVersion...)
			data = append(data, rest...)
		}
	}
	return data, nil
}

func readConfigVersion(config *ChangeVersionConfig) (string, error) {
	if len(config.Places) == 0 || config.Places[0].Source == "" || len(config.Places[0].Items) == 0 {
		return "", errors.New("no places specified")
	}
	data, err := ioutil.ReadFile(config.Places[0].Source)
	if err != nil {
		return "", err
	}
	pos, err1 := findItem(data, &config.Places[0].Items[0], config.Places[0].Source)
	if err1 != nil {
		return "", err1
	}
	n := len(data)
	npos := pos
	for npos < n && data[npos] >= '0' && data[npos] <= '9' {
		npos++
	}
	if pos == npos {
		return "", errors.New("no version in first place in " + config.Places[0].Source)
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
