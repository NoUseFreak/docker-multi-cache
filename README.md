# Docker Multi Cache

[![Build Status](https://travis-ci.org/NoUseFreak/docker-multi-cache.svg?branch=master)](https://travis-ci.org/NoUseFreak/docker-multi-cache)

This project aim to make life easier when working with multi-stage docker builds.

## Usage

```
docker-multi-cache docker build -t repo:version .
docker-multi-cache docker push repo:version
```

## How it works

 - It searches for a Dockerfile and scrapes all build stages.
 - It tries to pull each pre-cached build if any.
 - These stages are then build using the original build command and tags each stage with a cache prefix.
 - When pushing it will also push all cached stages for later use.

## Install

### Official release

Download the latest [release](https://github.com/NoUseFreak/docker-multi-cache/releases).

```bash
curl -sL http://bit.ly/get-docker-multi-cache | bash
```

### Build from source

```sh
$ git clone https://github.com/NoUseFreak/docker-multi-cache.git
$ cd docker-multi-cache
$ make
$ make install
```

### Upgrade

To upgrade to the latest repeat the install step