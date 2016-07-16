#! /bin/bash
wget https://github.com/google/protobuf/releases/download/v2.6.1/protobuf-2.6.1.tar.gz
tar xzf protobuf-2.6.1.tar.gz
cd protobuf-2.6.1
apt-get update -y
apt-get install -y build-essential
./configure
make
make check
make install
ldconfig
protoc --version
