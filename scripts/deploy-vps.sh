#!/bin/bash

# Build for Virtual Private Server

build=$1
path_env=$

path_scripts=./scripts
path_target_release="$GOPATH/swpc/release"
path_target_swpc="$GOPATH/swpc"
path_target="$path_target_swpc/vps"
path_target_bin="$path_target/bin"

if [[ $build == "-build" ]]; then
	$path_scripts/build.sh
fi

rm -r "$path_target"
mkdir "$path_target"
mkdir "$path_target_bin"

cp $path_scripts/vps/* "$path_target"

cp "$path_env" "$path_target/swpc.env"

cp "$path_target_release/swpc-server" "$path_target_bin"
cp -r "$path_target_release/public" "$path_target_bin"
cp -r "$path_target_release/ai" "$path_target_bin"

chmod +x $path_target/*

cd $path_target_swpc
rm swpc.zip
zip -r swpc.zip vps/*
cd -

echo "Target deployment: '$path_target'"

