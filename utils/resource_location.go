package utils

func IsValidResourceLocationNamespace(string string) bool {
	for _, element := range string {
		if !isValidNamespaceChar(element) {
			return false
		}
	}

	return true
}

func isValidNamespaceChar(char int32) bool {
	return char == '_' || char == '-' || char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '.'
}
