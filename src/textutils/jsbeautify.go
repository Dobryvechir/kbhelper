// Copyright by Volodymyr Dobryvechir 2019 (dobrivecher@yahoo.com, vdobryvechir@gmail.com)

package main

import (
	"fmt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
	"strconv"
	"strings"
)

var copyright = "Copyright by Volodymyr Dobryvechir 2019"

func findEndOfComment(buf []byte, pos int, kind byte) int {
	n := len(buf)
	switch kind {
	case '/':
		for ; pos < n && buf[pos] != 13 && buf[pos] != 10; pos++ {
		}
	case '*':
		for ; pos < n && !(buf[pos] == '*' && pos+1 < n && buf[pos+1] == '/'); pos++ {
		}
		if pos < n {
			pos += 2
		}
	}
	return pos
}

func nextNonSpace(buf []byte, pos int) (byte, int) {
	n := len(buf)
	for ; pos < n; pos++ {
		if buf[pos] > ' ' {
			return buf[pos], pos
		}
	}
	return ' ', n
}

func prevNonSpace(buf []byte, pos int) (byte, int) {
	for ; pos >= 0; pos-- {
		if buf[pos] > ' ' {
			return buf[pos], pos
		}
	}
	return ' ', -1
}

func isDigitLetter(b byte) bool {
	return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

func prevWord(buf []byte, pos int) (string, int) {
	for ; pos >= 0; pos-- {
		if buf[pos] > ' ' {
			if !isDigitLetter(buf[pos]) {
				return string(buf[pos : pos+1]), pos
			}
			begin := pos
			for ; begin > 0 && isDigitLetter(buf[begin-1]); begin-- {
			}
			return string(buf[begin : pos+1]), pos
		}
	}
	return "", -1
}

func beautify(buf []byte, endings []byte, indent int, excludeComments bool) (o []byte, err error) {
	n := len(buf)
	o = make([]byte, 0, n+(n>>4))
	maxLevel := 24
	k := maxLevel * indent
	indentation := make([]byte, k)
	for i := 0; i < k; i++ {
		indentation[i] = ' '
	}
	level := 0
	for i := 0; i < n; i++ {
		for ; i < n && buf[i] <= 32; i++ {
		}
		if i == n {
			break
		}
		b := buf[i]
		if b == '}' {
			if level > 0 {
				level--
			} else {
				fmt.Printf("Unclosed } at " + strconv.Itoa(i))
			}
			if i != 0 {
				o = append(o, endings...)
			}
			o = append(o, indentation[:level*indent]...)
			o = append(o, b)
			nxtChar, nxtPos := nextNonSpace(buf, i+1)
			for nxtChar == ';' || nxtChar == ',' || nxtChar == ')' {
				o = append(o, nxtChar)
				i = nxtPos
				nxtChar, nxtPos = nextNonSpace(buf, i+1)
			}
			continue
		}
		if i != 0 {
			o = append(o, endings...)
		}
		o = append(o, indentation[:level*indent]...)
		prev := i
	L:
		for ; i < n && buf[i] != 10 && buf[i] != 13; i++ {
			b = buf[i]
			switch b {
			case '"', '\'', '`':
				for i++; i < n && buf[i] != b; i++ {
					if buf[i] == '\\' {
						i++
					}
				}
			case ';':
				i++
				break L
			case '{':
				if level >= maxLevel {
					maxLevel <<= 1
					indentation = append(indentation, indentation...)
				}
				level++
				i++
				break L
			case '}':
				break L
			case '/':
				if i+1 < n {
					b = buf[i+1]
					if b == '*' || b == '/' {
						p := findEndOfComment(buf, i+2, b)
						if excludeComments {
							o = append(o, buf[prev:i]...)
							prev = p
							i = p
							if b == '/' {
								break L
							}
						} else {
							i = p - 1
						}
					} else {
						var lastPos int
						b, lastPos = prevNonSpace(buf, i-1)
						retCase := b == 'n'
						if retCase {
							s, _ := prevWord(buf, lastPos)
							if s != "return" {
								retCase = false
							}
						}
						if retCase || !(b == ')' || b == '_' || b >= '0' && b <= '9' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z') {
							for i++; i < n && buf[i] != '/'; i++ {
								if buf[i] == '\\' {
									i++
								}
							}
						}
					}
				}
			}
		}
		if prev < i {
			o = append(o, buf[prev:i]...)
		}
		i--
	}
	return
}

func compressScript(buf []byte, excludeComments bool) (o []byte, err error) {
	n := len(buf)
	o = make([]byte, 0, n)
	for i := 0; i < n; i++ {
		for ; i < n && buf[i] <= 32; i++ {
		}
		if i == n {
			break
		}
		b := buf[i]
		prev := i
	L:
		for ; i < n && buf[i] != 10 && buf[i] != 13; i++ {
			b = buf[i]
			switch b {
			case '"', '\'', '`':
				for i++; i < n && buf[i] != b; i++ {
					if buf[i] == '\\' {
						i++
					}
				}
			case ';', ',':
				i++
				break L
			case '{', '}', '[', ']', '(', ')':
				if i == prev || buf[i-1] > ' ' {
					i++
					break L
				}
				pos := i
				i--
				for ; i > prev && buf[i-1] <= ' '; i-- {
				}
				o = append(o, buf[prev:i]...)
				prev = pos
				i = pos
				break L
			case '/':
				if i+1 < n {
					b = buf[i+1]
					if b == '*' || b == '/' {
						p := findEndOfComment(buf, i+2, b)
						if excludeComments {
							o = append(o, buf[prev:i]...)
							prev = p
							i = p
							if b == '/' {
								break L
							}
						} else {
							i = p - 1
						}
					}
				}
			}
		}
		if prev < i {
			o = append(o, buf[prev:i]...)
		}
		i--
	}
	return
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 2 {
		fmt.Printf(copyright)
		fmt.Printf("\nCommand line: jsbeautify inputJsFileName outputJsFileName <options:[P][L][Inumber][C]>")
		fmt.Printf("Options:P-compress L-Linux endings,C-exclude comments(default for beautifying),c - keep comments(default for compress),I set indent(default is 4)")
		return
	}
	inFile := args[0]
	outFile := args[1]
	options := ""
	if l > 2 {
		options = args[2]
	}
	indent := 4
	posIndent := strings.Index(options, "I")
	if posIndent >= 0 {
		posIndent++
		if posIndent >= len(options) || !(options[posIndent] >= '0' && options[posIndent] <= '9') {
			panic("Option I must be followed by the number, which is indentation, for example I4")
		}
		indent = int(options[posIndent]) - 48
		posIndent++
		if posIndent < len(options) && options[posIndent] >= '0' && options[posIndent] <= '9' {
			indent = indent*10 + (int(options[posIndent]) - 48)
		}
	}
	excludeComments := strings.Index(options, "C") >= 0
	compress := strings.Index(options, "P") >= 0
	endings := []byte{13, 10}
	if strings.Index(options, "L") >= 0 {
		endings = []byte{10}
	}
	buf, err := ioutil.ReadFile(inFile)
	if err != nil {
		panic("Error reading file " + inFile + ": " + err.Error())
	}
	var newBuf []byte
	if compress {
		excludeComments = strings.Index(options, "c") < 0
		newBuf, err = compressScript(buf, excludeComments)
	} else {
		newBuf, err = beautify(buf, endings, indent, excludeComments)
	}
	if err != nil {
		panic("Error in script: " + inFile + ": " + err.Error())
	}
	err2 := ioutil.WriteFile(outFile, newBuf, 0644)
	if err2 != nil {
		panic("Error writing file " + outFile + ": " + err2.Error())
	}
	fmt.Printf("Successfully converted and written to %s", outFile)
}
