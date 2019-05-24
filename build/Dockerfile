FROM registry.access.redhat.com/ubi8-minimal

# install operator binary
COPY build/_output/bin/kie-cloud-operator /usr/local/bin/kie-cloud-operator
COPY build/_output/bin/console-cr-form /usr/local/bin/console-cr-form

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER 1001
