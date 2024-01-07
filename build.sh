#!/bin/bash
set -e
while getopts "d:a:v:" opt
do
  case $opt in
    d) dir="$OPTARG";;
    a) app="$OPTARG";;
    v) version="$OPTARG";;
    *) ;;
  esac
done
if [ -z "$dir" ];then
  echo "-d: The compile directory cannot be empty"
  exit;
fi
if [ -z "$app" ];then
  echo "-a: The app name cannot be empty"
  exit;
fi
if [ -z "$version" ];then
  echo "-v: The version cannot be empty"
  exit;
fi


PLATFORM="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"
cd ./$dir


for platform in $PLATFORM
do
  os=${platform%/*}
  arch=${platform#*/}
  ext=""
  if [ "$os" = "windows" ]; then
    ext=".exe"
  fi
  TARGET="$app-$version-$os-$arch"
  mkdir -p ./$TARGET/config
  cp -f ../config/gomq.toml ./$TARGET/config
  CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -ldflags "-s -w" -o ./$TARGET/$app$ext ..
  chmod +x ./$TARGET/$app$ext
  tar zcf ./$TARGET.tar.gz ./$TARGET
  rm -rf ./$TARGET
done
