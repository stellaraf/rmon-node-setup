package systemd

import (
	"fmt"
	"os"
	"os/exec"

	util "github.com/stellaraf/rmon-node-setup/util"
)

// RootFuncs is a container for systemd commands run as root.
type RootFuncs struct {
	CheckService   func(string) bool
	ReloadServices func()
	WriteSystemd   func(string, string, ...interface{})
	EnableService  func(string)
	StartService   func(string)
	StopService    func(string)
}

// Root is a container for systemd commands run as root.
func Root() (funcs RootFuncs) {
	funcs = RootFuncs{
		// CheckService checks if a service is active.
		CheckService: func(name string) (result bool) {
			result = false
			check, err := exec.Command("systemctl", "is-active", name).CombinedOutput()
			if err != nil {
				result = false
			}

			outputString := util.AsString(check)
			if outputString == "active" {
				result = true
			}
			return
		},
		// ReloadServices reloads systemd services.
		ReloadServices: func() {
			reload, err := exec.Command("systemctl", "daemon-reload").CombinedOutput()
			util.Check("Error reloading systemd services:\n%s", err, util.AsString(reload))
		},
		// WriteSystemd generates & writes a systemd service file.
		WriteSystemd: func(name, content string, f ...interface{}) {

			filename := fmt.Sprintf("/etc/systemd/system/%s.service", name)

			formatted := fmt.Sprintf(content, f...)

			file, err := os.Create(filename)
			util.Check("Error creating %s service file: ", err, name)

			_, writeErr := file.WriteString(formatted)
			util.Check("Error writing %s service file: ", writeErr, name)

			defer file.Close()

			util.Success("Wrote %s service to %s", name, filename)
			return
		},
		EnableService: func(name string) {
			filename := fmt.Sprintf("/etc/systemd/system/%s.service", name)
			serviceName := fmt.Sprintf("%s.service", name)

			active := funcs.CheckService(name)

			if active {
				stop, err := exec.Command("systemctl", "stop", serviceName).CombinedOutput()
				util.Check("Error stopping %s service:\n%s", err, name, util.AsString(stop))
				util.Info("Stopping %s service...", name)
			}

			if !util.FileExists(filename) {
				util.Check("%s service file is missing (%s)", os.ErrNotExist, name, filename)
			}

			enable, err := exec.Command("systemctl", "enable", serviceName).CombinedOutput()
			util.Check("Error enabling %s service:\n%s", err, name, util.AsString(enable))

			util.Success("Set %s service to start at login", name)
		},
		// StartService starts a service via systemd.
		StartService: func(name string) {
			util.Info("Starting %s service...", name)

			serviceName := fmt.Sprintf("%s.service", name)

			start, err := exec.Command("systemctl", "restart", serviceName).CombinedOutput()
			util.Check("Error starting %s service:\n%s", err, name, util.AsString(start))

			active := funcs.CheckService(name)

			if active {
				util.Success("Started %s service", name)
			} else {
				util.Warning("%s service failed to start", name)
			}
		},
		StopService: func(name string) {
			filename := fmt.Sprintf("/etc/systemd/system/%s.service", name)
			serviceName := fmt.Sprintf("%s.service", name)

			if util.FileExists(filename) {
				active := funcs.CheckService(name)

				if active {
					stop, err := exec.Command("systemctl", "stop", serviceName).CombinedOutput()
					util.Check("Error stopping %s service:\n%s", err, name, util.AsString(stop))
				}

				active = funcs.CheckService(name)
				if !active {
					util.Success("Stopped %s service", name)
				}
			}
		},
	}
	return
}
