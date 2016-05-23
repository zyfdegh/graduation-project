@echo off

set GOPATH=%cd%
echo GOPATH=%GOPATH%

del bin

echo "Start to go third party code from github.com ..."
echo "Downloading gojenkins"
go get github.com/bndr/gojenkins
echo "Downloading logrus ..."
go get github.com/Sirupsen/logrus
echo "Downloading go-restful ..."
go get github.com/emicklei/go-restful
echo "Downloading properties ..."
go get -v -u github.com/magiconair/properties
echo "Downloading swagger ..."
go get github.com/emicklei/go-restful/swagger
echo "Downloading mejson ..."
go get github.com/compose/mejson
echo "Downloading mgo ..."
go get gopkg.in/mgo.v2
go get gopkg.in/mgo.v2/bson
echo "Downloading go-dockerclient ..."
go get github.com/fsouza/go-dockerclient
echo "Downloading jsonq ..."
go get github.com/jmoiron/jsonq
echo "Download xmlpath"
go get launchpad.net/xmlpath

echo Copying linkernetworks's libs ...
xcopy /e /y /r /i ..\linker_common_lib\linkernetworks.com src\linkernetworks.com\

echo "Start to build linker cicd ..."
del bin
go build -a -o bin/cicd.exe ./src/linkernetworks.com/linker_cicd_platform/main.go

echo "Copying properties file to bin/ ..."
copy src\cicd.properties bin\cicd.properties
copy src\javaproj_job.xml bin\javaproj_job.xml


