#!/bin/bash

# Build docker from bin

command_docker=$1
no_build=$2

path_scripts=./scripts
path_target=$GOPATH/swpc

if [[ $no_build != "--no-release" ]]; then
	$path_scripts/build.sh
fi


echo "Target Deployment: '$path_target'"

cp ./deploy/Dockerfile $path_target/
cp ./deploy/docker-compose.yml $path_target/
cp ./deploy/swpc.env $path_target/

current=$(pwd)

cd $path_target

docker compose $command_docker

cd $current
