FROM scratch

# prepare logging directory
WORKDIR /tmp

WORKDIR /
COPY bin bin
COPY config.yaml config.yaml
ENV TRICARB_CONFIG /

ENTRYPOINT ["./bin"]
