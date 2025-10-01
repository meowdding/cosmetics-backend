package utils

import (
	"encoding/json"
	"fmt"
)

func PrintData(data interface{}) {
	g, _ := json.Marshal(data)
	fmt.Println(string(g))
}

func Log(prefix string, data interface{}) {
	g, _ := json.Marshal(data)
	fmt.Println(prefix, " ", string(g))
}
