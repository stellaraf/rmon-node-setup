package docker

import (
	"os/exec"

	g "github.com/stellaraf/rmon-node-setup/globals"
	util "github.com/stellaraf/rmon-node-setup/util"
)

// CreateGroup creates the docker group so docker can run without root privileges.
func CreateGroup(user string) {
	const group string = "docker"
	exists := false
	memberOf := false
	groups := util.AllGroups()

	for _, g := range groups {
		if g == group {
			exists = true
		}
	}

	userGroups := util.GetUserGroups(g.LocalUser)

	for _, g := range userGroups {
		if g == group {
			memberOf = true
		}
	}

	if exists {
		util.Info("Docker group already exists")
	} else {
		util.NewGroup(group)
	}
	if memberOf {
		util.Info("%s is already a member of docker group", g.LocalUser)
	} else {
		util.UserToGroup(g.LocalUser, group)
	}
}

// EnableStartup enables docker to start on boot.
func EnableStartup() {
	out, err := exec.Command("systemctl", "enable", "docker").CombinedOutput()
	util.Check("Error setting Docker to start on boot:\n%s", err, util.AsString(out))
}
