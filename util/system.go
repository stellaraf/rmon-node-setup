package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	g "github.com/stellaraf/rmon-node-setup/globals"
)

// Dependencies installs system dependencies.
func Dependencies() {
	deps := []string{"autossh", "libffi-dev", "libssl-dev", "python3", "python3-pip"}
	cmdArgs := []string{"install", "-y"}
	args := append(cmdArgs, deps...)

	cmd := exec.Command("apt-get", args...)
	Info("Installing dependencies...")

	_, err := cmd.CombinedOutput()
	Check("Error installing dependencies: ", err)

	styledDeps := []string{}
	for _, d := range deps {
		df := fmt.Sprintf("\n  - %s", d)
		styledDeps = append(styledDeps, df)
	}

	Success("Installed Dependencies:%s", strings.Join(styledDeps, ""))
}

// SetHostname sets the hostname of this node.
func SetHostname(hostname string) {
	cmd := exec.Command("/usr/bin/hostnamectl", "set-hostname", hostname)

	err := cmd.Run()
	Check("Error setting hostname: ", err)

	Success("Set hostname to %s", hostname)
}

// SetTimezone sets the timezone of this node.
func SetTimezone() {
	timezone := "Etc/UTC"
	cmd := exec.Command("/usr/bin/timedatectl", "set-timezone", timezone)

	err := cmd.Run()
	Check("Error setting timezone: ", err)

	Success("Set timezone to %s", timezone)
}

// CheckSSHKeys ensures SSH keys exist and have the correct permissions.
func CheckSSHKeys() {
	privkey := fmt.Sprintf(g.HomeDir, g.LocalUser) + "/.ssh/id_rsa"
	pubkey := privkey + ".pub"

	if !FileExists(pubkey) {
		Check("Public key missing at %s", os.ErrNotExist, pubkey)
	}

	if !FileExists(privkey) {
		Check("Private key missing at %s", os.ErrNotExist, pubkey)
	}

	stat, err := os.Stat(privkey)
	Check("Error reading SSH private key: ", err)

	mode := stat.Mode()
	if mode != 0600 {
		os.Chmod(privkey, 0600)
		Info("Set permissions for %s to 0600 (-rw-------)", privkey)
	}
}

// IsInstalled determines if a package is installed by checking to see if it is in $PATH
func IsInstalled(pkg string) (i bool) {
	cmd := exec.Command("which", pkg)
	err := cmd.Run()
	if err == nil {
		i = true
	}
	return
}
