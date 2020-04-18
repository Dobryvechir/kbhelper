/***********************************************************************
DvServer
Copyright 2018 - 2020 by Danyil Dobryvechir (dobrivecher@yahoo.com ddobryvechir@gmail.com)
************************************************************************/

package main

import (
        "fmt"
	"github.com/Dobryvechir/dvserver/src/dvcrypt"
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"io/ioutil"
)

var copyright = "Copyright by Danyil Dobryvechir 2019"

func encrypt(src string, key string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	res, err1 := dvcrypt.EncryptString(key, string(data), true)
	if err1 != nil {
		return err1
	}
	return ioutil.WriteFile(dst, []byte(res), 0644)
}

func decrypt(src string, key string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	res, err1 := dvcrypt.DecryptString(key, string(data), true)
	if err1 != nil {
		return err1
	}
	return ioutil.WriteFile(dst, []byte(res), 0644)
}

func sign(src string, key string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	res, err1 := dvcrypt.DecryptString(key, string(data), true)
	if err1 != nil {
		return err1
	}
	return ioutil.WriteFile(dst, []byte(res), 0644)
}

func verify(src string, key string, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	res, err1 := dvcrypt.DecryptString(key, string(data), true)
	if err1 != nil {
		return err1
	}
	return ioutil.WriteFile(dst, []byte(res), 0644)
}

func main() {
	args := dvparser.InitAndReadCommandLine()
	l := len(args)
	if l < 4 {
		fmt.Println(copyright)
		fmt.Println("Folder for public key is determined by DVSERVER_DVCRYPT_KEY_FOLDER")
		fmt.Println("dvencrypt [decrypt | encrypt | sign | verify] <name of file to be encrypted> <name of public key> <file name to be saved>")
		return
	}
	var err error
	c := args[0][0]
	if c >= 'a' {
		c -= 32
	}
	switch c {
	case 'D':
		err = decrypt(args[1], args[2], args[3])
	case 'E':
		err = encrypt(args[1], args[2], args[3])
	case 'S':
		err = sign(args[1], args[2], args[3])
	case 'V':
		err = verify(args[1], args[2], args[3])
	}

	if err == nil {
		fmt.Printf("Successfully %s in %s", args[0], args[2])
	} else {
		fmt.Printf("Error: %s", err.Error())
	}
}
