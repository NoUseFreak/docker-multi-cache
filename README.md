# Docker Multi Cache

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