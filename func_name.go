package server

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

var funcNameRegExp = regexp.MustCompile("([a-zA-Z0-9_]+)(.*)")

func funcName(i interface{}) string {
	raw := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	splitted := strings.Split(raw, ".")
	last := splitted[len(splitted)-1]
	return funcNameRegExp.ReplaceAllString(last, "$1")
}
