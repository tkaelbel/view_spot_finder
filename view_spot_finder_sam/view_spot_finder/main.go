package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/exp/slices"
)

type File struct {
	Nodes    []Node    `json:"nodes"`
	Elements []Element `json:"elements"`
	Values   []Value   `json:"values"`
}

type Node struct {
	Id int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

type Element struct {
	Id    int   `json:"id"`
	Nodes []int `json:"nodes"`
}

type Value struct {
	Element_id int     `json:"element_id"`
	Value      float64 `json:"value"`
}

type RequestBody struct {
	Input  File `json:"file"`
	NSpots int  `json:"nSpots"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	requestBody := RequestBody{}
	err := json.Unmarshal([]byte(request.Body), &requestBody)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Could not start because json could not be unmarshalled %v", err),
			StatusCode: 400,
		}, nil
	}

	if requestBody.NSpots <= 0 {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintln("Could not start because NSpots is less or equal 0"),
			StatusCode: 400,
		}, nil
	}

	if len(requestBody.Input.Nodes) == 0 || len(requestBody.Input.Elements) == 0 || len(requestBody.Input.Values) == 0 {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintln("Could not start because either nodes, elements or values is empty"),
			StatusCode: 400,
		}, nil
	}

	result := do(requestBody.Input.Values, requestBody.Input.Elements, requestBody.Input.Nodes, requestBody.NSpots)

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintln("Could not convert result to json"),
			StatusCode: 400,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprint(string(jsonResult)),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}

func do(values []Value, elements []Element, nodes []Node, nspots int) []Value {

	sort.Slice(values, func(i, j int) bool { return values[i].Value > values[j].Value })

	var output []Value
	// var lastElement element = element{Id: -1}
	var loop bool = true
	var valueIndex int = 0
	var currentValue Value

	for loop && valueIndex < len(values) {
		currentValue = values[valueIndex]
		elementIdx := slices.IndexFunc(elements, func(e Element) bool { return e.Id == currentValue.Element_id })
		ele := elements[elementIdx]

		var neighbours []Element

		for _, elementNode := range ele.Nodes {
			idx := slices.IndexFunc(elements, func(e Element) bool { return contains(e.Nodes, elementNode) })
			neighbours = append(neighbours, elements[idx])
		}

		var isViewSpot bool = true
		for _, neighbourElement := range neighbours {
			foundValueIdx := slices.IndexFunc(values, func(v Value) bool { return v.Element_id == neighbourElement.Id })
			foundValue := values[foundValueIdx]
			if foundValue.Value > currentValue.Value {
				isViewSpot = false
			}
		}

		if isViewSpot == true {
			output = append(output, currentValue)
		}

		if len(output) == nspots {
			loop = false
		}
		valueIndex++
	}

	return output
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
