# Want to contribute?

Weblens aims to eventually be feature-full and rock-solid stable, but it is still early in development (the term "beta" may be pushing it), so it is likely to have bugs or missing features. Bug reports, feature suggestions, and pull requests are all welcome and encouraged here on GitHub 

## Development Setup
Weblens should only require docker to get up and running with a dev environment. Inside the container, all dependencies are installed, and the code is mounted in a volume, so you can edit the code on your host machine and see the changes immediately, both frontend and backend.

You can run the dev environment with:
```bash
make dev # or `make dev-s` for https with self-signed certs
```

On the first run, this will build the docker image, install all dependencies, build the server and client, and then start the container, this can take a couple minutes. The container will run the api server on port 8080, and a HMR server for the frontend on port 3000. You can access the web server at `https://local.weblens.io:3000`.

You first may want to set a dns entry for `local.weblens.io` to point to your localhost, so you can access the server in your browser. This can be done by adding to your `/etc/hosts` file:

```bash
sudo echo '127.0.0.1       local.weblens.io' >> /etc/hosts
```

### Building / Testing
Once you have made changes, and are ready to build and run the test suite, you can run:
```bash
make test
```
This will run all the tests for the repo, frontend and backend. If they pass: Congrats! You are ready to start contributing!

If any of what is described above is not working as expected, please give it another read through, and if there is still an issue, please leave a note on the [issues page](https://github.com/ethanrous/weblens/issues)

⚠️ **Note** that scripts must be run from the repo root, you cannot be in the scripts directory or anywhere else

### Debugging
TDB
