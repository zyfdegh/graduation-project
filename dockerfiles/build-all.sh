#!/bin/sh

DOCKERFILES=`pwd`

echo -e "\n================\n"

cd $DOCKERFILES/centos7/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/jre8/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/gerrit/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/jenkins/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/haproxy/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/marathon/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/mesos-master/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/mesos-slave/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/nginx/
./build.sh

echo -e "\n================\n"

cd $DOCKERFILES/zookeeper/
./build.sh
