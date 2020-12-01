package systemd

import (
	g "github.com/stellaraf/rmon-node-setup/globals"
	util "github.com/stellaraf/rmon-node-setup/util"
)

// AutoSSH creates & sets up AutoSSH as a systemd service.
func AutoSSH(nodeID, tunnelServer string) {
	service := `[Unit]
Description=AutoSSH
Wants=network-online.target
After=network-online.target
StartLimitIntervalSec=0

[Service]
ExecStart=/usr/bin/autossh -M 0 -N \
	-o "ServerAliveInterval 15" \
	-o "ServerAliveCountMax 3" \
	-o "ConnectTimeout 10" \
	-o "ExitOnForwardFailure yes" \
	-i /home/%s/.ssh/id_rsa \
	rmontunnel@%s \
	-R 100%s:localhost:22
Restart=always
RestartSec=10

[Install]
WantedBy=default.target
`
	util.RunAs(g.LocalUser, func() {
		u := User()
		u.WriteSystemd("autossh", service, g.LocalUser, tunnelServer, nodeID)
		u.ReloadServices()
		u.EnableService("autossh")
		u.StartService("autossh")
	})()
}
