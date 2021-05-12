ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="elonzh <elonzh@qq.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY clash_exporter /bin/clash_exporter

EXPOSE      9877
USER        nobody
ENTRYPOINT  [ "/bin/clash_exporter" ]
