package docker

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	g "github.com/stellaraf/rmon-node-setup/globals"
	util "github.com/stellaraf/rmon-node-setup/util"
)

const orgID = "17992"
const urlTmpl = "https://app-14.pm.appneta.com/api/v3/appliance/configuration/%s/DOCKER_COMPOSE/%s"

// AppNetaError is the JSON response received from AppNeta if there is an error with the request.
type AppNetaError struct {
	HTTPStatusCode int      `json:"httpStatusCode"`
	Messages       []string `json:"messages"`
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
			util.RunAs(g.LocalUser, func() {
				if err = os.MkdirAll(path, info.Mode()); err != nil {
					util.Check("Error unpacking AppNeta Docker image: ", err)
				}
			})()
			continue
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		util.Check(fmt.Sprintf("Error creating path %s: ", path), err)
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		util.Check("Error writing unpacked file: ", err)
	}

	if _, err := os.Stat(outTarget); os.IsNotExist(err) {
		util.Critical("Unable to download AppNeta Docker Image")
		os.Exit(1)
	}

	util.Success("Downloaded AppNeta Docker image to %s", outTarget)
}

// SetupCompose runs AppNeta's docker-compose setup script.
func SetupCompose(hostname, outDir string) {
	dir := filepath.Join(outDir, hostname)
	cmd := exec.Command("bash", "setup.sh", "-n", "bridge")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	util.Check("Error running AppNeta's docker-compose setup script:\n%v", err, util.AsString(out))
	util.Success("Completed AppNeta docker-compose setup")
}
