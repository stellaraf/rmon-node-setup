package util

import (
	"fmt"
	"io"
	"os"

	g "github.com/stellaraf/rmon-node-setup/globals"
)

// FileExists checks if a file exists.
func FileExists(f string) (e bool) {
	if _, exists := os.Stat(f); os.IsNotExist(exists) {
		e = false
	} else {
		e = true
	}
	return
}

// CopyFile copies a file from src to dst and OVERWRITES.
func CopyFile(src string, dst string) {

	if !FileExists(src) {
		Check("Source file %s does not exist", os.ErrNotExist, src)
	}

	source, err := os.Open(src)
	Check("Unable to open source file %s", err, src)
	defer source.Close()

	if FileExists(dst) {
		Info("Destination %s already exists. Deleting...", dst)
		err := os.Remove(dst)
		Check("Unable to delete destination %s", err, dst)
	}
	destination, err := os.Create(dst)
	Check("Unable to create destination file %s", err, dst)
	defer destination.Close()

	_, err = io.Copy(destination, source)
	Check("Error copying %s to %s", err, src, dst)
}

// Scaffold creates the required directory structure.
func Scaffold() {
	path := fmt.Sprintf(g.SystemdDir, g.LocalUser)
	ssh := fmt.Sprintf(g.HomeDir, g.LocalUser) + "/.ssh"
	if !FileExists(path) {
		err := os.MkdirAll(path, 0755)
		Check("Error while creating directories: ", err)
		Success("Created directory '%s'", path)
	} else {
		Info("Directory %s already exists", path)
	}

	if !FileExists(ssh) {
		err := os.Mkdir(ssh, 700)
		Check("Error creating SSH directory at %s", err, ssh)
	}

	return
}
