package constants

// LocalUser is the local user, used for writing files & determining privilege level.
const LocalUser string = "stellaraf"

// HomeDir is the unformatted path of the LocalUser's home directory.
const HomeDir string = "/home/%s"

// SystemdDir is the unformatted path to the local user's systemd service directory.
const SystemdDir string = HomeDir + "/.config/systemd/user"

// HostnameBase is the Base FQDN of the hostname.
const HostnameBase string = "rmon.orion.cloud"
