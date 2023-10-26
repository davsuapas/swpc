#!/bin/bash

# Build AWS Beanstalk ZIP

build=$1
https_single=$2

command_hs="-https-single"

if [[ -z $https_single ]]; then
	https_single=$command_hs
fi

path_scripts=./scripts
path_deploy_awsbt_is=./deploy/aws/beanstalk/instance-single
path_target_release="$GOPATH/swpc/release"
path_target="$GOPATH/swpc/awsbt"
path_target_bin="$path_target/bin"
path_deploy_self_certificate="./deploy/self-certificate"
path_target_ebextensions="$path_target/.ebextensions"

if [[ $build == "-build" ]]; then
	$path_scripts/build.sh
fi

rm -r "$path_target"

mkdir -p "$path_target_bin"

cp "$path_target_release/swpc-server" "$path_target_bin/application"
cp -r "$path_target_release/public" "$path_target"

if [[ $https_single == $command_hs ]]; then
	cp -r "$path_deploy_awsbt_is/.ebextensions" "$path_target/.ebextensions"
	cp -r "$path_deploy_awsbt_is/.platform" "$path_target/.platform"

	crt=$(sed 's/^/      /' "$path_deploy_self_certificate/swpc.crt")
	key=$(sed 's/^/      /' "$path_deploy_self_certificate/swpc.key")

	awk -v r="$crt" '{gsub("@@crt",r)}1' "$path_target/.ebextensions/https-instance.config" > "$path_target/.ebextensions/temp"
	awk -v r="$key" '{gsub("@@key",r)}1' "$path_target/.ebextensions/temp" > "$path_target/.ebextensions/https-instance.config"

	rm "$path_target/.ebextensions/temp"
fi

current=$(pwd)
cd "$path_target"

if [[ $https_single == $command_hs ]]; then
	zip -r source.zip bin/application public/* .platform/* .ebextensions/*
else
	zip -r source.zip bin/application public/*
fi

cd "$current"

rm -r "$path_target_bin"
rm -r "$path_target/public"
rm -r "$path_target/.ebextensions"
rm -r "$path_target/.platform"

echo "Target deployment: '$path_target'"

