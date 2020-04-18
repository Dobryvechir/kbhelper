// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvjson"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"github.com/Dobryvechir/dvserver/src/dvtemp"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

const (
	S_INITIAL = iota
	V_EXPECT_COLON
	V_EXPECT_VALUE
	V_KEY
	V_KEY_COLON
	V_VALUE
	V_COMMA_OR_END

	S_EXPECT_PARAMS
	S_EXPECT_PARAMS_COLON
	S_EXPECT_PARAMS_SQUARE_OPEN
	S_EXPECT_PARAM_OPEN
	S_KEY
	S_KEY_COLON_OF_VALUE
	S_KEY_COLON_OF_NAME
	S_KEY_COLON_OF_OTHER
	S_PARAMEND_OR_COMMA
	S_VALUE_OF_VALUE
	S_VALUE_OF_NAME
	S_VALUE_OF_OTHER
	S_PARAMS_END_OR_COMMA
)

const (
	MODE_COMPLEX = iota
	MODE_SIMPLE
	MODE_NOPARAMETERS
)

func isWordLetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_'
}

func presentError(inp []byte, pos int, mes string) {
	line := 1
	column := 1
	for i := 0; i < pos; i++ {
		if inp[i] == 13 || inp[i] == 10 {
			line++
			column = 1
			if i+1 < pos && inp[i] == 13 && inp[i+1] == 10 {
				i++
			}
		}
	}
	fmt.Printf("Error - %s at line %d column %d ", mes, line, column)
	fmt.Println("Fix your template!")
	//panic("Bye")
	os.Exit(1)
}

func replaceVariableByMap(inp []byte, outp []byte, params map[string]string, pos int, n int) ([]byte, int) {
	nSquares := 1
	if pos >= n-3 || inp[pos] != '$' || inp[pos+1] != '{' {
		presentError(inp, pos, "expected ${...}")
	}
	for pos += 2; pos < n && inp[pos] == '{'; {
		nSquares++
		pos++
	}
	startPos := pos
	closedRest := nSquares
	for pos < n && closedRest != 0 {
		for ; pos < n && inp[pos] != '}'; pos++ {
		}
		closedRest = nSquares
		for ; closedRest > 0 && pos < n && inp[pos] == '}'; closedRest-- {
			pos++
		}
	}
	if closedRest != 0 {
		presentError(inp, startPos-nSquares, "unclosed with } variable")
	}
	word := string(inp[startPos : pos-nSquares])
	if rep, ok := params[word]; ok {
		outp = append(outp, []byte(rep)...)
	} else {
		presentError(inp, startPos, "unknown variable "+word)
	}
	return outp, pos
}

func processStringValue(inp []byte, outp []byte, params map[string]string, pos int, n int) []byte {
	for ; pos < n; pos++ {
		c := inp[pos]
		if pos+1 < n && c == '$' && inp[pos+1] == '{' {
			outp, pos = replaceVariableByMap(inp, outp, params, pos, n)
			pos--
			continue
		}
		outp = append(outp, c)
	}
	return outp
}

func processOpenshiftSimplePart(inp []byte, outp []byte, params map[string]string, pos int, exclude map[string]bool) []byte {
	if exclude != nil {
		for k, _ := range exclude {
			params[k] = "${" + k + "}"
		}
	}
	if outp == nil {
		outp = make([]byte, 0, len(inp)+1024)
	}
	return processStringValue(inp, outp, params, pos, len(inp))
}

func processValueOfValue(inp []byte, outp []byte, key string, valStart int, valEnd int, params map[string]string, quoted bool) []byte {
	if key == "" {
		presentError(inp, valStart, "parameter name must go before the value")
	}
	if quoted {
		outp = append(outp, '"')
	}
	if res, ok := params[key]; ok {
		outp = append(outp, []byte(escapeQuote(res))...)
	} else {
		newPos := len(outp)
		outp = processStringValue(inp, outp, params, valStart, valEnd)
		newEnd := len(outp)
		params[key] = string(outp[newPos:newEnd])
	}
	if quoted {
		outp = append(outp, '"')
	}
	return outp
}

func escapeQuote(data string) string {
	inBuf := []byte(data)
	n := len(inBuf)
	outBuf := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		c := inBuf[i]
		if c == '\\' {
			outBuf = append(outBuf, c, c)
		} else if c == '"' {
			outBuf = append(outBuf, '\\', '"')
		} else {
			outBuf = append(outBuf, c)
		}
	}
	return string(outBuf)
}

func findClosingQuote(inp []byte, pos int) int {
	n := len(inp)
	for ; pos < n; pos++ {
		if inp[pos] == '"' {
			return pos
		}
		if inp[pos] == '\\' {
			pos++
		}
	}
	return -1
}

