FROM ubuntu-debootstrap:14.04
MAINTAINER Eldarion, Inc.

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates curl net-tools \
    && rm -rf /var/lib/apt/lists/*

RUN cd /tmp \
    && curl -LO https://github.com/coreos/etcd/releases/download/v2.0.10/etcd-v2.0.10-linux-amd64.tar.gz \
    && tar xzvf etcd-v2.0.10-linux-amd64.tar.gz \
    && mv etcd-v2.0.10-linux-amd64/etcdctl /usr/local/bin/ \
    && rm -rf etcd-v2.0.10-linux-amd64

ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

EXPOSE 80 443
CMD ["/app/bin/boot"]
ADD . /app
