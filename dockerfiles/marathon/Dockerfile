FROM graduation-project/jre8
MAINTAINER RedMothball zyfdegg@gmail.com

RUN rpm -Uvh 'http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm'

RUN yum makecache && yum install -y marathon

CMD ["marathon","--no-logger","--event_subscriber http_callback"]
