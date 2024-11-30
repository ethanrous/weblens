<h1 align="center">Weblens</h1>
<h3 align="center">Self-Hosted file manager and photo server</h3>

<p align="center">
    <img width="300" src="images/brand/logo.png" alt="weblens logo" />
    <br/>
    <br/>    
    <a href="https://github.com/ethanrous/weblens/actions/workflows/go.yml"></a>
    <img alt="Weblens Fulltest" src="https://github.com/ethanrous/weblens/actions/workflows/go.yml/badge.svg?branch=main"/>
</p>
<br/>

# Overview

Weblens is a self-hosted file and photo management system that boasts a simple and snappy experience.

## Features lightning round
* File management, including history, backup, and restore
* Clean, productive web interface
* Users, permissions & sharing of both files and media
* Photo gallery and albums
* API (not yet stable, documentation at /docs/index.html when running)

### Experimental Features
* WebDav

<br/>

# Ready to get started?
## Installation
Weblens is distributed as a Docker image. Here is a minimal docker setup to get started:
```bash
docker run --name weblens \
-p 8080:8080 \ 
-v /files/on/host:/media/users \ 
-v /cache/on/host:/media/cache \
-e MONGODB_URI="mongodb://{{ MONGO_USER }}:{{ MONGO_PASS }}@weblens-mongo:27017"
docker.io/ethrous/weblens:latest
```
Also, Weblens uses MongoDB. This can easily be setup using another container
```bash
docker run --name weblens-mongo \
-v /db/on/host:/data/db \
-e MONGO_INITDB_ROOT_USERNAME: {{ MONGO_USER }} \
-e MONGO_INITDB_ROOT_PASSWORD: {{ MONGO_PASS }} \
mongo
```
Replace `{{ MONGO_USER }}` and `{{ MONGO_PASS }}` and host paths with values of your choosing.

⚠️ **Note** Having the containers on the same Docker network is extremely helpful. [Read how to set up a Docker network](https://docs.docker.com/reference/cli/docker/network/create/). If you wish not to do this, you will have to modify the MONGODB_URI to something routable, and export port 27017 on the mongo container.

If you prefer to use docker-compose, a sample [docker-compose.yml](scripts/docker-compose.yml) is provided in the scripts directory 

## Setup
Once you have the containers configured and running, you can begin setting up your Weblens server. 

By default, Weblens uses port 8080, so I will be using `http://localhost:8080` as the example url here

A Weblens server can be configured as a ["core"](#weblens-core) server, or as a ["backup"](#weblens-backup) server. A core server is the main server that is used day to day, and an optional backup server is a one that mirrors the contents of 1 or more core servers. Ideally a backup server is physically distant from any core it backs up.

![WeblensSetup.png](images/screenshots/WeblensSetup.png)

### Weblens Core
If you are new to Weblens, you will want to set up a *core* server. Alternatively, if you already have a core server, and want to create an offsite backup, see [Weblens Backup](#weblens-backup)

Configuring a core server is very simple

![CoreSetup.png](images/screenshots/CoreSetup.png)

You will need to create a user, give the server a name, and optionally set the server address (i.e. it is behind a reverse proxy). Finally, hit "Start Weblens"

### Weblens Backup

⚠️ **Note** that a Backup server requires an existing [core server](#weblens-core), and for you to be an admin of that server

![WeblensBackupConfiguration.png](images/screenshots/WeblensBackupConfiguration.png)

1. Give your server a name. Again, it can be anything!
   - If you have hosts `host1` and `host2` and `host2` is a backup of `host1`, don't name it `host1-backup`, simply name it `host2`
   - Support for a backup server to back up multiple core servers is planned for the future
2. Add the public address where the core server can be reached
3. Generate an API key to allow access to the core server
   1. Navigate to the "files" page on your existing core server
   2. Open the admin settings menu via the button on the top right of the page
   3. Click `New Api Key` under the `API Keys` header, then click the clipboard to copy the new key
   4. Return to the weblens backup setup, paste your new API key in the "API Key" box
4. Hit "Attach To Core", then login as an existing user on the core server

In the "remotes" section of the server settings on the core, you can now view the status of your backup server

<br/>

# Want to contribute?

Weblens aims to be feature-full and rock-solid stable, but it is still early in development (the term "beta" may be pushing it), so it is likely to have bugs or missing features. Bug reports, feature suggestions, and pull requests are all welcome and encouraged here on GitHub 

## Development Setup
Weblens has a few dependencies that are needed for developing. Easy install instructions per platform are listed below

* Go 1.23
* LibVips
* MongoDB
* ExifTool
* Node and NPM for the React/Vite frontend

### MacOS
```bash
brew tap mongodb/brew &&
brew install go@1.23 mongodb-community vips mongodb-community@7.0 exiftool node npm &&
brew services start mongodb-community
```

### Linux (Ubuntu)
⚠️ On Ubuntu, installing the Go compiler and MongoDB have a few extra steps.
[Install Go compiler on Linux](https://go.dev/doc/install)
and
[Install MongoDB on Ubuntu](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/)
```bash
sudo apt update &&
sudo apt-get install -y pkg-config libvips-dev exiftool nodejs npm
```

### Building / Testing
Verify the environment is set up correctly by running tests:
```bash
./scripts/testWeblens -a -l
```
If they pass: Congrats! You are ready to start contributing!

If they don't, there is likely a configuration issue. Please re-read the instructions and ensure the environment is set up as described, and if there is still an issue, please be descriptive in asking for help on the [issues page](https://github.com/ethanrous/weblens/issues)

⚠️ **Note** that scripts must be run from the repo root, you cannot be in the scripts directory or anywhere else

### Debugging

Building and running the server can be done with the following in the shell
```bash
./scripts/startWeblens
```
This should start your Weblens server running at `localhost:8080`. To change the host, port or other config options such as log level, see `./config/config.json` and edit the `DEBUG-CORE` config section, or create your own section.

In an IDE, you will need to choose the entry point for the compiler to `./cmd/weblens/main.go`. You will also need to set the following environment variables (replace `{{ WEBLENS_REPO }}` with the absolute path to this repo, i.e. `$(pwd)`):

```
CONFIG_NAME=DEBUG-CORE
APP_ROOT={{ WEBLENS_REPO }} 
```

### WebUI
After starting your Weblens server, in a separate shell, you can run:
```bash
cd ./ui && npm start
```
This will launch the web UI at `localhost:3000`, and proxy requests to the server running at `localhost:8080`.

If the port is already in use, vite will pick the next port not in use, check the logs for which port it is using, however it will likely open in browser for you, so you don't need to worry about that. If you'd like to choose the port yourself, set `VITE_PORT`.

If you must change the Weblens server address or port, make sure to set `VITE_PROXY_HOST` and `VITE_PROXY_PORT` in the environment before running `npm start`
