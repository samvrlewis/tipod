#!/usr/bin/env bash

# Sets up a server NS and a client NS with a veth pair to allow direct
# communications. The idea is that this allows for testing of tipod, with the
# client running in the client namespace and the server running in the server
# namespace. The client and server can communicate directly through the veth
# link.

ip netns add server
ip netns add client
ip link add veth0 type veth peer name veth1
ip link set veth0 netns server
ip link set veth1 netns client

ip netns exec server ip addr add 10.1.1.1/24 dev veth0
ip netns exec server ip link set dev veth0 up

ip netns exec client ip addr add 10.1.1.2/24 dev veth1
ip netns exec client ip link set dev veth1 up

ip netns exec server ip link set lo up
ip netns exec client ip link set lo up