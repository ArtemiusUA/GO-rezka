package helpers

import (
	"os"
	"reflect"
	"strings"
)

func IsDebug() bool {
	debug := strings.ToLower(os.Getenv("DEBUG"))
	switch debug {
	case
		"1",
		"true",
		"yes":
		return true
	}
	return false
}

func ReverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
