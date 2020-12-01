package docker

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	g "github.com/stellaraf/rmon-node-setup/globals"
	util "github.com/stellaraf/rmon-node-setup/util"
)

func getArch() (arch string) {
	cmd := exec.Command("dpkg", "--print-architecture")
	output, err := cmd.Output()
	util.Check("Error getting CPU architecture: ", err)
	return util.AsString(output)
}

func getRelease() (osID string, release string) {
	osPattern := regexp.MustCompile(`^ID=(\w+)$`)
	releasePattern := regexp.MustCompile(`^VERSION_CODENAME=(\w+)$`)
	file, err := os.Open("/etc/os-release")
	util.Check("Error reading OS info from /etc/os-release: ", err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	osID = ""
	release = ""
	for scanner.Scan() {
		line := scanner.Text()
		osMatch := osPattern.FindStringSubmatch(line)
		relMatch := releasePattern.FindStringSubmatch(line)
		if len(osMatch) > 1 {
			osID = osMatch[1]
		}
		if len(relMatch) > 1 {
			release = relMatch[1]
		}
	}
	if osID == "" {
		util.Critical("No OS was detected")
		os.Exit(1)
	}
	if release == "" {
		util.Critical("No release was detected")
		os.Exit(1)
	}

	util.Info("OS/Release: %s/%s", osID, release)
	return osID, release
}

func aptSetup() {
	arch := getArch()
	osID, release := getRelease()

	repoTmpl := "deb [arch=%s] https://download.docker.com/linux/%s %s stable"
	repo := fmt.Sprintf(repoTmpl, arch, osID, release)

	out, err := exec.Command("apt-add-repository", repo).CombinedOutput()
	util.Check("Error adding docker APT repository:\n%s", err, util.AsString(out))

	util.Success("Added %s to APT sources", repo)
}

// Install installs docker.
func Install() {
	util.Info("Installing docker...")
	aptSetup()
	out, err := exec.Command("apt-get", "update").CombinedOutput()
	util.Check("Error updating APT:\n%s", err, util.AsString(out))

	install := exec.Command("apt-get", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io")
	out, err = install.CombinedOutput()
	util.Check("Error installing Docker:\n%s", err, util.AsString(out))
}

// InstallCompose installs docker-compose.
func InstallCompose() {
	user := util.CurrentUser()
	composePath := path.Join(user.HomeDir + "/.local/bin/docker-compose")
	if user.Name == "root" {
		composePath = "/usr/local/bin/docker-compose"
	}

	if !util.FileExists(composePath) {
		util.Info("Docker compose executable not found at %s, installing Docker Compose...", composePath)
		out, err := exec.Command("pip3", "install", "docker-compose").CombinedOutput()
		util.Check("Error installing Docker Compose:\n%s", err, util.AsString(out))
	}
}

// Scaffold creates a directory in the docker config directory for docker-compose files.
func Scaffold(hostname string) {
	dir := "/etc/docker/compose"

	if !util.FileExists(dir) {
		err := os.Mkdir(dir, 0755)
		util.Check("Error creating %s directory", err, dir)
		if util.FileExists(dir) {
			util.Success("Created %s", dir)
		} else {
			util.Check("Unable to create %s directory", os.ErrNotExist, dir)
		}

	} else {
		util.Info("%s already exists", dir)
	}
	srcDir := path.Join(fmt.Sprintf(g.HomeDir, g.LocalUser), hostname)

	cmpFileSrc := path.Join(srcDir, "mp-compose.yaml")
	cmpFileDst := path.Join(dir, "appneta-cmp.yaml")

	util.CopyFile(cmpFileSrc, cmpFileDst)
	util.Success("Copied %s to %s", cmpFileSrc, cmpFileDst)

	envFileSrc := path.Join(srcDir, ".env")
	envFileDst := path.Join(dir, ".env")

	util.CopyFile(envFileSrc, envFileDst)
	util.Success("Copied %s to %s", envFileSrc, envFileDst)

}

// ReadEnv loads the AppNeta .env file & reads its lines.
func ReadEnv(hostname string) (vars []string) {
	srcDir := path.Join(fmt.Sprintf(g.HomeDir, g.LocalUser), hostname)
	filename := path.Join(srcDir, ".env")
	if !util.FileExists(filename) {
		util.Check("AppNeta .env file does not exist at %s", os.ErrNotExist, filename)
	}
	b, err := ioutil.ReadFile(filename)
	util.Check("Error reading AppNeta .env file %s", err, filename)
	parts := strings.Split(string(b), "\n")
	for _, p := range parts {
		if p != "\n" {
			vars = append(vars, p)
		}
	}
	return
}
