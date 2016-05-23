#!/bin/bash
CICD_HOME=$(cd `dirname $0` && pwd)
export GOPATH=${CICD_HOME}/../../../

echo "Start to go third party code from github.com ..."
echo "Downloading gojenkins"
go get -v -u github.com/bndr/gojenkins
echo "Downloading logrus ..."
go get -v -u github.com/Sirupsen/logrus
echo "Downloading go-restful ..."
go get -v -u github.com/emicklei/go-restful
echo "Downloading properties ..."
go get -v -u github.com/magiconair/properties
echo "Downloading swagger ..."
go get -v -u github.com/emicklei/go-restful/swagger
echo "Downloading mejson ..."
go get -v -u github.com/compose/mejson
echo "Downloading mgo ..."
go get -v -u gopkg.in/mgo.v2
go get -v -u gopkg.in/mgo.v2/bson
echo "Downloading go-dockerclient ..."
go get -v -u github.com/fsouza/go-dockerclient
echo "Downloading jsonq ..."
go get -v -u github.com/jmoiron/jsonq
echo "Downloading zk ..."
go get -v -u github.com/samuel/go-zookeeper/zk
echo "Download xmlpath"
go get -v -u launchpad.net/xmlpath
# echo "Downloading mqu/openldap ..."
# go get github.com/mqu/openldap

# Dependency for openldap
# yum install -y openldap-devel

echo "Start to build linker cicd ..."
output_dir=${GOPATH}/bin
artifact=${output_dir}/cicd
rm -f ${artifact}
go build -a -o ${artifact} ${CICD_HOME}/main.go

if [[ $? -ne 0 ]]; then
	#build error
	echo "build ERROR"
	exit 1
fi

echo "Copying properties file to bin/ ..."
cp ${CICD_HOME}/cicd.properties ${output_dir}/cicd.properties
cp ${CICD_HOME}/javaproj_job.xml ${output_dir}/javaproj_job.xml

echo "Compress to bin/linkerops.zip..."
ARCHIVE="linkerops.zip"
cd ${output_dir}
	rm -f ${ARCHIVE}
	zip $ARCHIVE cicd
	zip $ARCHIVE cicd.properties
	zip $ARCHIVE javaproj_job.xml
cd ..


#scp bin/controller root@ansible:/root/Linker_Ansible/linker_ansible_repo/Linker_Mesos_Cluster/roles/controller/files/


