package utils

func IsAuthenticated(data string) bool {
	return config.ApiToken == data
}
