# syntax=docker/dockerfile:1

FROM golang:1-alpine AS builder

# For available labels, see OCI Annotations Spec docs:
# https://specs.opencontainers.org/image-spec/annotations/#pre-defined-annotation-keys
LABEL org.opencontainers.image.source="https://github.com/logdyhq/logdy-core"

WORKDIR /go/src/logdy-core/
COPY ./ /go/src/logdy-core/

RUN \
  --mount=type=cache,mode=0755,target=/root/.cache/go-build/ \
  --mount=type=cache,mode=0755,target=/go/pkg/mod/cache/ \
  go build -x -v -o /go/bin/logdy-core \
  && go install -x -v \
  && go clean -v

FROM alpine:latest

COPY --from=builder /go/bin/logdy-core /logdy

EXPOSE 8080
ENTRYPOINT ["/logdy", "--ui-ip", "0.0.0.0" ]
CMD [ "stdin" ]