func processOpenshiftTemplate(inp []byte, params map[string]string, isDebug bool) (r []byte) {
	n := len(inp)
	r = make([]byte, 0, n)
	state := S_INITIAL
	wordStart := 0
	currentName := ""
	level := 0
	standardParams := make(map[string]bool)
	for i := 0; i < n; i++ {
		c := inp[i]
		if c <= 32 {
			r = append(r, c)
			continue
		}
		if isWordLetter(c) {
			for wordStart = i; i+1 < n && isWordLetter(inp[i+1]); i++ {
			}
			switch state {
			case S_VALUE_OF_VALUE:
				r = processValueOfValue(inp, r, currentName, wordStart, i+1, params, false)
				state = S_PARAMEND_OR_COMMA
				currentName = ""
			case S_VALUE_OF_OTHER:
				r = append(r, inp[wordStart:i+1]...)
				state = S_PARAMEND_OR_COMMA
			default:
				presentError(inp, i, "unexpected word "+string(inp[wordStart:i+1]))
			}
			continue
		}
		appendC := true
		switch c {
		case '{':
			switch state {
			case S_INITIAL:
				state = S_EXPECT_PARAMS
			case S_EXPECT_PARAM_OPEN:
				state = S_KEY
			case V_EXPECT_VALUE:
				state = V_KEY
				level++
			default:
				presentError(inp, i, "unexpected {")
			}
		case '}':
			switch state {
			case S_PARAMEND_OR_COMMA:
				state = S_PARAMS_END_OR_COMMA
				if currentName != "" {
					if res, ok := params[currentName]; ok {
						for len(r) > 0 && r[len(r)-1] <= 32 {
							r = r[:len(r)-1]
						}
						r = append(r, []byte(",\n      \"value\": \""+escapeQuote(res)+"\"\n    ")...)
					} else {
						presentError(inp, i, "value for param "+currentName+" is not specified")
					}
				}
			case V_COMMA_OR_END:
				level--
				if level < 0 {
					presentError(inp, i, "unexpected }, no parameters")
				}
			default:
				presentError(inp, i, "unexpected }")
			}
		case ':':
			switch state {
			case S_EXPECT_PARAMS_COLON:
				state = S_EXPECT_PARAMS_SQUARE_OPEN
			case S_KEY_COLON_OF_VALUE:
				state = S_VALUE_OF_VALUE
			case S_KEY_COLON_OF_NAME:
				state = S_VALUE_OF_NAME
			case S_KEY_COLON_OF_OTHER:
				state = S_VALUE_OF_OTHER
			case V_EXPECT_COLON:
				state = V_EXPECT_VALUE
			default:
				presentError(inp, i, "unexpected :")
			}
		case '"':
			appendC = false
			i++
			endPos := findClosingQuote(inp, i)
			if endPos < 0 {
				presentError(inp, i, "unclosed quote")
			}
			simpleCase := true
			keyName := string(inp[i:endPos])
			switch state {
			case S_EXPECT_PARAMS:
				if keyName == "parameters" {
					state = S_EXPECT_PARAMS_COLON
				} else {
					state = V_EXPECT_COLON
				}
			case S_KEY:
				switch keyName {
				case "name":
					state = S_KEY_COLON_OF_NAME
				case "value":
					state = S_KEY_COLON_OF_VALUE
				default:
					state = S_KEY_COLON_OF_OTHER
				}
			case S_VALUE_OF_VALUE:
				state = S_PARAMEND_OR_COMMA
				simpleCase = false
				r = processValueOfValue(inp, r, currentName, i, endPos, params, true)
				currentName = ""
			case S_VALUE_OF_NAME:
				state = S_PARAMEND_OR_COMMA
				currentName = keyName
				if !isDebug {
					standardParams[keyName] = true
				}
			case S_VALUE_OF_OTHER:
				state = S_PARAMEND_OR_COMMA
			case V_EXPECT_VALUE:
				state = V_COMMA_OR_END
			case V_KEY:
				state = V_EXPECT_COLON
			default:
				presentError(inp, i, "unexpected \""+keyName+"\" ["+strconv.Itoa(state)+"]")
			}
			if simpleCase {
				r = append(r, '"')
				r = processStringValue(inp, r, params, i, endPos)
				r = append(r, '"')
			}
			i = endPos
		case '[':
			switch state {
			case S_EXPECT_PARAMS_SQUARE_OPEN:
				state = S_EXPECT_PARAM_OPEN
			default:
				presentError(inp, i, "unexpected [")
			}
		case ']':
			switch state {
			case S_PARAMS_END_OR_COMMA:
				return processOpenshiftSimplePart(inp, r, params, i, standardParams)
			default:
				presentError(inp, i, "unexpected ]")
			}
		case ',':
			switch state {
			case S_PARAMS_END_OR_COMMA:
				state = S_EXPECT_PARAM_OPEN
			case S_PARAMEND_OR_COMMA:
				state = S_KEY
			case V_COMMA_OR_END:
				if level == 0 {
					state = S_EXPECT_PARAMS
				} else {
					state = V_KEY
				}
			default:
				presentError(inp, i, "unexpected ,")
			}
		default:
			presentError(inp, i, "unexpected character "+string(inp[i:i+1]))
		}
		if appendC {
			r = append(r, c)
		}
	}
	presentError(inp, n, "unexpected end of template")
	return
}

