#!/bin/bash

# Create self-certificate

path_target="./deploy/self-certificate"

mkdir -p $path_target

openssl genrsa -traditional -out "$path_target/swpc.key" 2048
openssl req -new -key "$path_target/swpc.key" -out "$path_target/swpc-csr.key"
openssl x509 -req -days 365 -in "$path_target/swpc-csr.key" -signkey "$path_target/swpc.key" -out "$path_target/swpc.crt"