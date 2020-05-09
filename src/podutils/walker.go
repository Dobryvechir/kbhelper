package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func walkingAddCrLf(path string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() {
		return addCRLF(path)
	}
	return nil
}

func walkingRemoveCrLf(path string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() {
		return removeCRLF(path)
	}
	return nil
}

func walkCrLf(path string, handler filepath.WalkFunc) {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Path %s is incorrect: %v", path, err)
		return
	}
	if info.IsDir() {
		err = filepath.Walk(path, handler)
	} else {
		err = handler(path, info, nil)
	}
	if err != nil {
		fmt.Printf("Problem for %s occurred: %v", path, err)
		return
	}
}

func walkAddCrLf(path string) {
	walkCrLf(path, walkingAddCrLf)
}

func walkRemoveCrLf(path string) {
	walkCrLf(path, walkingRemoveCrLf)
}
