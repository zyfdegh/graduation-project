FROM  mesoscloud/mesos-slave:0.24.1-centos-7

RUN 	yum erase -y docker-engine && \
	yum erase -y docker-engine-selinux && \
	yum install -y docker-engine-1.8.2-1.el7.centos.x86_64

COPY ./bin/executor.properties /usr/local/bin/executor.properties
COPY ./bin/startExecutor.sh /usr/local/bin/startExecutor.sh
COPY ./bin/linkerexecutor /usr/local/bin/linkerexecutor

RUN chmod +x /usr/local/bin/startExecutor.sh