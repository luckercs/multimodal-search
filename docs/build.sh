#!/bin/bash

set -e

if command -v npm > /dev/null 2>&1
then
    echo "check npm ok"
else
    echo "npm/node not found, please install it first"
    exit 1
fi

if command -v go > /dev/null 2>&1
then
    echo "check go ok"
else
    echo "go not found, please install it first"
    exit 1
fi

if command -v docker > /dev/null 2>&1
then
    echo "check docker ok"
else
  echo "docker not found, please install it first"
  exit 1
fi

script_dir=$(dirname "$0")
cd "$script_dir"
project_dir=$(dirname "$(pwd)")

echo "build fe..."
cd $project_dir/server-fe
npm install
npm run build

echo "build be..."
rm -rf $project_dir/server-be/web/dist
cp -r $project_dir/server-fe/dist $project_dir/server-be/web/
cd $project_dir/server-be
export GOOS=linux
go mod tidy
go build -a -o multimodal_search multimodal_search.go

echo "build docker image..."
cd $project_dir/dockerfile
docker build -f multimodal_search.dockerfile -t registry.cn-beijing.aliyuncs.com/luckercs/multimodal-search:1.0 ..

echo "build successfully"
