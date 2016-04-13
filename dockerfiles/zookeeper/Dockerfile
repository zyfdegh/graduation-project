FROM graduation-project/jre8
MAINTAINER RedMothball zyfdegg@gmail.com

RUN rpm -Uvh http://archive.cloudera.com/cdh4/one-click-install/redhat/6/x86_64/cloudera-cdh-4-0.x86_64.rpm

RUN yum clean all && yum makecache

RUN yum -y install zookeeper zookeeper-server

EXPOSE 2181 2888 3888

VOLUME ["/var/lib/zookeeper"]
