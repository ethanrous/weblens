<h1 align="center">Weblens</h1>
<h3 align="center">Self-Hosted file manager and photo server</h3>

<p align="center">
    <img style="float: center;" src="images/brand/logo.png" alt="weblens logo"  width=200 />
    <br/>
    <br/>    
    <a href="https://github.com/ethanrous/weblens/actions/workflows/go.yml"/>
    <img alt="Weblens Fulltest" src="https://github.com/ethanrous/weblens/actions/workflows/go.yml/badge.svg?branch=main"/>
</p>
<br/>

# Overview

Weblens is a self-hosted File Management System that boasts a simple and snappy experience.

### Features lightning round
* Clean, productive web GUI
* Users & sharing
* Photo gallery and albums
* File history, backup, and restore
* API (not yet stable, documentation coming soon)

<br/>

# Ready to get started?

Weblens is distributed as a Docker container, which can be configured minimally as such:
```bash
docker run --name weblens \
-p 8080:8080 \ 
-v /files/on/host:/media/users \ 
-v /cache/on/host:/media/cache \
-e MONGODB_URI="mongodb://{{ MONGO_USER }}:{{ MONGO_PASS }}@weblens-mongo:27017"
docker.io/ethrous/weblens:latest
```
Weblens uses MongoDB. This can easily be done using another container
```bash
docker run --name weblens-mongo \
-v /db/on/host:/data/db \
-e MONGO_INITDB_ROOT_USERNAME: {{ MONGO_USER }} \
-e MONGO_INITDB_ROOT_PASSWORD: {{ MONGO_PASS }} \
mongo
```
Replace `{{ MONGO_USER }}` and `{{ MONGO_PASS }}` with a username and password, for the database, of your choosing.
Also replace all of the `/(files | cache | db)/on/host` with paths to where you want the your stored. If you have fast and slow storage,
such as an Unraid pool and array respectively, the "files" should go on the array, and the cache and db should be on the fast cache.

Having the containers on the same docker network is extremely helpful, as it allows Docker to do DNS for you,
and does not require you to open the port on your Mongo container. [Read how to set up a docker network](https://docs.docker.com/reference/cli/docker/network/create/).

If the containers cannot be on the same newtwork, you may need to add `-p 27017:27017` to the Mongo container, 
and change the "weblens-mongo" in the `MONGODB_URI` to a route the container will understand.

<br/>

# Want to contribute?

Weblens is very early in development, and is likely to have bugs or missing features. 
Feature suggestions and pull requests are encouraged here on GitHub

## Development Setup
Weblens has a few dependencies that are needed during runtime,
and a few more just for building. The Go compiler and MongoDB (on Ubuntu) are 
to be installed manually via the links provided. For the rest, easy install instructions
per platform are below.

* Go 1.23 or later - https://go.dev/doc/install
* LibVips
* MongoDB - [Ubuntu Only](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/) see below for MacOS
* ExifTool
* Node, NPM, Vite for frontend

### MacOS
```bash
brew install vips mongodb-community@7.0 exiftool node npm
```

### Linux (Ubuntu)
```bash
sudo apt update &&
sudo apt-get install -y pkg-config libvips-dev exiftool nodejs npm
```

### Building / Testing
Once you have successfully installed the dependencies for your platform, the easiest way
to ensure your environment is correctly set up is by running 
```bash
./scripts/testWeblens
```
This will build the frontend and backend, and run the backend tests. If you are pulling from the main branch, 
these tests should pass. If this is the case: Congrats! You are ready to start writing! 

If they don't, there is likely a configuration issue. Please re-read the instructions and ensure the environemnt is set up as described, 
and if there is still an issue, please leave a descriptive question on the [issues page](https://github.com/ethanrous/weblens/issues).

### Debugging

Building the server for debugging can be done with the following in the shell.
```bash
./scripts/startWeblens
# On MacOS using Apple's new ld-prime linker you may get a 
# `ld: warning` about malformed symbols, this can safely be ignored.
```
This should start your Weblens server running at `localhost:8080`. To change the host, port or other config options 
such as log level, see `./config/config.json` and edit the `DEBUG-CORE` config section, or create your own section.

In an IDE, you must build from the main file `./cmd/weblens/main.go`, 
and set the following environment variables:
(make sure to replace `{{ WEBLENS_REPO }}` with the full path to this repo)

```
CONFIG_NAME=DEBUG-CORE
CONFIG_PATH={{ WEBLENS_REPO }}/config
```

### WebUI
After starting your Weblens server, in a separate shell, you can run:
```bash
cd ./ui && npm start
```
This will launch the web UI at `localhost:3000`, and proxy requests to the server running at `localhost:8080`.

If the port is already in use, vite will pick the next port not in use, check the logs for which port it is using,
however it will likely open in browser for you, so you don't need to worry about that. If you'd like to choose the port yourself, set `VITE_PORT`.

If you must change the Weblens server address or port, make sure to set `VITE_PROXY_HOST` and `VITE_PROXY_PORT` in the environment before running `npm start`

<br/>

### Experimental Features
* WebDav
