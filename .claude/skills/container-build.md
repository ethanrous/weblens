---
name: container-build
description: Build, test, and publish the weblens Docker container image
---

# Container Build

## Quick commands

```bash
make docker:build        # Build full Docker image
make container           # Build + push (amd64, single-platform)
make roux                # Build and push the base image (weblens-roux)
```

## Build pipeline

The main build script is `scripts/gogogadgetdocker.bash`. The Dockerfile is at `docker/Dockerfile`.

### Base image

The base image `ethrous/weblens-roux` is built from `scripts/build-base-image.bash`. It pre-installs system dependencies to speed up CI builds. Rebuild only when system deps change.

### CI/CD

GitHub Actions workflow: `.github/workflows/container-build.yml`

## Dev vs production

- Dev: `make dev` (air hot-reload + nuxt dev)
- Production: Docker container packages compiled Go binary + pre-built frontend

## Agno (Rust image library)

The Rust `libagno.a` must be built before the Go binary. In Docker, this happens in the build stage. Locally:

```bash
make agno
```

## Troubleshooting

- If container build fails, check `scripts/gogogadgetdocker.bash` logs
- Build the base image first if deps changed: `make roux`
- The frontend must be generated before container build: `make gen-ui`
