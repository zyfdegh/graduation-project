FROM centos:7

#RUN apk update && \
#        apk add bash && \
#        rm -rf /var/cache/apk/*

ENV CONTROLLER_HOST /usr/local/bin

EXPOSE 8081

COPY ./scripts/controller_entrypoint.sh ${CONTROLLER_HOST}/entrypoint.sh

# add depends files
COPY ./bin/controller.properties ${CONTROLLER_HOST}/controller.properties
COPY ./bin/swagger-ui ${CONTROLLER_HOST}/swagger-ui
COPY ./bin/controller ${CONTROLLER_HOST}/controller

RUN chmod +x ${CONTROLLER_HOST}/entrypoint.sh && \
        chmod +x ${CONTROLLER_HOST}/controller


ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]