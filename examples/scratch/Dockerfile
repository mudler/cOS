ARG LUET_VERSION=0.16.7

FROM quay.io/luet/base:$LUET_VERSION AS luet

FROM opensuse/leap:15.3 AS ca

FROM scratch

# Copy luet from the official images
COPY --from=luet /usr/bin/luet /usr/bin/luet
COPY --from=ca /etc/ssl/certs/. /etc/ssl/certs/

# Copy the luet config file pointing to the cOS repository
ADD conf/luet.yaml /etc/luet/luet.yaml
ENV USER=root
ENV LUET_NOLOCK=true
SHELL ["/usr/bin/luet", "install", "-y", "-d"]

RUN system/cos-container

SHELL ["/bin/sh", "-c"]
RUN rm -rf /var/cache/luet/packages/ /var/cache/luet/repos/

ENV TMPDIR=/tmp
ENTRYPOINT ["/bin/sh"]