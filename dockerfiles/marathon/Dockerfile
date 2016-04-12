FROM centos:7
MAINTAINER RedMothball zyfdegg@gmail.com

RUN mv /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.backup
COPY CentOS7-Base-163.repo /etc/yum.repos.d/

RUN rpm -Uvh 'http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm'

RUN yum clean all && yum makecache

RUN yum install -y java-1.8.0-openjdk-headless
RUN yum install -y marathon

RUN ln -f -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

CMD ["marathon"]
