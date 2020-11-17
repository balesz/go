package logging

import (
	"fmt"
)

//Critical -
func Critical(err error) {
	printLog(err.Error(), "critical")
}

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

func printLog(message string, severity string) {
	msg := fmt.Sprintf("{\"message\": \"%v\", \"severity\": \"%v\"}", message, severity)
	fmt.Println(msg)
}
