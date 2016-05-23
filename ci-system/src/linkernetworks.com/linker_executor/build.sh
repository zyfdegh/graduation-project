#!/bin/bash
EXECUTOR_HOME=$(cd `dirname $0` && pwd)
export GOPATH=${EXECUTOR_HOME}/../../../
OUTPUT_DIR=${GOPATH}/bin
ARTIFACT=${OUTPUT_DIR}/linkerexecutor

echo "Start to go third party code from github.com ..."
echo "Downloading logrus ..."
go get -v -u github.com/Sirupsen/logrus

echo "Downloading properties ..."
go get -v -u github.com/magiconair/properties

echo "Downloading go-dockercleint ..."
go get -v -u github.com/fsouza/go-dockerclient

echo "Downloading jsonq ..."
go get -v -u github.com/jmoiron/jsonq

echo "Downloading zk ..."
go get -v -u github.com/samuel/go-zookeeper/zk

echo "Downloading mgo ..."
go get -v -u gopkg.in/mgo.v2

echo "Downloading gophercloud ..."
go get -v -u github.com/rackspace/gophercloud

echo "Downloading mesos-go ..."
go get -v -u github.com/mesos/mesos-go

echo "Downloading godep ..."
go get -v -u github.com/tools/godep

echo "Downloading proto ..."
go get -v -u github.com/gogo/protobuf/proto

echo "Downloading glog ..."
go get -v -u github.com/golang/glog

echo "Downloading uuid ..."
go get -v -u github.com/pborman/uuid

echo "Go install godep"
go install github.com/tools/godep

echo "Generate godeps.json"
cp -r ${GOPATH}/src/github.com/mesos/mesos-go/examples/Godeps/ ${GOPATH}

echo "Start to download mesos-go's deps ..."
${OUTPUT_DIR}/godep restore

echo "Start to build linker executor ..."

rm -f ${ARTIFACT}
go build -a -o ${ARTIFACT} ${EXECUTOR_HOME}/main.go

if [[ $? -ne 0 ]]; then
	#build error
	echo "build ERROR"
	exit 1
fi

echo "Copying properties file to bin/ ..."
cp ${EXECUTOR_HOME}/startExecutor.sh ${OUTPUT_DIR}/startExecutor.sh
cp ${EXECUTOR_HOME}/executor.properties ${OUTPUT_DIR}/executor.properties

#scp bin/executor root@ansible:/root/Linker_Ansible/linker_ansible_repo/Linker_Mesos_Cluster/roles/mesos/files/
