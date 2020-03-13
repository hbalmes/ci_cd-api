package utils

import "encoding/json"

func Stringify(a string) *string {
	return &a
}

func GetBytes(e interface{}) []byte {
	bytes, _ := json.Marshal(e)
	return bytes
}
