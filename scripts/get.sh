#!/usr/bin/env bash

# Usage: curl https://raw.githubusercontent.com/NoUseFreak/docker-multi-cache/master/scripts/get.sh | bash

PROJECT=docker-multi-cache

get_latest_release() {
  curl --silent "https://api.github.com/repos/NoUseFreak/$1/releases/latest" |
	grep '"tag_name":' |
	sed -E 's/.*"([^"]+)".*/\1/'
}

download() {
  curl -Ls -o /tmp/$PROJECT.tar.gz https://github.com/NoUseFreak/$PROJECT/releases/download/$1/`uname`_amd64.tar.gz
}

extract() {
  rm -rf /usr/local/bin/$PROJECT
  tar -xf /tmp/$PROJECT.tar.gz -C /usr/local/bin/
}

echo "Looking up latest release"
RELEASE=$(get_latest_release $PROJECT)

echo "Downloading package"
$(download $RELEASE)

echo "Extract package"
$(extract)

echo "Making executable"
chmod +x /usr/local/bin/$PROJECT

echo "Installed $PROJECT in /usr/local/bin/$PROJECT"
