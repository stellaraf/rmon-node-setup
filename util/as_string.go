package util

import "strings"

// AsString converts bytes from CLI output to a string & trims the trailing newline.
func AsString(b []byte) string {
	return strings.TrimSuffix(string(b), "\n")
}
