package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type NullRayConfig struct {
	R   []float64 `json:"r"`
	N   []float64 `json:"n"`
	D   []float64 `json:"d"`
	H   []float64 `json:"h"`
	Tgs []float64 `json:"tgs"`
}

func Calculate(c *NullRayConfig) {
	n := len(c.Tgs) - 1
	for k := 1; k < n; k++ {
		c.Tgs[k+1] = (c.N[k]/c.N[k+1])*c.Tgs[k] + ((c.N[k+1]-c.N[k])/c.N[k+1])*(c.H[k]/c.R[k])
		if k != n-1 {
			c.H[k+1] = c.H[k] - c.D[k]*c.Tgs[k+1]
		}
	}
}

func ReadFromFile(fileName string) (*NullRayConfig, error) {
	c := &NullRayConfig{}
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(dat), c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Initialize(c *NullRayConfig) {
	n := len(c.N)
	c.H = make([]float64, n)
	c.Tgs = make([]float64, n)
	c.H[0] = 0
	c.H[1] = 10
	c.Tgs[0] = 0
	c.Tgs[1] = 0
}

func PrintResults(c *NullRayConfig) {
	n := len(c.Tgs)
	for i := 2; i < n; i++ {
		fmt.Printf("tgs%d = %.15f\n", i, c.Tgs[i])
		if i != n-1 {
			fmt.Printf("h%d = %.15f\n", i, c.H[i])
		}
	}
}

func main() {
	fileName := "./indata.json"
	if len(os.Args) > 1 {
		fileName = os.Args[1]
	}
	c, err := ReadFromFile(fileName)
	if err != nil {
		fmt.Printf("Error reading file %s %v", fileName, err)
		return
	}
	Initialize(c)
	Calculate(c)
	PrintResults(c)
	fmt.Println("Done")
}
