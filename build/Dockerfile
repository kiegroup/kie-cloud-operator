FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/kie-cloud-operator /usr/local/bin/kie-cloud-operator
