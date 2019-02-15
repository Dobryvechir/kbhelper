// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"os"
	"strconv"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

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
	panic("Fix your template!")
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

func processOpenshiftSimplePart(inp []byte, outp []byte, params map[string]string, pos int) []byte {
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

func processOpenshiftTemplate(inp []byte, params map[string]string) (r []byte) {
	n := len(inp)
	r = make([]byte, 0, n)
	state := S_INITIAL
	wordStart := 0
	currentName := ""
	level := 0
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
				return processOpenshiftSimplePart(inp, r, params, i)
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
		panic(err.Error())
	}
	return dvparser.GlobalProperties
}

func main() {
	l := len(os.Args)
	if l < 4 {
		fmt.Println(copyright)
		fmt.Println("ocdvhelper <openshift template input> <openshift template output> <properties file>")
		return
	}
	templateInput := os.Args[1]
	templateOutput := os.Args[2]
	propertiesFile := os.Args[3]
	data, e := ioutil.ReadFile(templateInput)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", templateInput, e.Error())
		panic("Fatal error")
	}
	params := readTemplate(propertiesFile)
	res := processOpenshiftTemplate(data, params)
	e = ioutil.WriteFile(templateOutput, res, 0664)
	if e != nil {
		fmt.Printf("Cannot write file %s: %s\n", templateOutput, e.Error())
		panic("Fatal error")
	}
	fmt.Printf("Ready template written in %s \n", templateOutput)
}
