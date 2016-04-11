#!/bin/bash

# This script can create and boot rancheros(s) in virtualbox
# Usage:
# To create a vm named rancheros-01 with 512M RAM
# $./create-machine-params.sh rancheros-01 512


## create vm powered by virtualbox
##
# iso: path to iso
# mem: memory size(in MB)
# name: name for docker machine
##
# return 0 if success
function create_vm(){
	local iso=$1
	local mem=$2
	local name=$3
	
	echo "start to create machine..."
	echo "iso: $iso"
	echo "mem: $mem MB"
	echo "name: $name"
	
	docker-machine create -d virtualbox --virtualbox-boot2docker-url=$iso --virtualbox-memory=$mem $name
}

## copy public ssh key to machine by docker-machine scp
# The key a newly generated ssh key and will be copied to all
# rancheros VMs.
function copy_key(){
	local pubkey_path=$1
	local vm_name=$2
	local saving_path=$3
	docker-machine scp $pubkey_path $vm_name:$saving_path
}

## append text key file
# append key2 content to key1 tail
function append_key(){
	local vm_name=$1
	local key1=$2
	local key2=$3
	echo "append key: $key2 to $key1"
	docker-machine ssh $vm_name "cat $key2 >> $key1"
}

## delete file
function delete(){
	local vm_name=$1
	local file=$2
	docker-machine ssh $vm_name "rm $file"
}

echo "machine name: $1"
name=$1

# CONFIGURATIONS AND SETUPS
# config iso path and memory size
iso=$PWD/rancheros.iso
mem=$2
# config ssh key path and saving path
pubkey_path=$PWD/.ssh/id_rsa.pub
saving_path=/home/docker/.ssh/authorized_keys2
authorized_keys=/home/docker/.ssh/authorized_keys

printf "\n"
printf "call docker machine to create vm \n"
# CREATE AND BOOT MACHINES
create_vm $iso $mem $name

printf "\n extra configurations begin...\n"

# COPY MY SSH KEY
printf "copy my united public ssh key file...\n"
copy_key $pubkey_path $name $saving_path

# APPEND KEY
printf "append key..."
append_key $name $authorized_keys $saving_path

# DELETE KEY
printf "delete key2..."
delete $name $saving_path

printf "\n memory status:\n"
free -h

# list all docker machines
docker-machine ls

