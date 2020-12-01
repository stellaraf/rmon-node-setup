package util

import (
	"fmt"
	"strings"

	color "github.com/fatih/color"
)

const reset = iota

func unsetChar() string {
	return fmt.Sprintf("%s[%dm", "\x1b", reset)
}

// Logger is a base logger utility to conveniently format console output.
func Logger(t color.Attribute, m string, f ...interface{}) {
	c := color.New(t)
	msg := c.Sprint(m)

	if f != nil {
		var args []interface{}

		bold := color.New(color.Bold).SprintFunc()

		for _, s := range f {

			/** fatih/color adds a reset code to the end of any Sprintf results, so in order to
			 *  keep the color from resetting after the format arguments, re-add the color-only
			 *  ansi code to the end of the format argument. Note: this currently only works
			 *  with foreground colors.
			 */
			e := bold(s)
			if strings.Contains(e, "\x1b[0m") {
				e = e + fmt.Sprintf("\x1b[%dm", t)
			}
			args = append(args, e)
		}
		c.Printf(msg+"\n", args...)

	} else {
		fmt.Println(msg)
	}
}

// Debug logs a blue message to the console.
func Debug(m string, f ...interface{}) {
	m = "[DEBUG] " + m
	Logger(color.FgBlue, m, f...)
}

// Success logs a green message to the console.
func Success(m string, f ...interface{}) {
	Logger(color.FgGreen, m, f...)
}

// Info logs a blue message to the console.
func Info(m string, f ...interface{}) {
	Logger(color.FgBlue, m, f...)
}

// Warning logs a yellow message to the console.
func Warning(m string, f ...interface{}) {
	Logger(color.FgYellow, m, f...)
}

// Critical logs a red message to the console.
func Critical(m string, f ...interface{}) {
	Logger(color.FgRed, m, f...)
}
