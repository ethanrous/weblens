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

## Overview

---

Weblens is a self-hosted File Management System that boasts a simple and snappy experience.

### Features lightning round
* Clean, productive web GUI
* Users & sharing
* Photo gallery and albums
* File history, backup, and restore
* API (not yet stable, documentation coming soon)

<br/>

## Ready to get started?

---

```bash
docker run -p 8080:8080 \ 
-v /files/on/host:/media/users \ 
-v /cache/on/host:/media/cache \
-e MONGODB_URI="mongodb://user:pass@mongo:27017"
docker.io/ethrous/weblens:latest
```
### or

Check out the `example.compose.yaml` for docker compose

<br/>

## Want to contribute?

---
Weblens is very early in development, and is likely to have bugs or missing features. Pull requests are encouraged

<br/>

### Experimental Features
* WebDav
