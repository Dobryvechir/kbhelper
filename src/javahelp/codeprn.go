// Copyright by Danyil Dobryvechir 2019 (dobrivecher@yahoo.com, ddobryvechir@gmail.com)

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
var cmdLine1="javahelp <filename> <addr> <replace or nothing>";
var cmdLine= cmdLine1 + "\n if replace is specified, the byte at this address is replaced, otherwise <filename>.hex file is created and info from this addr is printed into it. address by default is equal to 0";

func helpAndExit(message string) {
   fmt.Println(copyright)
   fmt.Println(cmdLine)
   panic(message)
}

func ReadHexInput(s string) int {
   n:=len(s)
   res:=0
   for i:=0;i<n;i++ {
      c:=s[i];
      d:=0
      if c>='0' && c<='9' {
          d = c - '0'
      } else if c>='a' && c<='f' {
          d = c - 87
      } else if c>='A' && c<='F' {
          d = c - 55
      } else {
         helpAndExit("Incorrect hex input: "+s);
      }
      res = res * 10 + d
   }
   return res
}

func main() {
    n = len(os.Args)
    if (n<2) {
       helpAndExit("File name is not specified");
    }
    fileName:=os.Args[1]
    data,err:=ioutil.ReadFile(fileName)
    if err!=nil {
         fmt.Printf("Error reading file %s: %v", fileName, err)
    }
    addr := 0;
    if (n>=3) {
        addr = ReadHexInput(os.Args[2]);
    }
    replace:=-1
    if (n>=4) {
         replace = ReadHexInput(os.Args[2]);
    }
    if addr>=len(data) || addr<0 {
         fmt.Printf("Address %x must be less than the length of the file %x", addr, len(data))
         return   
    }
    if replace<-2 || replace>255 {
         fmt.Printf("Replace %x must be within 00 - ff")
         return
    }
    if replace<0 {
         ShowFilePresentation(fileName, addr, data)
    } else {
         ReplaceByteInsideFile(fileName, addr, data, replace)
    }
}

func ReplaceByteInsideFile(fileName string, addr int, data []byte, replace int) {
  data[addr]=byte(replace)
  
}
