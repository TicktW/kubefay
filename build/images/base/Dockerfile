ARG OVS_VERSION
FROM ubuntu:20.04 as cni-binaries

ARG CNI_BINARIES_VERSION
ARG WHEREABOUTS_VERSION=v0.5.1

RUN apt-get update && \
    apt-get install -y --no-install-recommends wget ca-certificates

# Leading dot is required for the tar command below
ENV CNI_PLUGINS="./host-local ./loopback ./portmap ./bandwidth"

# Download containernetworking plugin binaries for the correct architecture
RUN set -eux; \
    mkdir -p /opt/cni/bin

ADD plugins/bin /opt/cni/bin/

FROM kubefay/openvswitch:${OVS_VERSION}

LABEL maintainer="kubefay <projectkubefay-dev@googlegroups.com>"
LABEL description="An Ubuntu based Docker base image for kubefay."

USER root

# See https://github.com/kubernetes-sigs/iptables-wrappers
# /iptables-wrapper-installer.sh will have permissions of 600.
# --chmod=700 doesn't work with older versions of Docker and requires DOCKER_BUILDKIT=1, so we use
# chmod in the RUN command below instead.
ADD iptables-wrapper-installer.sh /iptables-wrapper-installer.sh

RUN apt-get update && apt-get install -y --no-install-recommends ipset jq inetutils-ping && \
    rm -rf /var/lib/apt/lists/* && \
    chmod +x /iptables-wrapper-installer.sh && \
    /iptables-wrapper-installer.sh

COPY --from=cni-binaries /opt/cni/bin /opt/cni/bin
