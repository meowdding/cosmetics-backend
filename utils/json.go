package utils

import (
	"encoding/json"
	"strings"
)

func IsJson(str string) bool {
	if !(strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) {
		return false
	}
	if !json.Valid([]byte(str)) {
		return false
	}
	var rawMessage json.RawMessage
	return json.Unmarshal([]byte(str), &rawMessage) == nil
}
