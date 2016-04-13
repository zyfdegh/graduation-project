FROM graduation-project/jre8
MAINTAINER RedMothball zyfdegg@gmail.com

RUN rpm -Uvh 'http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm'

RUN yum clean all && yum makecache

RUN yum install -y marathon

CMD ["marathon"]
