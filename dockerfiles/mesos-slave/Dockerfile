FROM graduation-project/centos7
MAINTAINER RedMothball zyfdegg@gmail.com

RUN rpm -Uvh 'http://repos.mesosphere.io/el/7/noarch/RPMS/mesosphere-el-repo-7-1.noarch.rpm'

RUN yum clean all && yum makecache

RUN yum install -y mesos

CMD ["/usr/sbin/mesos-slave"]

ENV MESOS_WORK_DIR /tmp/mesos
ENV MESOS_CONTAINERIZERS docker,mesos

# https://mesosphere.github.io/marathon/docs/native-docker.html
ENV MESOS_EXECUTOR_REGISTRATION_TIMEOUT 5mins

VOLUME /tmp/mesos

# RUN echo 'docker,mesos' > /etc/mesos-slave/containerizers
# RUN echo '5mins' > /etc/mesos-slave/executor_registration_timeout

# CMD ["/usr/sbin/mesos-slave"]

COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
