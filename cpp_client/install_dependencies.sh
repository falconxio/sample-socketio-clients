#!/bin/sh

# Update package list and install dependencies
sudo apt-get update
sudo apt-get install -y git build-essential cmake libssl-dev libboost-all-dev

# Clone socket.io-client-cpp repository
git clone https://github.com/socketio/socket.io-client-cpp.git

# Checkout to 2.x branch, Build and install socket.io-client-cpp
cd socket.io-client-cpp
git checkout 2.x-tls
git submodule update --init --recursive
mkdir build
cd build
cmake ..
make
sudo make install
