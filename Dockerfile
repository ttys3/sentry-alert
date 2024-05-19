FROM ubuntu:24.04

COPY slaxy  /usr/local/bin/

RUN mkdir /etc/slaxy; \
    apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates; \
    update-ca-certificates -f; \
    apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false; \
    apt autoremove -y; \
    rm -rf /var/lib/apt/lists/*

# Add Tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

WORKDIR /usr/local/bin/

ENV TZ=Asia/Shanghai \
SLAXY_LOG_LEVEL=info \
SLAXY_LOG_FORMAT=json \
SLAXY_TOKEN="" \
SLAXY_ADDR=":3000" \
SLAXY_GRACE_PERIOD="6s" \
SLAXY_EXCLUDED_FIELDS=""

ENTRYPOINT ["/tini", "--"]
# Run your program under Tini
CMD ["/usr/local/bin/slaxy"]
