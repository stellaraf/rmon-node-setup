package docker

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	g "github.com/stellaraf/rmon-node-setup/globals"
	"github.com/stellaraf/rmon-node-setup/systemd"
	util "github.com/stellaraf/rmon-node-setup/util"
	gjson "github.com/tidwall/gjson"
)

const orgID = "17992"
const urlTmpl = "https://app-14.pm.appneta.com/api/v3/appliance/configuration/%s/DOCKER_COMPOSE/%s"

// AppNetaError is the JSON response received from AppNeta if there is an error with the request.
type AppNetaError struct {
	HTTPStatusCode int      `json:"httpStatusCode"`
	Messages       []string `json:"messages"`
}

/*
appNetaEnv is a Golang representation of the AppNeta .env file, e.g.:
	APPNETA_SERVER_ADDRESS=app-14.pm.appneta.com
	APPNETA_SERVER_KEY=9U5AG-Y71V-W-P
	APPNETA_SERVER_PORTS=80,8080
	APPNETA_CONTAINER_UUID=0C5D62FF-3EB3-46B1-A2D2-0707BE8A2820
	APPNETA_CONTAINER_NAME=test
*/
type appNetaEnv struct {
	ServerAddress string
	ServerKey     string
	ServerPorts   []string
	ContainerUUID string
	ContainerName string
}

// GetCompose downloads the docker compose image from the AppNeta portal.
func GetCompose(apiKey string, hostname string, outDir string) {
	util.Info("Downloading AppNeta Docker image...")

	url := fmt.Sprintf(urlTmpl, orgID, hostname)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("accept", "application/gzip")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", apiKey))

	res, err := client.Do(req)
	util.Check("Error getting AppNeta Docker image: ", err)
	defer res.Body.Close()

	if res.StatusCode > 399 {
		appNetaErr := AppNetaError{}

		err := json.NewDecoder(res.Body).Decode(&appNetaErr)
		util.Check("Unable to read error response: ", err)

		for _, m := range appNetaErr.Messages {
			util.Critical(m)
		}
		os.Exit(1)
	}

	outTarget := filepath.Join(outDir, hostname)

	if _, err := os.Stat(outTarget); !os.IsNotExist(err) {
		err := os.RemoveAll(outTarget)
		util.Check(fmt.Sprintf("Directory %s already exists, and it was not able to be deleted: ", outTarget), err)
		util.Info("Directory %s already exists, and it was deleted.", outTarget)
	}

	tarReader := tar.NewReader(res.Body)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		util.Check("Header error: ", err)

		path := filepath.Join(outDir, header.Name)
		info := header.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				util.Check("Error unpacking AppNeta Docker image: ", err)
			}
			continue
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		util.Check(fmt.Sprintf("Error creating path %s: ", path), err)
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		util.Check("Error writing unpacked file: ", err)

		user, err := user.Lookup(g.LocalUser)
		uid, _ := strconv.Atoi(user.Uid)
		gid, _ := strconv.Atoi(user.Gid)
		err = os.Chown(path, uid, gid)
		util.Check("Error changing ownership of %s to user %s", err, path, user.Name)
	}

	if _, err := os.Stat(outTarget); os.IsNotExist(err) {
		util.Critical("Unable to download AppNeta Docker Image")
		os.Exit(1)
	}

	util.Success("Downloaded AppNeta Docker image to %s", outTarget)
}

func readEnv(filename string) (env appNetaEnv) {
	if !util.FileExists(filename) {
		util.Check("AppNeta .env file does not exist at %s", os.ErrNotExist, filename)
	}
	b, err := ioutil.ReadFile(filename)
	util.Check("Error reading AppNeta .env file %s", err, filename)
	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		if l != "\n" {
			v := strings.Split(l, "=")
			switch v[0] {
			case "APPNETA_SERVER_ADDRESS":
				env.ServerAddress = v[1]
			case "APPNETA_SERVER_KEY":
				env.ServerKey = v[1]
			case "APPNETA_SERVER_PORTS":
				ports := strings.Split(v[1], ",")
				env.ServerPorts = ports
			case "APPNETA_CONTAINER_UUID":
				env.ContainerUUID = v[1]
			case "APPNETA_CONTAINER_NAME":
				env.ContainerName = v[1]
			}
		}
	}
	return
}

func modifyCompose(filename string) {
	if !util.FileExists(filename) {
		util.Check("AppNeta docker-compose file does not exist at %s", os.ErrNotExist, filename)
	}
	in, err := ioutil.ReadFile(filename)
	util.Check("Error reading docker-compose file %s", err, filename)
	lines := strings.Split(string(in), "\n")
	cnOld := "--containername localhost"
	cnNew := "--containername talos-001"
	networkMode := "network_mode: host"
	for i, l := range lines {
		if strings.Contains(l, networkMode) {
			// See line 92 of the AppNeta bash script. This is the equivalent to:
			// `sed -i '/network_mode:\ host/d' mp-compose.yaml`
			lines[i] = ""
			util.Info("Deleted %s line from %s", networkMode, filename)
		} else if strings.Contains(l, cnOld) {
			// See line 93 of the AppNeta bash script. This is the equivalent to:
			// `sed -i 's/containername\ localhost/containername\ talos-001/g' mp-compose.yaml`
			strings.ReplaceAll(lines[i], cnOld, cnNew)
			util.Info("Replaced %s with %s in %s", cnOld, cnNew, filename)
		}
	}
	out := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filename, []byte(out), 0644)
	util.Check("Error writing to docker-compose file %s", err, filename)
}

