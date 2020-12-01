package systemd

import (
	"fmt"
	"os"

	g "github.com/stellaraf/rmon-node-setup/globals"
	util "github.com/stellaraf/rmon-node-setup/util"
)

// UserFuncs is a container for systemd --user
type UserFuncs struct {
	CheckService   func(string) bool
	ReloadServices func()
	WriteSystemd   func(string, string, ...interface{})
	EnableService  func(string)
	StartService   func(string)
}

// User is a container for systemd --user
func User() (funcs UserFuncs) {
	funcs = UserFuncs{
		// CheckService checks if a service is active.
		CheckService: func(name string) (result bool) {
			result = false
			util.RunAs(g.LocalUser, func() {
				check, err := util.UserCommand("systemctl", "--user", "is-active", name)
				if err != nil {
					result = false
				}

				outputString := util.AsString(check)
				if outputString == "active" {
					result = true
				}
			})()
			return
		},
		// ReloadServices reloads user systemd services.
		ReloadServices: func() {
			reload, err := util.UserCommand("systemctl", "--user", "daemon-reload")
			util.Check("Error reloading systemd services:\n%s", err, util.AsString(reload))
		},
		// WriteSystemd generates & writes a systemd service file.
		WriteSystemd: func(name, content string, f ...interface{}) {

			filename := fmt.Sprintf(g.SystemdDir+"/%s.service", g.LocalUser, name)

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
			filename := fmt.Sprintf(g.SystemdDir+"/%s.service", g.LocalUser, name)
			serviceName := fmt.Sprintf("%s.service", name)

			active := funcs.CheckService(name)

			if active {
				stop, err := util.UserCommand("systemctl", "--user", "stop", serviceName)
				util.Check("Error stopping %s service:\n%s", err, name, util.AsString(stop))
				util.Info("Stopping %s service...", name)
			}

			if !util.FileExists(filename) {
				util.Check("%s service file is missing (%s)", os.ErrNotExist, name, filename)
			}

			enable, err := util.UserCommand("systemctl", "--user", "enable", serviceName)
			util.Check("Error enabling %s service:\n%s", err, name, util.AsString(enable))

			util.Success("Set %s service to start at login", name)
		},
		// StartService starts a service via systemd.
		StartService: func(name string) {
			util.Info("Starting %s service...", name)

			serviceName := fmt.Sprintf("%s.service", name)

			start, err := util.UserCommand("systemctl", "--user", "restart", serviceName)
			util.Check("Error starting %s service:\n%s", err, name, util.AsString(start))

			active := funcs.CheckService(name)

			if active {
				util.Success("Started %s service", name)
			} else {
				util.Warning("%s service failed to start", name)
			}
		},
	}
	return
}
