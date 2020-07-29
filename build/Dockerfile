FROM registry.access.redhat.com/ubi8-minimal

# install operator binary
COPY build/_output/bin/kie-cloud-operator /usr/local/bin/kie-cloud-operator
COPY build/_output/bin/console-cr-form /usr/local/bin/console-cr-form

USER 1001
