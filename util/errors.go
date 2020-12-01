package util

import (
	"os"
)

// Check checks an error, and if present, logs an error message to the console & exits.
func Check(m string, e error, format ...interface{}) {
	args := append(format, e)
	if e != nil {
		Critical(m+"\n%v", args...)
		os.Exit(1)
	}
}
