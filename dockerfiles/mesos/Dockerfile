FROM graduation-project/centos7
MAINTAINER RedMothball zyfdegg@gmail.com

RUN rpm -Uvh 'http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm'

COPY daocloud-docker.repo /etc/yum.repos.d/

RUN yum clean all && yum makecache

RUN yum install -y mesos

RUN yum install -y docker-engine
