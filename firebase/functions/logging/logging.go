package logging

import (
	"fmt"
	"strings"
)

//Debug -
func Debug(message string) {
	printLog(message, "debug")
}

//Error -
func Error(err error) {
	printLog(err.Error(), "error")
}

//Info -
func Info(message string) {
	printLog(message, "info")
}

//Warning -
func Warning(message string) {
	printLog(message, "warning")
}

func printLog(message string, severity string) {
	msg := strings.ReplaceAll(message, "\"", "'")
	fmt.Println(fmt.Sprintf("{\"message\": \"%v\", \"severity\": \"%v\"}", msg, severity))
}
