#!/bin/sh

docker run --rm -it -p 5050:5050 graduation-project/mesos-master bash

echo zk://172.17.0.1:2181/mesos > /etc/mesos/zk
echo 2 > /etc/mesos-master/quorum
echo 172.17.0.2 > /etc/mesos-master/ip
echo ae93c0f9ed3d > /etc/mesos-master/hostname
nohup mesos-master --work_dir=/var/lib/mesos/master > /var/log/mesos-master.log 2>&1 &
