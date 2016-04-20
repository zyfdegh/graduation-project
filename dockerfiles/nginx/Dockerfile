FROM graduation-project/centos7
MAINTAINER RedMothball zyfdegg@gmail.com

COPY nginx.repo /etc/yum.repos.d/nginx.repo

RUN yum makecache && yum install -y nginx

COPY nginx.conf /etc/nginx/nginx.conf

ENTRYPOINT ["/usr/sbin/nginx"]
