package logger

import (
	"log"
)

// ANSI color codes
const (
	red    = "\033[31m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

func Info(msg string) {
	log.Printf("%sINFO%s : %s", blue, reset, msg)
}

func Warn(msg string) {
	log.Printf("%sWARN%s : %s", red, reset, msg)
}

func Error(msg string) {
	log.Printf("%sERROR%s : %s", red, reset, msg)
}

func Debug(msg string) {
	log.Printf("%sDEBUG%s : %s", yellow, reset, msg)
}