func getPassword(filename string) (pw string) {
	if !util.FileExists(filename) {
		util.Check("AppNeta token/password file does not exist at %s", os.ErrNotExist, filename)
	}
	b, err := ioutil.ReadFile(filename)
	util.Check("Error reading the AppNeta token/password file at %s", err, filename)
	pw = strings.Trim(string(b), "\n")
	return
}

/*
getRegistry reads the setup.sh bash script downloaded from AppNeta, and extracts the $ACR_REGISTRY
variable's value.
*/
func getRegistry(filename string) (r string) {
	if !util.FileExists(filename) {
		util.Check("AppNeta bash script does not exist at %s", os.ErrNotExist, filename)
	}
	b, err := ioutil.ReadFile(filename)
	util.Check("Error reading AppNeta bash script", err)
	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		/*
			There is a line in the script that reassigns the variable if it is provided as a
			command argument, so we need to ensure we're just pulling the initial value at the top
			of the script.
		*/
		if match, _ := regexp.MatchString("^ACR_REGISTRY=", l); match {
			p := strings.Split(l, "=")
			r = strings.Trim(p[1], `"`)
			break
		}
	}
	return
}

/*
dockerLoginStatus reads the ~/.docker/config.json file, which stores active Docker Registry login
states. If we're authenticated to the AppNeta registry, return true. If not, return false.
*/
func dockerLoginStatus(r string) (s bool) {
	user, _ := user.Current()
	filename := filepath.Join(user.HomeDir, ".docker", "config.json")
	ex := util.FileExists(filename)
	if !ex {
		s = false
		return
	}
	file, err := os.Open(filename)
	util.Check("Error opening docker config file %s", err, filename)
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	util.Check("Error reading docker config file %s", err, filename)

	/* The registry name is a URL with a path, e.g. host.example.com/path. However, Docker stores
	the registry name as an FQDN, e.g. host.example.com.
	*/
	reg := strings.Split(r, "/")[0]

	/* GJSON supports some sweet path finding, but requires that keys containing real dots,
	be escaped. For example in {"person": {"first.name": "Bob"}}, we would need to get "Bob"
	with: `person.first\.name`.
	*/
	gk := "auths." + strings.ReplaceAll(reg, ".", `\.`)

	j := gjson.ParseBytes(b)
	ke := j.Get(gk)
	if ke.Exists() {
		s = true
	}

	return
}

/*
SetupCompose runs AppNeta's docker-compose setup script.

Username is set to: TOK-$APPNETA_SERVER_KEY

Command should mimic:
echo "$ACR_PASSWORD" | $DOCKER_CMD login -u "$ACR_USERNAME" --password-stdin "$ACR_REGISTRY"
*/
func SetupCompose(hostname, outDir string) {
	dir := filepath.Join(outDir, hostname)

	modifyCompose(filepath.Join(dir, "mp-compose.yaml"))
	env := readEnv(filepath.Join(dir, ".env"))
	un := "TOK-" + env.ServerKey
	pw := getPassword(filepath.Join(dir, "tok.txt"))
	reg := getRegistry(filepath.Join(dir, "setup.sh"))

	rsd := systemd.Root()
	dockerRunning := rsd.CheckService("docker")

	if !dockerRunning {
		rsd.StartService("docker")
	}
	util.Info("Logging in to AppNeta Docker registry %s as %s...", reg, un)

	/* Docker CLI throws an error if using the --password argument, and prefers that the password
	be piped in via stdin, with the --password-stdin argument.

	See: https://docs.docker.com/engine/reference/commandline/login/#provide-a-password-using-stdin
	*/
	cmd := exec.Command("docker", "login", "--username", un, "--password-stdin", reg)
	stdin, err := cmd.StdinPipe()
	util.Check("Error accessing stdin for docker login command, which is required for inputting the password", err)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Start()
	util.Check("Error running command to log in to AppNeta Docker Registry", err)

	io.WriteString(stdin, pw+"\n")
	stdin.Close()

	err = cmd.Wait()
	util.Check("Error logging in to AppNeta Docker Registry %s:\n%v", err, reg, util.AsString(out.Bytes()))

	loggedIn := dockerLoginStatus(reg)
	if !loggedIn {
		util.Warning("AppNeta Docker setup completed, but %s is not logged into AppNeta Docker Registry %s", hostname, reg)
	} else {
		util.Success("Logged into AppNeta Docker Registry %s", reg)
	}
	return
}
