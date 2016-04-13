FROM graduation-project/centos7
MAINTAINER RedMothball zyfdegg@gmail.com

COPY nginx.repo /etc/yum.repos.d/nginx.repo

RUN yum clean all && yum makecache

RUN yum install -y nginx

COPY nginx.conf /etc/nginx/nginx.conf

ENTRYPOINT ["/usr/sbin/nginx"]
