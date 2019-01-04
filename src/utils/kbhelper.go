// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

const (
	JSON_WAIT_VALUE = iota
	JSON_WAIT_ARRAY_VALUE
	JSON_WAIT_KEY
	JSON_WAIT_COLON
	JSON_WAIT_COMMA_OR_CLOSING
	JSON_WAIT_END
)

func findPod(src string, pod string) (string, string) {
	project := ""
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("@echo Cannot read file %s: %s\n", src, e.Error())
		fmt.Printf("@exit\n")
		return "error pod", "failed"
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "PROJECT ") {
			project = strings.TrimSpace(s[8:])
		}
		if strings.HasPrefix(s, pod) {
			k := strings.Index(s, " ")
			if k > 0 {
				s = s[:k]
			}
			return s, project
		}
	}
	fmt.Printf("@echo Cannot find pod %s in file %s\n", pod, src)
	fmt.Printf("@exit\n")
	return "error pod", "failed"
}

func normalizeForLinux(buf []byte) ([]byte, bool) {
	change := false
	l := len(buf)
	p := 0
	for i := 0; i < l; i++ {
		if buf[i] != 13 {
			buf[p] = buf[i]
			p++
		}
	}
	if p > 0 && buf[p-1] != 10 {
		if p == l {
			buf = append(buf, 10)
			return buf, true
		}
		buf[p] = 10
		p++
		change = true
	}
	if p < l {
		buf = buf[:p]
		change = true
	}
	return buf, change
}

func normalizeFileForLinux(src string) ([]byte, bool) {
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
		panic("Fatal error")
	}
	return normalizeForLinux(data)
}

func removeCRLF(src string) {
	data, change := normalizeFileForLinux(src)
	if !change {
		fmt.Printf("File %s is already normal for Linux", src)
	} else {
		e := ioutil.WriteFile(src, data, 0644)
		if e != nil {
			fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
			panic("Fatal error")
		}
	}
}

func checkAlreadyPresentLine(buf []byte, line string) bool {
	return strings.Index(string(buf), line) >= 0
}

func addNonRepeatedLine(src string, line string) {
	data, _ := normalizeFileForLinux(src)
	byteLine := []byte(line + "\n")
	if checkAlreadyPresentLine(data, line) {
		fmt.Printf("Line %s is already present", line)
	} else {
		data = append(data, byteLine...)
		e := ioutil.WriteFile(src, data, 0644)
		if e != nil {
			fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
			panic("Fatal error")
		}
	}
}

func getFirstNonSpaceByte(data []byte) byte {
	l := len(data)
	for i := 0; i < l; i++ {
		if data[i] > 32 {
			return data[i]
		}
	}
	return 0
}

func getLastNonSpaceByte(data []byte) byte {
	l := len(data)
	for i := l - 1; i >= 0; i-- {
		if data[i] > 32 {
			return data[i]
		}
	}
	return 0
}

func isCurrentFormatJson(data []byte) bool {
	first := getFirstNonSpaceByte(data)
	last := getFirstNonSpaceByte(data)
	return first == '[' && last == ']' || first == '{' && last == ']'
}

func changeExtension(src string, ext string) string {
	lastPos := strings.LastIndex(src, ".")
	if lastPos > 0 {
		slashPos := strings.Index(src[lastPos:], "\\")
		if slashPos < 0 {
			slashPos = strings.Index(src[lastPos:], "/")
		}
		if slashPos < 0 {
			src = src[:lastPos]
		}
	}
	var newName string
	for i := 0; i < 1000000; i++ {
		if i == 0 {
			newName = src + "." + ext
		} else {
			newName = src + strconv.Itoa(i) + "." + ext
		}
		if _, err := os.Stat(newName); os.IsNotExist(err) {
			return newName
		}
	}
	return "tmp." + ext
}
func provideIndentation(res []byte, n int) []byte {
	for i := 0; i < n; i++ {
		res = append(res, ' ')
	}
	return res
}

func provideIndentationBeforeValue(res []byte, state int, indentation int) []byte {
	if state == JSON_WAIT_KEY {
		res = provideIndentation(res, indentation)
	} else if state == JSON_WAIT_ARRAY_VALUE {
		res = provideIndentation(res, indentation)
		res = append(res, '-', 10)
		res = provideIndentation(res, indentation+2)
	} else {
		res = provideIndentation(res, 1)
	}
	return res

}

