FROM centos:7
MAINTAINER RedMothball zyfdegg@gmail.com

RUN mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.backup
COPY CentOS7-Base-QCloud.repo /etc/yum.repos.d/

RUN yum makecache && yum install -y net-tools

RUN ln -f -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
