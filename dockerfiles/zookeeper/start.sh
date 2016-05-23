#!/bin/sh

zookeeper-server-initialize --myid=1

echo server.1=172.17.0.1:2888:3888 >> /etc/zookeeper/conf/zoo.cfg
echo server.2=172.17.0.1:2888:3888 >> /etc/zookeeper/conf/zoo.cfg

echo -e "172.17.0.2\tdb59c2d8a061" >> /etc/hosts

/usr/lib/zookeeper/bin/zkServer.sh start

# check zookeeper run mode leader|follower|standalone
yum install -y nmap-ncat
echo stat | nc localhost 2181 | grep Mode