func readTemplate(name string) map[string]string {
	err := dvparser.ReadPropertiesFileWithEnvironmentVariablesInCurrentDirectory(name)
	if err != nil {
		fmt.Printf("Error %v", err)
		os.Exit(1)
	}
	return dvparser.GlobalProperties
}

const (
	templateProperties = "template.properties"
)

func main() {
	if _, err := os.Stat(templateProperties); err == nil {
		dvparser.DvServerPropertiesInCurrentFolderFileName = templateProperties
	} else {
		fmt.Printf("No template.properties: %v\n", err)
	}
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("ocdvhelper <openshift template input> <openshift template output> <optional: debug | noparameters | simple | complex>")
		return
	}
	templateInput := args[0]
	templateOutput := args[1]
	isDebug := false
	processMode := MODE_COMPLEX
	if l > 2 {
		options := args[2]
		switch options {
		case "debug":
			isDebug = true
			processMode = MODE_SIMPLE
		case "noparameters":
			processMode = MODE_NOPARAMETERS
		case "simple":
			processMode = MODE_SIMPLE
		case "complex":
			processMode = MODE_COMPLEX
		default:
			fmt.Println("You specified options = " + options + " but only debug or noparameters options are accepted")
			os.Exit(1)
		}
	}
	data, e := ioutil.ReadFile(templateInput)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", templateInput, e.Error())
		os.Exit(1)
	}
	params := dvparser.GlobalProperties
	var res []byte
	switch processMode {
	case MODE_COMPLEX:
		mapInfo := dvparser.CopyStringMap(params)
		res = processOpenshiftTemplate(data, mapInfo, isDebug)
		res = processOpenshiftSimplePart(res, nil, params, 0, nil)
		folder := createUniqueFolderByName(templateOutput)
		produceYamlFilesForObjects(res, folder)
	case MODE_SIMPLE:
		res = processOpenshiftTemplate(data, params, isDebug)
	case MODE_NOPARAMETERS:
		res = processOpenshiftSimplePart(data, nil, params, 0, nil)
	}
	e = ioutil.WriteFile(templateOutput, res, 0664)
	if e != nil {
		fmt.Printf("Cannot write file %s: %s\n", templateOutput, e.Error())
		os.Exit(1)
	}
	fmt.Printf("Ready template written in %s \n", templateOutput)
}

func createUniqueFolderByName(template string) string {
	p := strings.LastIndex(template, ".")
	if p < 0 {
		template += "_create"
	} else {
		template = template[:p] + "_create"
	}
	if err := dvtemp.CreateOrCleanDir(template, 0755); err != nil {
		fmt.Printf("Failed to create dir %s: %v", template, err)
		return ""
	}
	return template + "/"
}

func produceYamlFilesForObjects(d []byte, folder string) {
	if folder == "" {
		return
	}
	res, err := dvjson.JsonFullParser(d)
	if err != nil {
		fmt.Printf("Failed to parse json %v", err)
		return
	}
	res = res.ReadSimpleChild("objects")
	if res == nil {
		fmt.Println("No objects field inside template")
		return
	}
	n := len(res.Fields)
	s := ""
	for i := 0; i < n; i++ {
		p := res.Fields[i]
		if p == nil {
			continue
		}
		kind := p.ReadSimpleChildValue("kind")
		fileName := dvtemp.GetUniqueFileName(folder, kind, ".yaml")
		s += "oc create -f " + fileName + "\n"
		data := p.PrintToYaml(4)
		err = ioutil.WriteFile(fileName, data, 0644)
		if err != nil {
			fmt.Printf("Failed to save to %s: %v", fileName, err)
		}
	}
	err = ioutil.WriteFile(folder+"servicesUp.cmd", []byte(s), 0644)
	if err != nil {
		fmt.Printf("Cannot save serverUp.cmd file: %v", err)
	}
}
