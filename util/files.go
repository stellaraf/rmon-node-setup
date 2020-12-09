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

	srcInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		Check("Source file %s does not exist", err, src)
	}

	source, err := os.Open(src)
	Check("Unable to open source file %s", err, src)
	defer source.Close()

	if FileExists(dst) {
		Info("Destination %s already exists and will be overwritten", dst)
	}

	destination, err := os.Create(dst)
	Check("Unable to create destination file %s while copying from %s", err, dst, src)
	defer destination.Close()

	_, err = io.Copy(destination, source)
	Check("Error copying %s to %s", err, src, dst)

	err = os.Chmod(dst, srcInfo.Mode())
	Check("Error setting permissions on %s to %d", err, dst, srcInfo.Mode())

	err = destination.Sync()
	Check("Error saving contents of copied file %s", err, dst)
	return
}

// ScaffoldUser creates the required directory structure for the non-root user.
func ScaffoldUser() {
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
		err := os.Mkdir(ssh, 0700)
		Check("Error creating SSH directory at %s", err, ssh)
	}
	return
}

// ScaffoldRoot creates the required system directory structure.
func ScaffoldRoot() {
	dirs := []string{"/etc/docker/compose"}
	for _, d := range dirs {
		if !FileExists(d) {
			err := os.MkdirAll(d, 0777)
			Check("Error creating directory %s:\n", err, d)
			if FileExists(d) {
				Success("Created directory %s", d)
			} else {
				Warning("Unable to create directory %s", d)
			}
		} else {
			Info("Directory %s already exists", d)
		}
	}
	return
}
