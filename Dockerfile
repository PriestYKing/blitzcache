# Dockerfile
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# For dockers_v2, GoReleaser puts artifacts in $TARGETPLATFORM directory
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/blitzcache /usr/local/bin/blitzcache

EXPOSE 6380
ENTRYPOINT ["blitzcache"]
CMD ["-addr", ":6380", "-shards", "256"]
