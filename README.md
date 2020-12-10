<div align="center">
  <br/>
  <img src="https://res.cloudinary.com/stellaraf/image/upload/v1604277355/stellar-logo-gradient.svg" width=300 />
  <br/>
  <h3>Remote Monitoring Node Setup</h3>
  <br/>
  <a href="https://github.com/stellaraf/rmon-node-setup/actions?query=workflow%3Agoreleaser">
    <img alt="GitHub Workflow Status" src="https://img.shields.io/github/workflow/status/stellaraf/rmon-node-setup/goreleaser?color=9100fa&style=for-the-badge">
  </a>
  <br/>
  This repository contains source code for Stellar's remote monitoring node setup. The compiled binary installs dependencies, registers the node with the RMON reverse SSH-tunnel server, and sets up AppNeta docker-compose container(s).
</div>

## Usage

### Download the latest [release](https://github.com/stellaraf/rmon-node-setup/releases/latest)

There are multiple builds of the release, for different CPU architectures/platforms:

| Release Suffix |                Platform |
| :------------- | ----------------------: |
| `linux_amd64`  | Linux, Intel or AMD x86 |
| `linux_armv5`  |     Linux, Raspberry Pi |

Right click the one matching your situation, and copy the link. Run the following commands to download and extract:

```shell
wget <release url>
tar xvfz <release file> rmon-node-setup
```

### Run the binary

```console
$ sudo ./rmon-node-setup
Orion RMON Raspberry Pi Setup

You'll need:

  - ID number of the unit, a unique 2 digit number between 1-99.
  - FQDN of the remote SSH tunnel server.
```

You'll receive the following prompts, so have this information ready:

```
Node ID (2 digit number):
SSH Tunnel Server (FQDN):
Enter the AppNeta API Key from IT Glue:
```

You should see a number of log messages explaining what the script is doing in the background, ending with `Setup complete!`. After this is done, the node should be available on the SSH Tunnel server via port `100xx` where `xx` is the Node ID.

## Creating a New Release

This project uses [GoReleaser](https://goreleaser.com/) to manage releases. After completing code changes and committing them via Git, be sure to tag the release before pushing:

```
git tag <release>
```

Once a new tag is pushed, GoReleaser will automagically create a new build & release.
