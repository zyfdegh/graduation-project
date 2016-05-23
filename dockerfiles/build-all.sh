#!/bin/sh

DOCKERFILES=`pwd`

echo -e "\n================\n"

echo "Building centos7..."
cd $DOCKERFILES/centos7/
./build.sh
if [[ $? -ne 0 ]]; then 
	 echo "build failed"
	 exit 1
fi

echo -e "\n================\n"

echo "Building jre8..."
cd $DOCKERFILES/jre8/
./build.sh
if [[ $? -ne 0 ]]; then 
	 echo "build failed"
	 exit 1
fi

echo -e "\n================\n"

echo "Building gerrit..."
cd $DOCKERFILES/gerrit/
./build.sh
if [[ $? -ne 0 ]]; then 
	 echo "build failed"
	 exit 1
fi

echo -e "\n================\n"

echo "Building jenkins..."
cd $DOCKERFILES/jenkins/
./build.sh

echo -e "\n================\n"

echo "Building haproxy..."
cd $DOCKERFILES/haproxy/
./build.sh
if [[ $? -ne 0 ]]; then 
	 echo "build failed"
	 exit 1
fi

echo -e "\n================\n"

echo "Building marathon..."
cd $DOCKERFILES/marathon/
./build.sh
if [[ $? -ne 0 ]]; then
	 echo "build failed"
	 exit 1
fi

echo -e "\n================\n"

echo "Building mesos..."
cd $DOCKERFILES/mesos/
./build.sh
if [[ $? -ne 0 ]]; then
	echo "build failed"
	exit 1
fi

echo -e "\n================\n"

echo "Building mesos-master..."
cd $DOCKERFILES/mesos-master/
./build.sh
if [[ $? -ne 0 ]]; then 
	 echo "build failed"
	 exit 1
fi

echo -e "\n================\n"

echo "Building mesos-slave..."
cd $DOCKERFILES/mesos-slave/
./build.sh
if [[ $? -ne 0 ]]; then 
	 echo "build failed"
	 exit 1
fi

# echo -e "\n================\n"
# 
# echo "Building nginx..."
# cd $DOCKERFILES/nginx/
# ./build.sh
# if [[ $? -ne 0 ]]; then 
# 	 echo "build failed"
# 	 exit 1
# fi

echo -e "\n================\n"

echo "Building zookeeper..."
cd $DOCKERFILES/zookeeper/
./build.sh
if [[ $? -ne 0 ]]; then 
	 echo "build failed"
	 exit 1
fi
