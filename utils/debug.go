package utils

import (
	"encoding/json"
	"fmt"
)

type LogData struct {
	Message string
	Data    interface{}
}

func (data LogData) Log() {
	PrintData(&data)
}

func PrintData(data interface{}) {
	g, _ := json.Marshal(data)
	fmt.Println(string(g))
}
