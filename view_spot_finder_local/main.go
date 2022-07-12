package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"golang.org/x/exp/slices"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

type File struct {
	Nodes    []Node    `json:"nodes"`
	Elements []Element `json:"elements"`
	Values   []Value   `json:"values"`
}

type Node struct {
	Id int     `json:"id"`
	X  float32 `json:"x"`
	Y  float32 `json:"y"`
}

type Element struct {
	Id    int   `json:"id"`
	Nodes []int `json:"nodes"`
}

type Value struct {
	Element_id int     `json:"element_id"`
	Value      float32 `json:"value"`
}

func init() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	InfoLogger.Print("Started")
	start := time.Now()
	do(os.Args[1:])
	end := time.Now()
	InfoLogger.Printf("Finished - took %v", end.Sub(start))
}

func do(args []string) {

	if len(args) < 2 {
		ErrorLogger.Println("Not enough parameters")
		os.Exit(1)
	}

	nSpots, err := strconv.Atoi(args[1])
	if err != nil {
		ErrorLogger.Printf("Could not convert N to int because of input %v", args[1])
		os.Exit(1)
	}

	file, err := readFile(args[0])
	if err != nil {
		ErrorLogger.Printf("Could not read file %v", args[0])
		os.Exit(1)
	}

	sort.Slice(file.Values, func(i, j int) bool { return file.Values[i].Value > file.Values[j].Value })

	var output []Value
	var loop bool = true
	var valueIndex int = 0
	var currentValue Value

	for loop && valueIndex < len(file.Values) {
		currentValue = file.Values[valueIndex]
		elementIdx := slices.IndexFunc(file.Elements, func(e Element) bool { return e.Id == currentValue.Element_id })
		ele := file.Elements[elementIdx]

		var neighbours []Element

		for _, elementNode := range ele.Nodes {
			idx := slices.IndexFunc(file.Elements, func(e Element) bool { return contains(e.Nodes, elementNode) })
			neighbours = append(neighbours, file.Elements[idx])
		}

		var isViewSpot bool = true
		for _, neighbourElement := range neighbours {
			foundValueIdx := slices.IndexFunc(file.Values, func(v Value) bool { return v.Element_id == neighbourElement.Id })
			foundValue := file.Values[foundValueIdx]
			if foundValue.Value > currentValue.Value {
				isViewSpot = false
			}
		}

		if isViewSpot {
			output = append(output, currentValue)
		}

		if len(output) == nSpots {
			loop = false
		}
		valueIndex++
	}

	jsonResult, err := json.Marshal(output)
	if err != nil {
		ErrorLogger.Printf("Could not convert output to json %v", err)
	}
	fmt.Println(string(jsonResult))
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func readFile(path string) (File, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return File{}, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	var temp File

	json.Unmarshal(byteValue, &temp)

	return temp, nil
}
