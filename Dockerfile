FROM ubuntu:21.10

COPY slaxy  /usr/local/bin/

RUN mkdir /etc/slaxy

# Add Tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

WORKDIR /usr/local/bin/

ENV TZ=Asia/Shanghai \
SLAXY_TOKEN="" \
SLAXY_ADDR=":8080" \
SLAXY_GRACE_PERIOD="6s" \
SLAXY_EXCLUDED_FIELDS=""

ENTRYPOINT ["/tini", "--"]
# Run your program under Tini
CMD ["/usr/local/bin/slaxy"]