func provideStringForYaml(res []byte, data []byte) []byte {
	l := len(data)
	if l == 0 {
		res = append(res, '"')
		res = append(res, '"')
		return res
	}
	isNumber := true
	isSimple := true
	for i := 0; i < l; i++ {
		b := data[i]
		if !(b >= '0' && b <= '9' || b == '-' && i == 0) {
			if b != '.' {
				isNumber = false
			}
			if !(b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b == '-') {
				isSimple = false
				if !isNumber {
					break
				}
			}
		}
	}
	if isNumber {
		res = append(res, '\'')
		res = append(res, data...)
		res = append(res, '\'')
		return res
	}
	if isSimple {
		res = append(res, data...)
		return res
	}
	res = append(res, '"')
	for i := 0; i < l; i++ {
		b := data[i]
		if b == '\\' {
			i++
			c := byte(' ')
			if i < l {
				c = data[i]
			}
			if c == '\\' || c == '"' {
				res = append(res, b, c)
			} else {
				res = append(res, c)
			}
		} else {
			res = append(res, b)
		}
	}
	res = append(res, '"')
	return res
}

func provideIndentationAfterValue(state int, res []byte) (int, []byte) {
	switch state {
	case JSON_WAIT_KEY:
		res = append(res, ':')
		state = JSON_WAIT_COLON
	case JSON_WAIT_VALUE, JSON_WAIT_ARRAY_VALUE:
		state = JSON_WAIT_COMMA_OR_CLOSING
	default:
		panic("Unexpected state after value: " + strconv.Itoa(state))
	}
	return state, res
}

func convertToYaml(data []byte, indentation int) (res []byte) {
	if indentation < 1 {
		indentation = 2
	}
	sizePerItem := indentation * 10
	l := len(data)
	newLen := l + sizePerItem
	for i := 0; i < l; i++ {
		if data[i] == '[' || data[i] == '{' {
			newLen += sizePerItem
		}
	}
	res = make([]byte, 0, newLen)
	state := JSON_WAIT_VALUE
	level := -1
	levels := make([]byte, 0, 20)

	for i := 0; i < l; i++ {
		for ; i < l; i++ {
			if data[i] > 32 {
				break
			}
		}
		switch data[i] {
		case '{':
			if state != JSON_WAIT_VALUE && state != JSON_WAIT_ARRAY_VALUE {
				panic("Unexpected { at " + strconv.Itoa(i) + " in " + string(data[i:]))
			}
			state = JSON_WAIT_KEY
			level++
			levels = append(levels, '}')
		case '[':
			if state != JSON_WAIT_VALUE && state != JSON_WAIT_ARRAY_VALUE {
				panic("Unexpected [ at " + strconv.Itoa(i) + " in " + string(data[i:]))
			}
			state = JSON_WAIT_ARRAY_VALUE
			level++
			levels = append(levels, ']')
		case '"':
			if state != JSON_WAIT_VALUE && state != JSON_WAIT_KEY && state != JSON_WAIT_ARRAY_VALUE {
				panic("Unexpected { at " + strconv.Itoa(i) + " in " + string(data[i:]))
			}
			i++
			n := i
			for ; i < l; i++ {
				if data[i] == '\\' {
					i++
				} else if data[i] == '"' {
					break
				}
			}
			res = provideIndentationBeforeValue(res, state, level*indentation)
			res = provideStringForYaml(res, data[n:i])
			state, res = provideIndentationAfterValue(state, res)
		case ',':
			if state != JSON_WAIT_COMMA_OR_CLOSING || level < 0 {
				panic("Unexpected comma at " + strconv.Itoa(i) + " in " + string(data[i:]))
			}
			switch levels[level] {
			case '}':
				state = JSON_WAIT_KEY
			case ']':
				state = JSON_WAIT_ARRAY_VALUE
			}
		case ']', '}':
			if state != JSON_WAIT_COMMA_OR_CLOSING || level < 0 || levels[level] != data[i] {
				panic("Unexpected closing bracket  at " + strconv.Itoa(i) + " in " + string(data[i:]))
			}
			levels = levels[:level]
			level--
			if level < 0 {
				state = JSON_WAIT_END
			} else {
				state = JSON_WAIT_COMMA_OR_CLOSING
			}
		case ':':
			if state != JSON_WAIT_COLON {
				panic("Unexpected colon  at " + strconv.Itoa(i) + " in " + string(data[i:]))
			}
			state = JSON_WAIT_VALUE
		default:
			if state != JSON_WAIT_VALUE && state != JSON_WAIT_ARRAY_VALUE {
				panic("Unexpected { at " + strconv.Itoa(i) + " in " + string(data[i:]))
			}
			isNumber := true
			isWord := true
			p := i
			for ; i < l; i++ {
				b := data[i]
				if !(b >= '0' && b <= '9' || b == '.') {
					isNumber = false
				}
				if !(b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z') {
					isWord = false
				}
			}
			if isWord {
				word := strings.ToLower(string(data[p:i]))
				isWord = word == "true" || word == "false" || word == "null"
			}
			if isWord || isNumber {
				res = append(res, data[p:i]...)
				i--
				state, res = provideIndentationAfterValue(state, res)
			} else {
				panic("Unexpected " + string(data[p:i]) + " at " + strconv.Itoa(i))
			}
		}
	}
	return res
}

func indentCharacter(res []byte, b byte, indentation int) []byte {
	res = provideIndentation(res, indentation)
	res = append(res, b)
	if indentation > 0 {
		res = append(res, 10)
	}
	return res
}

