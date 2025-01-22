# Want to contribute?

Weblens aims to be feature-full and rock-solid stable, but it is still early in development (the term "beta" may be pushing it), so it is likely to have bugs or missing features. Bug reports, feature suggestions, and pull requests are all welcome and encouraged here on GitHub 

## Development Setup
Weblens has a few dependencies that are needed for developing. Easy install instructions per platform are listed below

* Go 1.23
* LibVips
* ImageMagick
* MongoDB
* ExifTool
* Node and NPM for the React/Vite frontend

### MacOS
```bash
brew tap mongodb/brew &&
brew install go@1.23 mongodb-community vips imagemagick mongodb-community@7.0 exiftool node npm &&
brew services start mongodb-community
```

### Linux (Ubuntu)
⚠️ On Ubuntu, installing the Go compiler ImageMagick, and MongoDB have a few extra steps.
[Install Go compiler on Linux](https://go.dev/doc/install)
and
[Install MongoDB on Ubuntu](https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/)
```bash
sudo apt update &&
sudo apt-get install -y pkg-config libvips-dev exiftool nodejs npm
```
Building ImageMagick from source is recommended on Linux, as the version in the package manager is often outdated. [Instructions here](https://imagemagick.org/script/install-source.php). ImageMagick version 7 is required, which is correctly pulled with apk on Alpine Linux (like inside our docker containers), but version 6 is still pulled from apt on Ubuntu.

### Building / Testing
Verify the environment is set up correctly by running tests (run with --help for all options):
```bash
./scripts/testWeblens -a -l
```
If they pass: Congrats! You are ready to start contributing!

If they don't, there is likely a configuration issue. Please re-read the instructions and ensure the environment is set up as described. If there is still an issue, please be descriptive in asking for help on the [issues page](https://github.com/ethanrous/weblens/issues)

⚠️ **Note** that scripts must be run from the repo root, you cannot be in the scripts directory or anywhere else

### Debugging

Building and running the server can be done with the following in the shell
```bash
./scripts/startWeblens
```
This should start your Weblens server running at `localhost:8080`. To change the host, port or other config options such as log level, see `./config/config.json` and edit the `DEBUG-CORE` config section, or create your own section.

In an IDE, you will need to choose the entry point for the compiler to `./cmd/weblens/main.go`. You will also need to set the following environment variables (`{{ WEBLENS_REPO }}` is the absolute path to this repo, i.e. `$(pwd)`):

```
CONFIG_NAME=DEBUG-CORE
APP_ROOT={{ WEBLENS_REPO }} 
```

### WebUI
After starting your Weblens server, in a separate shell, you can run:
```bash
cd ./ui && npm install && npm start
```
This will launch the web UI at `localhost:3000` (if it is not in use), and proxy requests to the server running at `localhost:8080`.

If you'd like to choose the port yourself, set `VITE_PORT`. If you change the backend server address or port, make sure to set `VITE_PROXY_HOST` and `VITE_PROXY_PORT` in the environment before running `npm start`
