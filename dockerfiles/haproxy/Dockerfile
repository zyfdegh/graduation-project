FROM centos:7
MAINTAINER RedMothball zyfdegg@gmail.com

RUN mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.backup
COPY CentOS7-Base-163.repo /etc/yum.repos.d/

RUN yum clean all && yum makecache

RUN yum install -y haproxy

EXPOSE 8080 3868 9090 