func findSpecialCharacter(data []byte, i int, l int) int {
	for ; i < l; i++ {
		b := data[i]
		if b == ':' || b == '|' || b == '[' || b == '{' || b == 13 || b == 10 {
			break
		}
	}
	return i
}

func findBeforeSpecialCharacter(data []byte, i int, l int) (lastPos int, endChar byte) {
	lastPos = findSpecialCharacter(data, i, l)
	if lastPos == l {
		endChar = 10
	} else {
		endChar = data[lastPos]
	}
	for ; lastPos > 0 && data[lastPos] <= 32; lastPos-- {
	}
        return
}

func processValueJSON(res []byte, data []byte, noEscaping bool, sureString bool) []byte {
	return res
}

func convertToJson(data []byte, indentation int) (res []byte) {
	sizePerItem := indentation * 10
	if sizePerItem < 10 {
		sizePerItem = 10
	}
	l := len(data)
	newLen := l + sizePerItem
	for i := 0; i < l; i++ {
		if data[i] == 10 {
			newLen += sizePerItem
		}
	}
	res = make([]byte, 0, newLen)
	level := 0
	currentIndent := 0
	levels := make([]byte, 0, 20)
	indents := make([]int, 0, 20)
	for i := 0; i < l; i++ {
		n := i
		for ; i < l; i++ {
			if data[i] == 10 || data[i] == 13 {
				n = i + 1
			} else if data[i] > 32 {
				break
			}
		}
		n = i - n
		b := data[i]
		switch b {
		case '-':
		case '[':
			res = indentCharacter(res, '[', level*indentation)
			level++
			levels = append(levels, ']')
			indents = append(indents, currentIndent)
			currentIndent = -1
		case '|':
		case '{':
			res = indentCharacter(res, '{', level*indentation)
			level++
			levels = append(levels, '}')
			indents = append(indents, currentIndent)
			currentIndent = -1
		case '}', ']':
		case '?':
		case ':':
		case '\'', '"':
		case '%':
			//we do not convert tags YAML or TAG
			for ; i < l; i++ {
				if data[i] == 10 || data[i] == 13 {
					break
				}
			}
		case '#':
			//we do not convert comments
			for ; i < l; i++ {
				if data[i] == 10 || data[i] == 13 {
					break
				}
			}
		default:
			p, endChar := findBeforeSpecialCharacter(data, i+1, l)
			res = processValueJSON(res, data[i:p], true, false)
                        fmt.Println(endChar)
		}
	}
        return 
}

func convertYamlToJsonOrBack(src string, indentation int) {
	data, e := ioutil.ReadFile(src)
	if e != nil {
		fmt.Printf("Cannot read file %s: %s\n", src, e.Error())
		panic("Fatal error")
	}
	ext := "json"
	var newData []byte
	if isCurrentFormatJson(data) {
		newData = convertToYaml(data, indentation)
		ext = "yaml"
	} else {
		newData = convertToJson(data, indentation)
	}
	newSrc := changeExtension(src, ext)
	e = ioutil.WriteFile(newSrc, newData, 0644)
	if e != nil {
		fmt.Printf("Cannot write file %s: %s\n", newSrc, e.Error())
		panic("Fatal error")
	}
	fmt.Printf("Converted to %s", newSrc)
}

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println(copyright)
		fmt.Println("kbhelper <podname> <podlist, pods.txt by default> <cmd name, r.cmd by default> <command call by default>")
		fmt.Println("or kbhelper - <filename to convert all cr/lf to lf for linux")
		fmt.Println("or kbhelper + <filename> <line to be added if it is not present yet, everything in Linux style>")
		fmt.Println("or kbhelper % <filename in yaml format, to be converted to json format or back> <integer indentation defaults to 2>")
		return
	}
	podName := os.Args[1]
	podList := "pods.txt"
	podCmd := "r.cmd"
	podCaller := "call"
	if l >= 3 {
		podList = os.Args[2]
	}
	if l >= 4 {
		podCmd = os.Args[3]
	}
	if l >= 5 {
		podCaller = os.Args[4]
	}
	switch podName {
	case "-":
		if l < 3 {
			fmt.Println("File name is not specified")
		} else {
			removeCRLF(podList)
		}
	case "+":
		if l < 4 {
			fmt.Println("File name/line is not specified")
		} else {
			addNonRepeatedLine(podList, podCmd)
		}
	case "%":
		if l < 3 {
			fmt.Println("File name is not specified")
		} else {
			indentation := 2
			if indent, e := strconv.Atoi(podCmd); e == nil && indent >= 0 {
				indentation = indent
			}
			convertYamlToJsonOrBack(podList, indentation)
		}
	default:
		pod, project := findPod(podList, podName)
		fmt.Println("@" + podCaller + " " + podCmd + " " + pod + " " + project)
	}
}
