package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	docker "github.com/stellaraf/rmon-node-setup/docker"
	g "github.com/stellaraf/rmon-node-setup/globals"
	systemd "github.com/stellaraf/rmon-node-setup/systemd"
	util "github.com/stellaraf/rmon-node-setup/util"

	color "github.com/fatih/color"
)

// GetNodeID prompts the user for the 2 digit node ID.
func GetNodeID() (nodeID string) {
	fmt.Print("Node ID (2 digit number): ")
	fmt.Scanf("%s", &nodeID)
	matched, err := regexp.MatchString(`^[0-9]{1,2}$`, nodeID)
	util.Check("Error checking Node ID format: ", err)
	if !matched {
		util.Warning("Invalid Node ID. Node ID must be a 2 digit number.")
		return GetNodeID()
	}
	return fmt.Sprintf("%02s", nodeID)
}

// GetTunnelServer prompts the user for the remote SSH tunnel server.
func GetTunnelServer() (tunnelServer string) {
	fmt.Print("SSH Tunnel Server (FQDN): ")
	fmt.Scanf("%s", &tunnelServer)
	parts := strings.Split(tunnelServer, ".")
	if len(parts) < 3 {
		util.Warning("Invalid SSH Tunnel Server. Must be an FQDN. Got: %s", tunnelServer)
		return GetTunnelServer()
	}
	return
}

// GetAPIKey prompts the user for the AppNeta API Key.
func GetAPIKey() (apiKey string) {
	fmt.Print("Enter the AppNeta API Key from IT Glue: ")
	fmt.Scanf("%s", &apiKey)

	if len(apiKey) != 32 {
		util.Warning("Invalid API Key - expected 32 characters. Got %d characters", len(apiKey))
		return GetAPIKey()
	}
	return apiKey
}

func main() {

	isRoot := util.IsRoot()

	if !isRoot {
		util.Critical("Setup must be run with root privileges. Try again with sudo.")
		os.Exit(1)
	}

	status := 1

	util.AddToSudoers(g.LocalUser)

	blue := color.New(color.Bold, color.FgBlue).SprintFunc()
	yellow := color.New(color.Bold, color.FgYellow).SprintFunc()

	color.New(color.FgMagenta, color.Bold).Print("\nOrion RMON Raspberry Pi Setup\n\n")
	color.New(color.FgWhite, color.Bold).Println("You'll need:")

	fmt.Printf(`
  - %s of the unit, a unique 2 digit number between 1-99.
  - %s of the remote SSH tunnel server.

`, blue("ID number"), yellow("FQDN"))

	nodeID := GetNodeID()
	tunnelServer := GetTunnelServer()
	hostname := fmt.Sprintf("rpi%s.%s", nodeID, g.HostnameBase)

	if len(hostname) > 255 {
		util.Critical("Hostname must be no more than 255 characters long. Hostname %s is %s characters long", hostname, len(hostname))
		os.Exit(1)
	}

	hostnameDashes := strings.ReplaceAll(hostname, ".", "-")
	outDir := fmt.Sprintf(g.HomeDir, g.LocalUser)

	fmt.Println()
	util.SetHostname(hostname)
	util.SetTimezone()
	util.Dependencies()

	docker.Install()
	docker.CreateGroup(g.LocalUser)
	docker.EnableStartup()

	apiKey := GetAPIKey()
	docker.InstallCompose()
	docker.GetCompose(apiKey, hostnameDashes, outDir)
	docker.SetupCompose(hostnameDashes, outDir)
	systemd.Root().StopService("appneta-cmp")
	docker.Scaffold(hostnameDashes)
	systemd.DockerCompose()

	util.RunAs(g.LocalUser, func() {
		util.Scaffold()
		util.CheckSSHKeys()

		systemd.AutoSSH(nodeID, tunnelServer)

	})()
	status = 0

	util.Success("Setup complete!")
	os.Exit(status)
}
